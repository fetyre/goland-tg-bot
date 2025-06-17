package reminders

import (
	"sync"
	"time"
)

type Reminder struct {
	ChatID   int64     
	Text     string
	Time     time.Time
}

type Storage interface {
	Add( rem Reminder ) error            // добавить напоминание
	FetchDue( now time.Time ) []Reminder // получить все “сработавшие” (due) напоминания
	Delete( rem Reminder ) error         // удалить конкретное напоминание
	ListAll() []Reminder               // (опционально) получить все напоминания (для отладки)
}

type memoryStorage struct {
	mu        sync.Mutex
	reminders []Reminder
}

// NewMemoryStorage создаёт новый экземпляр in-memory хранилища
func NewMemoryStorage() Storage {
	return &memoryStorage{
		reminders: make([]Reminder, 0, 10),
	}
}

func (m *memoryStorage) Add(rem Reminder) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reminders = append(m.reminders, rem)
	return nil
}

func (m *memoryStorage) FetchDue(now time.Time) []Reminder {
	m.mu.Lock()
	defer m.mu.Unlock()

	var due []Reminder
	var remain []Reminder
	// Разделяем на “сработавшие” и “остающиеся”
	for _, r := range m.reminders {
		if !now.Before(r.Time) {
			due = append(due, r)
		} else {
			remain = append(remain, r)
		}
	}
	// Сохраняем только “остающиеся”
	m.reminders = remain
	return due
}

// Delete удаляет конкретное напоминание (по совпадению всех полей)
// Однако в нашей логике удаление осуществляется в FetchDue, поэтому
// метод Delete используем только, если нужно “ручное” удаление.
func (m *memoryStorage) Delete(rem Reminder) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var newSlice []Reminder
	for _, r := range m.reminders {
		// сравним точное совпадение полей
		if r.ChatID == rem.ChatID && r.Text == rem.Text && r.Time.Equal(rem.Time) {
			continue // пропускаем “удаляемое”
		}
		newSlice = append(newSlice, r)
	}
	m.reminders = newSlice
	return nil
}

// ListAll возвращает копию всех напоминаний (для отладки)
func (m *memoryStorage) ListAll() []Reminder {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Возвращаем копию, чтобы никто не менял внутренний slice извне
	copySlice := make([]Reminder, len(m.reminders))
	copy(copySlice, m.reminders)
	return copySlice
}