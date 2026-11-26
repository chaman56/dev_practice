// Package book holds everything about the "book" resource:
// the data model, validation, storage and HTTP handlers.
package book

import (
	"errors"
	"strings"
	"time"
)

// Book is the thing our API creates, reads, updates and deletes.
// The `json:"..."` tags control the field names in JSON responses.
type Book struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	Year      int       `json:"year"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Input is what a client is allowed to send us.
// We keep it separate from Book so clients can't set ID or timestamps themselves.
type Input struct {
	Title  string `json:"title"`
	Author string `json:"author"`
	Year   int    `json:"year"`
}

// Validate checks the input and returns a list of problems (field -> message).
// An empty map means the input is valid.
func (in *Input) Validate() map[string]string {
	problems := map[string]string{}

	in.Title = strings.TrimSpace(in.Title)
	in.Author = strings.TrimSpace(in.Author)

	if in.Title == "" {
		problems["title"] = "is required"
	} else if len(in.Title) > 200 {
		problems["title"] = "must be at most 200 characters"
	}
	if in.Author == "" {
		problems["author"] = "is required"
	}
	if in.Year < 0 || in.Year > time.Now().Year() {
		problems["year"] = "must be between 0 and the current year"
	}
	return problems
}

// ErrNotFound is returned by the store when a book doesn't exist.
// Using a named error lets callers check for it with errors.Is.
var ErrNotFound = errors.New("book not found")
