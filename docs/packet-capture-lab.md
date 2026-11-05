# Дополнительная лаборатория: увидеть пакеты

Это необязательный блок. Основной урок не зависит от Wireshark или `tcpdump`.

## Цель

Связать вызовы Go с наблюдаемыми сетевыми событиями:

- TCP SYN / SYN-ACK / ACK;
- HTTP/1.1 request и response на loopback;
- FIN или RST при завершении;
- UDP datagrams без TCP handshake.

## Вариант с Wireshark

1. Выберите loopback-интерфейс (`lo`, `lo0` или адаптер loopback в Windows).
2. Установите display filter `tcp.port == 8081`.
3. Запустите `make tcp-server`, затем `make tcp-client`.
4. Найдите handshake и сегменты с данными.
5. Повторите с фильтром `udp.port == 8082`, `make udp-server` и `make udp-client`.

Для raw HTTP сначала запустите знакомый Task API через `make run`, выберите
фильтр `tcp.port == 8080`, затем выполните `make raw-http`. Найдите те же
`GET /health` и JSON response, которые видели через curl на занятии 10.
Автоматические `make tcp-demo` и `make udp-demo` используют свободные порты;
для фиксированных фильтров выше запускайте server/client вручную.

## Что спросить

- Почему два вызова `Write` не обязаны стать двумя TCP packets?
- Почему UDP capture не содержит SYN/SYN-ACK/ACK?
- Почему HTTPS body нельзя прочитать как plain text в capture?

## Безопасность

Не захватывайте трафик чужой сети и не публикуйте captures с cookies, tokens или персональными данными. Для урока достаточно loopback и локальных программ.
