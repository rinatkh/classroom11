# Classroom 11 — сеть под знакомым Task API

**Classroom 11 начинается с финального состояния Classroom 10.** Сначала
запускаем тот же Task API, затем отвечаем на вопрос: «Что происходило ниже
`net/http` до того, как request попал в наш знакомый Handler?»

В прошлом занятии мы писали `curl GET /health` и рассматривали готовый HTTP
request. Сейчас видим байты и этапы соединения, которые скрывали curl и
`http.Client`. Task, Store, handlers, router и HTTP-контракт остаются прежними.

Подробная передача между занятиями: [`docs/continuity.md`](docs/continuity.md).

## Приложение, которое уже знаем

```text
Client → HTTP → ServeMux → Handler → Memory Store → JSON response
```

| Endpoint | Назначение | Успешный status |
|---|---|---:|
| `GET /health` | проверка сервера | `200` |
| `GET /tasks` | список задач | `200` |
| `GET /tasks/{id}` | одна задача | `200` |
| `POST /tasks` | создание | `201` + `Location` |
| `PATCH /tasks/{id}` | изменение присланных полей | `200` |
| `DELETE /tasks/{id}` | удаление, пустой body | `204` |

JSON задачи: `{"id":1,"title":"Learn HTTP","done":false}`. Пустой список —
`[]`. Ошибки прежние: `400` для некорректного ID/JSON/входа, `404` для
отсутствующей задачи, `405` для неподходящего method. Ошибки handlers имеют
JSON-формат `{"error":"..."}`; ответы самого ServeMux, например `405`, остаются
стандартными. Store — `map[int]Task` под `sync.RWMutex`; перезапуск очищает данные.

## Структура

```text
classroom11/
├── cmd/
│   ├── api/                         Task API из занятия 10
│   ├── 03_raw_http/                 ручной GET к запущенному Task API
│   ├── 01_tcp_echo/{server,client}   лаборатория свойств TCP
│   ├── 02_udp_echo/{server,client}   сравнение с UDP
│   ├── 04_http_trace/               два GET к тому же Task API
│   ├── 05_tls_http2/                тот же router в локальном TLS listener
│   └── 06_timeouts/                 отдельный локальный slow-эксперимент
├── internal/
│   ├── task/                       исходные Task, Changes, Store и тесты
│   ├── httpapi/                    исходные handlers, JSON helpers и router
│   └── networklab/                 функции сетевых экспериментов и тесты
├── test/integration/               прежний Task API lifecycle
├── docs/                          continuity, live coding, шпаргалки
├── scripts/                       curl lifecycle, TCP/UDP demo
├── Makefile
└── go.mod
```

Имена и номера шести лабораторий сохранены для ссылок в материалах. Порядок
занятия начинается с Task API и raw HTTP, поэтому номер каталога не задаёт
порядок live coding.

## Старт: сначала знакомый API

Нужен Go 1.22 или новее. Если сервер Classroom 10 ещё работает на `8080`,
остановите его: здесь запускаем его перенесённую версию из `cmd/api`.

```bash
make ci
make run
```

Во втором терминале, из каталога `classroom11`:

```bash
curl -i http://localhost:8080/health
```

Ожидаем `200 OK`, `Content-Type: application/json; charset=utf-8` и
`{"status":"ok"}`. Откройте `internal/httpapi/handler.go`: это знакомый `Health`.

Напомните: «На прошлом занятии мы начинали рассуждение с готового HTTP
request». Сегодня вопрос: «А что произошло ДО Handler?»

Для повторения прежнего CRUD есть `make curl-demo`. Сценарий использует ID
1 и 2, поэтому запускайте его один раз после старта с пустым Store. Для
повторения перезапустите API. Он не нужен для запуска сетевых лабораторий.

## Live coding: от API к сети и обратно

### 1. Raw HTTP к этому же Task API

```bash
make raw-http
# Дополнительно: прочитать задачи из того же Store
# go run ./cmd/03_raw_http -addr 127.0.0.1:8080 -path /tasks
```

Клиент подключается к работающему `cmd/api`, а не создаёт подменный endpoint.
Запрос выглядит так:

```http
GET /health HTTP/1.1
Host: 127.0.0.1:8080
Connection: close


```

В коде строки завершаются `\r\n`, а `\r\n\r\n` заканчивает headers.
Проследите путь: `net.Dial → TCP → байты HTTP → net/http → ServeMux → Health`.
Ответ — тот же JSON, что был у curl. `http.ReadResponse` только разбирает ответ;
запрос вручную пишет `io.WriteString`. `Connection: close` выбран для простоты
этого опыта; reuse увидим в `httptrace`.

### 2. TCP: почему HTTP нужен framing

Task API работает поверх HTTP. HTTP/1.1 и HTTP/2 обычно используют TCP.
Отдельный минимальный echo помогает увидеть свойства транспорта без HTTP.

```bash
make tcp-demo
# Или в двух терминалах: make tcp-server / make tcp-client
```

Откройте `internal/networklab/tcp.go`: `Listen`, `Accept`, `net.Conn`, чтение и
запись потока. Границы сообщений задаёт наш `\n`, а не TCP. Это лаборатория
транспорта, а предметным проектом остаётся Task API.

### 3. UDP как контраст

```bash
make udp-demo
# Или в двух терминалах: make udp-server / make udp-client
```

Сравните `ReadFrom` и `Accept`: UDP сохраняет границы датаграмм, но не обещает
доставку и порядок. Это подготовка к QUIC/HTTP/3: QUIC строит свой надёжный
транспорт поверх UDP. Мы не переводим Task API на UDP и не реализуем QUIC.

### 4. Наблюдаем http.Client на знакомом /health

```bash
make http-trace
# Другой знакомый маршрут:
# go run ./cmd/04_http_trace -url http://127.0.0.1:8080/tasks
```

Два последовательных GET используют один client/Transport. Ожидаем
`GotConn reused=false`, затем `GotConn reused=true`, и два
`Response 200 OK /health`. Body читается до конца и закрывается.

Handler не менялся — изменился способ наблюдать request. В адресе с IP DNS
не нужен; для обсуждения разрешения имени попробуйте
`-url http://localhost:8080/health`. События DNS зависят от resolver и кеша.
HTTP на `8080` не использует TLS; handshake рассмотрим следующим шагом.

### 5. TLS, затем HTTP/2

```bash
make tls-http2
```

Демо создаёт **отдельный локальный экземпляр того же Task API**:
`httpapi.NewRouter(task.NewStore())`. У него свой пустой Store и временный
порт; это не данные процесса на `8080`. Знакомый `GET /health` проходит через
TLS и согласование HTTP/2, затем попадает в прежний Handler.

Разделите три факта output: `HTTP/2.0`, `TLS 1.3`, `ALPN: h2`. Ручной текстовый
запрос прошлого опыта относится к HTTP/1.1; HTTP/2 использует бинарные frames.
Локальному сертификату доверяет только созданный тестовый client; проверка
сертификата не отключается. После опыта test server закрывается.

### 6. Timeout как инфраструктурный эксперимент

```bash
make timeouts
# Дайте клиенту больше времени: теперь ожидаем успех
# go run ./cmd/06_timeouts -timeout 300ms
```

Отдельный local test server сначала отвечает через тот же `GET /health`.
Только в `cmd/06_timeouts` к нему добавлен `GET /debug/slow`: задержка 150 мс,
клиентский deadline 40 мс. Ожидаем `errors.Is(context.DeadlineExceeded) == true`.
При `-timeout 300ms` запрос успевает, ошибка равна `nil`. Отмена прекращает и
учебную задержку сервера через `r.Context().Done()`.

`/debug/slow` — инфраструктурная лаборатория, не business feature. Его нет
в `internal/httpapi/router.go`, в `cmd/api` и в исходной точке занятия 12.

### 7. Возвращаемся к исходному Handler

Снова выполните `curl /health` на `8080`, откройте `Health`, `PatchTask`,
router и Store. Весь CRUD работает по прежнему договору.

**Handler не изменился.** Мы прошли глубже по пути request до HTTP layer:
`TCP → [TLS для HTTPS] → HTTP → ServeMux → Handler`. Для HTTP/3 транспортом
был бы QUIC поверх UDP; это объяснение, а не новая реализация Task API.

## Команды проверки

```bash
gofmt -w $(find . -name '*.go' -not -path './bin/*')
go test ./...
go vet ./...
go test -race ./...
```

| Команда | Назначение |
|---|---|
| `make run` | запуск перенесённого Task API |
| `make curl-demo` | прежний lifecycle на пустом Store |
| `make demos` | raw HTTP → TCP → UDP → trace → TLS/HTTP2 → timeout; API уже запущен |
| `make fmt-check` | проверка форматирования |
| `make coverage-check` | общий порог покрытия `internal/...` 85% |
| `make build` | сборка API и всех клиентов/серверов лабораторий |
| `make ci` | fmt, vet, tests, race, coverage и build |

`make ci` и тесты сами используют временные локальные порты; запущенный API
им не нужен. Автоматические echo-demo запускают собственные бинарники на
свободных портах и завершают их при выходе.

## Что дальше

В занятие 12 передаём Task API с хранением в памяти из занятия 10, тестами и
прежним HTTP-контрактом. Сетевые лаборатории остаются материалом занятия 11.
Здесь не вводим DTO layer, UseCase, Repository interface, DI abstraction,
authentication или PostgreSQL. Замечание о текущем `classroom12/before` в
архиве приведено в [`docs/continuity.md`](docs/continuity.md).

Дополнительные материалы сохранены: [шпаргалка](docs/network-cheatsheet.md),
[карта протоколов](docs/protocol-map.md), [антипримеры](docs/bad-examples.md),
[вопросы для собеседования](docs/interview.md),
[packet capture](docs/packet-capture-lab.md).

## Логи

[Уровни DEBUG / INFO / ERROR / FATAL, LOG_LEVEL и безопасные поля](docs/logging.md).

`LOG_LEVEL=DEBUG make run` включает диагностику; обычный запуск использует INFO.
