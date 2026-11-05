# Непрерывность Classroom 10 → 11 → 12

**Classroom 11 начинается с финального состояния Classroom 10.**
Classroom 10 — канонический исходник Task API. Classroom 11 исследует путь
request до `net/http` и Handler этого же приложения.

## Что пришло из 10 без изменения поведения

| Код | Что сохраняем |
|---|---|
| `internal/task/model.go` | Task с id/title/done и Changes для PATCH |
| `internal/task/store.go` | Store, CRUD, map, sync.RWMutex, ID с 1 |
| `internal/httpapi/handler.go` | Health, List/Get/Create/Patch/DeleteTask |
| `internal/httpapi/json.go` | JSON helpers и формат ошибок |
| `internal/httpapi/router.go` | ServeMux и те же шесть endpoints |
| `cmd/api/main.go` | запуск Task API на 8080 и штатные startup logs |
| `internal/task/*_test.go`, `internal/httpapi/handler_test.go` | исходные тесты; запросы endpoint-тестов проходят router |
| `test/integration/lifecycle_test.go` | прежний CRUD lifecycle |
| `scripts/curl-demo.sh` | прежний ручной сценарий |

В скопированных Go-файлах изменён только префикс import:
`github.com/rinatkh/classroom_backend_1/` → `github.com/rinatkh/classroom11/`.
Это отдельные каталоги Go-модулей для материалов двух занятий, а исходный
business/HTTP code совпадает. Исправление теста и удаление debug prints
сначала выполнены в каноническом Classroom 10, затем перенесены в 11.

## Что добавляет занятие 11

| Лаборатория | Связь с Task API |
|---|---|
| `03_raw_http` | вручную отправляет GET /health или /tasks в работающий cmd/api |
| `01_tcp_echo` | изолирует свойства TCP, поверх которого работают HTTP/1.1 и HTTP/2 |
| `02_udp_echo` | показывает контраст datagram/stream и готовит к обсуждению QUIC/HTTP3 |
| `04_http_trace` | два GET к знакомому API, connect, first byte, response и connection reuse |
| `05_tls_http2` | тот же router в отдельном локальном TLS listener; HTTP/2 и ALPN |
| `06_timeouts` | отдельный local test server с тем же API и лабораторным /debug/slow |

Все шесть лабораторий существовали и сохранены; нового седьмого сетевого
эксперимента нет. Добавлены рабочий API, его тесты, curl lifecycle и проверка
raw TCP client на настоящем router. TLS/timeout серверы существуют только
пока выполняется соответствующее демо, имеют свои Store и закрываются при выходе.

## Что НЕ меняется

- Business behavior: trim title, создание, чтение, частичное изменение,
  удаление задач; повторный PATCH done=true допустим как прежде.
- Хранение: только память процесса, без внешнего хранилища.
- HTTP contract: GET /health, GET /tasks, GET /tasks/{id}, POST /tasks,
  PATCH /tasks/{id}, DELETE /tasks/{id}.
- JSON модели: id, title, done; пустой список []; Location при создании;
  пустой body при 204; ошибки handlers в поле error.
- Статусы: 200/201/204 для успеха, 400/404/405 для прежних ошибок.

`/debug/slow` зарегистрирован только в `cmd/06_timeouts/main.go` на временном
сервере. В `cmd/api` такого endpoint нет. Не переносим его в следующие занятия.
Не вводим DTO layer, UseCase, Repository interface, DI abstraction,
authentication или PostgreSQL в 10 и 11.

## Порядок урока

Запустить Task API → curl /health → вспомнить готовый HTTP request прошлого
урока → спросить «Что было ДО Handler?» → raw TCP/HTTP → свойства TCP → UDP
как контраст → httptrace → TLS → HTTP/2 → timeout → вернуться к тому же Handler.

Номера существующих cmd-каталогов сохранены для ссылок из материалов;
`README.md`, `live-coding-script.md` и `make demos` задают этот новый порядок.

Финальный вывод: **Handler не изменился.** Меняется глубина наблюдения пути
request, а не предметная логика приложения.

## Исходная точка занятия 12

Для архитектурного рефакторинга берём финальный Task API занятия 10, который
сохранён и в 11: `cmd/api`, `internal/task`, `internal/httpapi`, тесты и curl
lifecycle. Сначала запускаем его с памятью и прежним контрактом. Затем уже
объясняем изменения архитектуры на знакомых обработчиках. `networklab` и
шесть сетевых cmd остаются материалом 11-го занятия.

**Обнаруженное несоответствие текущему архиву:** `classroom12/before/main.go`
уже требует PostgreSQL/pgx и реализует только POST /tasks; README занятия 12
также предполагает базу и новые статусы. Это не точная исходная точка после
Classroom 10/11. Для бесшовного продолжения нужен отдельный патч занятия 12:
начать с сохранённого Task API с Memory Store и явно объяснять дальнейшие
изменения. Этот патч меняет только Classroom 10 и 11; код занятия 12 не изменён.

## Проверка границы

1. В обоих занятиях: go test ./..., go vet ./..., go test -race ./....
2. На свежем процессе API: тот же curl lifecycle и проверки 400/404/405.
3. В 11: raw GET /health и /tasks, два traced GET с reuse, TCP/UDP echo,
   TLS с h2, timeout с 40 мс и успешный запрос с 300 мс.
4. В cmd/api GET /debug/slow возвращает 404.
5. Сопоставить task/httpapi/cmd/api/test после нормализации префикса import:
   различий нет. Homework не входит в патчи.

## Наследование логов

Единый `internal/logging` появился в classroom10. DEBUG / INFO / ERROR / FATAL,
`LOG_LEVEL` и безопасные поля сохраняются в финале и в `before/` следующих
занятий. Меняется набор наблюдаемых операций, прежний Task HTTP contract остаётся.
Подробности и показ уровней — [logging.md](logging.md).
