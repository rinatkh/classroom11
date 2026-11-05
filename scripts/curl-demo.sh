#!/usr/bin/env bash
set -euo pipefail

base_url="${BASE_URL:-http://localhost:8080}"

echo "1. Проверяем сервер"
curl -sS -i "$base_url/health"

echo
echo "2. Создаём первую задачу"
curl -sS -i -X POST "$base_url/tasks" \
  -H 'Content-Type: application/json' \
  -d '{"title":"Learn HTTP"}'

echo
echo "3. Создаём вторую задачу"
curl -sS -i -X POST "$base_url/tasks" \
  -H 'Content-Type: application/json' \
  -d '{"title":"Write handler tests"}'

echo
echo "4. Получаем список"
curl -sS -i "$base_url/tasks"

echo
echo "5. Получаем задачу с ID 1"
curl -sS -i "$base_url/tasks/1"

echo
echo "6. Отмечаем задачу выполненной"
curl -sS -i -X PATCH "$base_url/tasks/1" \
  -H 'Content-Type: application/json' \
  -d '{"done":true}'

echo
echo "7. Удаляем задачу с ID 2"
curl -sS -i -X DELETE "$base_url/tasks/2"

echo
echo "8. Проверяем итоговый список"
curl -sS -i "$base_url/tasks"

echo
echo "9. Специально отправляем сломанный JSON"
curl -sS -i -X POST "$base_url/tasks" \
  -H 'Content-Type: application/json' \
  -d '{"title":'
