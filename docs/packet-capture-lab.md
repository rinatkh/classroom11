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
5. Повторите с фильтром `udp.port == 8082` и UDP demo.

Для raw HTTP используйте `tcp.port == 8080` только если запускаете HTTP server на этом порту. Локальный автономный `make raw-http` выбирает случайный свободный port; его номер выводится в request header `Host`.

## Что спросить

- Почему два вызова `Write` не обязаны стать двумя TCP packets?
- Почему UDP capture не содержит SYN/SYN-ACK/ACK?
- Почему HTTPS body нельзя прочитать как plain text в capture?

## Безопасность

Не захватывайте трафик чужой сети и не публикуйте captures с cookies, tokens или персональными данными. Для урока достаточно loopback и локальных программ.
