package task

import (
	"errors"
	"sort"
	"strings"
	"sync"
)

var (
	// ErrNotFound означает, что объекта с таким ID нет.
	ErrNotFound = errors.New("задача не найдена")
	// ErrEmptyTitle означает, что заголовок состоит только из пробелов.
	ErrEmptyTitle = errors.New("заголовок задачи не должен быть пустым")
	// ErrNoChanges означает, что PATCH не содержит ни одного известного поля.
	ErrNoChanges = errors.New("не указано ни одного поля для изменения")
)

// Store хранит задачи в map, пока работает процесс сервера.
//
// HTTP-сервер может выполнять несколько handler-функций одновременно.
// Поэтому доступ к общей map защищён RWMutex — это прямое применение темы
// прошлого занятия про goroutine, гонки данных и mutex.
type Store struct {
	mu     sync.RWMutex
	tasks  map[int]Task
	nextID int
}

// NewStore создаёт пустое хранилище. Первый объект получит ID 1.
func NewStore() *Store {
	return &Store{
		tasks:  make(map[int]Task),
		nextID: 1,
	}
}

// List возвращает копию списка задач по возрастанию ID.
//
// Мы не отдаём наружу внутреннюю map: вызывающий код не должен менять
// состояние хранилища в обход методов Store.
func (s *Store) List() []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Task, 0, len(s.tasks))
	for _, item := range s.tasks {
		result = append(result, item)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result
}

// Get ищет одну задачу по ID.
func (s *Store) Get(id int) (Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.tasks[id]
	if !ok {
		return Task{}, ErrNotFound
	}
	return item, nil
}

// Create проверяет заголовок, назначает ID и сохраняет задачу.
func (s *Store) Create(title string) (Task, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Task{}, ErrEmptyTitle
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	item := Task{ID: s.nextID, Title: title, Done: false}
	s.tasks[item.ID] = item
	s.nextID++
	return item, nil
}

// Patch меняет только поля, которые клиент действительно прислал.
func (s *Store) Patch(id int, changes Changes) (Task, error) {
	if changes.Title == nil && changes.Done == nil {
		return Task{}, ErrNoChanges
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.tasks[id]
	if !ok {
		return Task{}, ErrNotFound
	}
	if changes.Title != nil {
		title := strings.TrimSpace(*changes.Title)
		if title == "" {
			return Task{}, ErrEmptyTitle
		}
		item.Title = title
	}
	if changes.Done != nil {
		item.Done = *changes.Done
	}
	s.tasks[id] = item
	return item, nil
}

// Delete удаляет задачу. Повторное удаление того же ID вернёт ErrNotFound.
func (s *Store) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tasks[id]; !ok {
		return ErrNotFound
	}
	delete(s.tasks, id)
	return nil
}

// Как не нужно:
//
//	var tasks = map[int]Task{}
//	// Два HTTP-запроса могут одновременно записать в tasks.
//
// Обычная map не предназначена для конкурентной записи. Результатом может
// стать гонка данных или panic. Проверяем подобные ошибки через go test -race.
