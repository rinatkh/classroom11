# Как не нужно: сетевые антипримеры

Эти фрагменты не запускаются из рабочего кода. Они нужны, чтобы ученица могла назвать ошибку и последствие.

## 1. Один `Read` якобы равен одному сообщению

```go
buffer := make([]byte, 1024)
n, _ := conn.Read(buffer)
fmt.Println(string(buffer[:n]))
```

TCP может вернуть часть сообщения или сразу несколько сообщений. Сначала выберите framing: строка с `\n`, длина перед body или готовый протокол.

## 2. Сервер обрабатывает только одного TCP-клиента

```go
conn, _ := listener.Accept()
handle(conn)
```

После первого клиента сервер больше не вызывает `Accept`. Обычно нужен цикл, а клиентов можно обрабатывать конкурентно.

## 3. Ошибка Accept игнорируется

```go
conn, _ := listener.Accept()
go handle(conn)
```

При ошибке `conn` может быть `nil`. Ошибку нужно проверить и решить, продолжать цикл или завершаться.

## 4. UDP принимается за «быстрый TCP»

```go
_, _ = udp.Write(largePayload)
// Значит данные точно дошли и в правильном порядке.
```

UDP сам не подтверждает доставку, не восстанавливает порядок и не повторяет потерянные датаграммы.

## 5. HTTP/1.1 без `Host`

```http
GET /health HTTP/1.1


```

В HTTP/1.1 `Host` обязателен. Один IP может обслуживать несколько имён, и серверу нужно выбрать нужный виртуальный host.

## 6. Нет пустой строки после headers

```http
GET /health HTTP/1.1
Host: localhost

```

Сервер ещё не увидел `\r\n\r\n` и может продолжать ждать headers. Внешне это выглядит как зависание.

## 7. Новый `Transport` на каждый запрос

```go
func load(url string) error {
    client := &http.Client{Transport: &http.Transport{}}
    _, err := client.Get(url)
    return err
}
```

Каждый Transport получает отдельный pool соединений. Переиспользуйте client/transport на протяжении жизни приложения.

## 8. Body не закрывается

```go
response, _ := client.Get(url)
data, _ := io.ReadAll(response.Body)
_ = data
```

Нужен `defer response.Body.Close()`. Чтобы HTTP/1.x-соединение обычно вернулось в pool, body также нужно дочитать до EOF.

## 9. Нет ограничения ожидания

```go
response, err := http.Get(url)
```

`http.DefaultClient` не задаёт общий timeout. Для backend используйте настроенный client и/или context с deadline.

## 10. Проверка TLS отключается

```go
tlsConfig := &tls.Config{InsecureSkipVerify: true}
```

Так клиент перестаёт нормально проверять сертификат и имя сервера. Это не исправление сертификата. В учебной локальной среде лучше использовать доверенный тестовый client, как в `cmd/05_tls_http2`.

## 11. `CloseIdleConnections` после каждого запроса

```go
response, _ := client.Get(url)
client.CloseIdleConnections()
```

Так мы уничтожаем пользу от pool. Метод нужен для явного освобождения простаивающих соединений, а не как обязательное завершение каждого request.

## 12. TCP keepalive называют HTTP keep-alive

TCP keepalive — механизм ОС для проверки соединения после простоя. HTTP persistence — повторное использование соединения для нескольких requests. Они решают разные задачи.
