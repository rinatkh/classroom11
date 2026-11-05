package task

import (
	"errors"
	"fmt"
	"sync"
	"testing"
)

func TestStoreLifecycle(t *testing.T) {
	store := NewStore()

	created, err := store.Create("  Изучить HTTP  ")
	if err != nil {
		t.Fatalf("Create вернул ошибку: %v", err)
	}
	if created.ID != 1 || created.Title != "Изучить HTTP" || created.Done {
		t.Fatalf("неожиданная созданная задача: %#v", created)
	}

	done := true
	updated, err := store.Patch(created.ID, Changes{Done: &done})
	if err != nil {
		t.Fatalf("Patch вернул ошибку: %v", err)
	}
	if !updated.Done {
		t.Fatalf("ожидали done=true, получили %#v", updated)
	}

	if err := store.Delete(created.ID); err != nil {
		t.Fatalf("Delete вернул ошибку: %v", err)
	}
	if _, err := store.Get(created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("после удаления ожидали ErrNotFound, получили %v", err)
	}
}

func TestStoreValidation(t *testing.T) {
	store := NewStore()
	if _, err := store.Create("   "); !errors.Is(err, ErrEmptyTitle) {
		t.Fatalf("для пустого title ожидали ErrEmptyTitle, получили %v", err)
	}

	item, err := store.Create("Задача")
	if err != nil {
		t.Fatalf("Create вернул ошибку: %v", err)
	}
	if _, err := store.Patch(item.ID, Changes{}); !errors.Is(err, ErrNoChanges) {
		t.Fatalf("для пустого PATCH ожидали ErrNoChanges, получили %v", err)
	}

	empty := "   "
	if _, err := store.Patch(item.ID, Changes{Title: &empty}); !errors.Is(err, ErrEmptyTitle) {
		t.Fatalf("для пустого нового title ожидали ErrEmptyTitle, получили %v", err)
	}
	if _, err := store.Patch(999, Changes{Title: &item.Title}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("для отсутствующего ID ожидали ErrNotFound, получили %v", err)
	}
	if err := store.Delete(999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("для удаления отсутствующего ID ожидали ErrNotFound, получили %v", err)
	}
}

func TestStoreListSorted(t *testing.T) {
	store := NewStore()
	for _, title := range []string{"первая", "вторая", "третья"} {
		if _, err := store.Create(title); err != nil {
			t.Fatalf("Create(%q): %v", title, err)
		}
	}
	if err := store.Delete(2); err != nil {
		t.Fatalf("Delete(2): %v", err)
	}

	items := store.List()
	if len(items) != 2 || items[0].ID != 1 || items[1].ID != 3 {
		t.Fatalf("List должен быть отсортирован по ID: %#v", items)
	}
}

func TestStoreConcurrentCreate(t *testing.T) {
	store := NewStore()
	const requests = 50

	var wg sync.WaitGroup
	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			if _, err := store.Create(fmt.Sprintf("задача %d", index)); err != nil {
				t.Errorf("Create: %v", err)
			}
		}(i)
	}
	wg.Wait()

	items := store.List()
	if len(items) != requests {
		t.Fatalf("ожидали %d задач, получили %d", requests, len(items))
	}
}
