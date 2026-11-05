# Live coding: что произошло до знакомого Handler?

Classroom 11 начинается с финального Classroom 10. Заранее откройте
`cmd/api/main.go`, `internal/httpapi/handler.go`, `router.go` и
`internal/task/store.go`: это тот же код с путями импортов нового Go-модуля.
Ориентир занятия — около 90 минут; echo-опыты можно показать готовыми.

## 1. Знакомый API — 8 минут

1. Остановите старый процесс занятия 10, если он занял `8080`.
2. Запустите `make run` в `classroom11`.
3. Во втором терминале выполните `curl -i http://localhost:8080/health`.
4. Найдите прежние `200`, Content-Type и `{"status":"ok"}`.
5. Откройте `Health` и напомните: «На прошлом занятии мы начинали рассуждение
   с готового HTTP request».
6. Спросите: «А что произошло ДО Handler?»

На экране оставьте Task API. Мы исследуем сеть под этим приложением.

## 2. Raw TCP/HTTP к Task API — 14 минут

```bash
make raw-http
```

Откройте `cmd/03_raw_http/main.go` и `internal/networklab/rawhttp.go`.

1. Адрес — `127.0.0.1:8080`, где уже работает знакомый API.
2. Найдите `DialContext`: это TCP-соединение, ещё не HTTP request.
3. В `BuildRawGET` прочитайте request line, Host и Connection.
4. Найдите реальные `\r\n` и завершающее `\r\n\r\n`.
5. `io.WriteString` отправляет эти байты; `net/http` на стороне API разбирает их.
6. Покажите прежний JSON `/health` и тот же `Health` в коде.

Произнесите главный переход: «В прошлом занятии мы писали curl GET /health.
Сейчас видим байты, которые curl/http.Client скрывали».

Дополнительно создайте задачу обычным POST, затем прочитайте её тем же raw
клиентом: `go run ./cmd/03_raw_http -path /tasks`. Это данные того же процесса.

Негативный опыт: временно уберите завершающую пустую строку в BuildRawGET.
Сервер ждёт конец headers, клиент ограничен двухсекундным deadline. Верните код.

## 3. TCP properties на маленьком echo — 12 минут

HTTP/1.1 и HTTP/2 обычно работают поверх TCP. Убираем HTTP из одного опыта,
чтобы увидеть транспорт; предметный проект по-прежнему Task API.

```bash
make tcp-demo
```

При ручном запуске используйте два терминала: `make tcp-server` и
`make tcp-client`. API на `8080` продолжает работать, echo слушает `8081`.

1. В `tcp.go` найдите Listen, Accept, net.Conn и goroutine для клиента.
2. Три сообщения проходят через одно соединение.
3. Найдите наше правило framing — `\n`. TCP сам не сохраняет границы Write.
4. Сопоставьте framing строк echo с завершением HTTP headers `\r\n\r\n`.

## 4. UDP contrast — 8 минут

```bash
make udp-demo
```

Откройте `udp.go` рядом с `tcp.go`: нет Accept, ReadFrom получает датаграмму
и адрес. Обсудите отсутствие гарантий доставки/порядка. Свяжите с QUIC:
HTTP/3 использует QUIC поверх UDP, где надёжность реализует QUIC. Это не
новый домен и не перенос нашего API на UDP.

## 5. httptrace на том же API — 12 минут

```bash
make http-trace
```

1. Клиент делает два GET к `/health` процесса на `8080`.
2. Найдите ConnectStart и первый GotConn с reused=false.
3. Во втором GotConn найдите reused=true и оба Response 200 OK /health.
4. Покажите io.Copy(io.Discard, response.Body) и Close: они позволяют reuse.
5. Спросите: «Меняли ли мы Health ради trace?» Ответ: нет.

Для URL с IP DNS не нужен. При `-url http://localhost:8080/health` события
DNS могут зависеть от resolver/кеша. На HTTP нет TLSHandshake; это нормально.
Не создавайте новый Transport на каждый GET.

## 6. TLS → HTTP/2 — 10 минут

```bash
make tls-http2
```

Покажите `httpapi.NewRouter(task.NewStore())` внутри локального test server.
Business code тот же; временный порт, отдельный Store и доверенный тестовый
сертификат нужны только опыту. Данные этого процесса не разделяются с `8080`.

Разведите понятия: TLS 1.3 защищает соединение, ALPN выбирает h2, HTTP/2.0 —
версия HTTP. Затем откройте прежний Health: он не знает, каким транспортом
пришёл request. Текстовый raw GET был HTTP/1.1; HTTP/2 кодирует frames иначе.

## 7. Timeouts — 10 минут

```bash
make timeouts
go run ./cmd/06_timeouts -timeout 300ms
```

1. У отдельного local test server `/health` работает через прежний router.
2. Только лаборатория добавляет `/debug/slow` с задержкой 150 мс.
3. При 40 мс deadline ждём context.DeadlineExceeded; при 300 мс — успех.
4. Найдите r.Context().Done(): отменённый запрос прекращает ожидание сервера.
5. Если запрос исходил бы из Handler, parent context берём из r.Context().

Не добавляйте `/debug/slow` в `internal/httpapi/router.go` или `cmd/api`.
Это инфраструктурный эксперимент, не часть Task API и не код следующего урока.

## 8. Возвращение к Handler — 6 минут

```bash
curl -i http://localhost:8080/health
go test ./...
```

Откройте Health, PatchTask, router и Store. Ученик должен проговорить:

«Handler не изменился. Сначала идут transport и, для HTTPS, TLS, затем
net/http разбирает HTTP, ServeMux выбирает Handler, Store и JSON работают
как в прошлом занятии».

Повторите прежний lifecycle через `make curl-demo` после перезапуска API с
пустым Store. Вопросы для закрепления: чем stream отличается от datagram,
почему нужен framing, что позволяет reuse, чем TLS отличается от HTTP/2.

В занятие 12 передаём исходный Task API с памятью. См. `continuity.md`:
сетевые лаборатории и slow endpoint в исходную точку следующего урока не входят.
