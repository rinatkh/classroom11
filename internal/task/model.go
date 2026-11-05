// Package task содержит модель задачи и хранилище в памяти.
package task

// Task — одна задача, которую Task API возвращает клиенту.
//
// JSON-теги задают имена полей во внешнем API. Имена полей Go остаются
// с заглавной буквы: только экспортируемые поля видит пакет encoding/json.
type Task struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// Changes описывает поля, которые можно изменить через PATCH.
//
// Указатели нужны, чтобы отличить два случая:
//   - поле не прислали: nil, значение оставляем прежним;
//   - поле прислали со значением false или пустой строкой: указатель не nil.
type Changes struct {
	Title *string
	Done  *bool
}
