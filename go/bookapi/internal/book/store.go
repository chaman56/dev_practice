package book

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sort"
	"sync"
	"time"
)

// Store describes what any storage must be able to do.
// Handlers depend on this interface, not on a concrete database,
// so later you can swap MemoryStore for a PostgresStore without touching handlers.
type Store interface {
	List(ctx context.Context) ([]Book, error)
	Get(ctx context.Context, id string) (Book, error)
	Create(ctx context.Context, in Input) (Book, error)
	Update(ctx context.Context, id string, in Input) (Book, error)
	Delete(ctx context.Context, id string) error
}

// MemoryStore keeps books in a map. Data is lost on restart,
// which is fine for learning; use a real database in production.
type MemoryStore struct {
	mu    sync.RWMutex // protects books: many requests run at the same time
	books map[string]Book
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{books: make(map[string]Book)}
}

func (s *MemoryStore) List(ctx context.Context) ([]Book, error) {
	s.mu.RLock() // read lock: many readers allowed at once
	defer s.mu.RUnlock()

	out := make([]Book, 0, len(s.books))
	for _, b := range s.books {
		out = append(out, b)
	}
	// Maps have random order; sort so results are predictable.
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}

func (s *MemoryStore) Get(ctx context.Context, id string) (Book, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	b, ok := s.books[id]
	if !ok {
		return Book{}, ErrNotFound
	}
	return b, nil
}

func (s *MemoryStore) Create(ctx context.Context, in Input) (Book, error) {
	now := time.Now().UTC()
	b := Book{
		ID:        newID(),
		Title:     in.Title,
		Author:    in.Author,
		Year:      in.Year,
		CreatedAt: now,
		UpdatedAt: now,
	}

	s.mu.Lock() // write lock: only one writer, no readers
	defer s.mu.Unlock()
	s.books[b.ID] = b
	return b, nil
}

func (s *MemoryStore) Update(ctx context.Context, id string, in Input) (Book, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, ok := s.books[id]
	if !ok {
		return Book{}, ErrNotFound
	}
	b.Title, b.Author, b.Year = in.Title, in.Author, in.Year
	b.UpdatedAt = time.Now().UTC()
	s.books[id] = b
	return b, nil
}

func (s *MemoryStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.books[id]; !ok {
		return ErrNotFound
	}
	delete(s.books, id)
	return nil
}

// newID returns a random 32-character hex string.
func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
