package book

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestServer gives each test a fresh, isolated server with an empty store.
func newTestServer() http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil)) // silence logs in tests
	mux := http.NewServeMux()
	NewHandler(NewMemoryStore(), logger).Register(mux)
	return mux
}

// do sends a fake request and returns the recorded response.
func do(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestCRUDFlow(t *testing.T) {
	h := newTestServer()

	// Create
	rec := do(t, h, "POST", "/books", `{"title":"Dune","author":"Frank Herbert","year":1965}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: want 201, got %d: %s", rec.Code, rec.Body)
	}
	var created Book
	json.NewDecoder(rec.Body).Decode(&created)
	if created.ID == "" {
		t.Fatal("create: expected an ID")
	}

	// Read
	rec = do(t, h, "GET", "/books/"+created.ID, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("get: want 200, got %d", rec.Code)
	}

	// Update
	rec = do(t, h, "PUT", "/books/"+created.ID, `{"title":"Dune Messiah","author":"Frank Herbert","year":1969}`)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Dune Messiah") {
		t.Fatalf("update: got %d: %s", rec.Code, rec.Body)
	}

	// List
	rec = do(t, h, "GET", "/books", "")
	var all []Book
	json.NewDecoder(rec.Body).Decode(&all)
	if len(all) != 1 {
		t.Fatalf("list: want 1 book, got %d", len(all))
	}

	// Delete, then make sure it's gone
	if rec = do(t, h, "DELETE", "/books/"+created.ID, ""); rec.Code != http.StatusNoContent {
		t.Fatalf("delete: want 204, got %d", rec.Code)
	}
	if rec = do(t, h, "GET", "/books/"+created.ID, ""); rec.Code != http.StatusNotFound {
		t.Fatalf("get after delete: want 404, got %d", rec.Code)
	}
}

// Table-driven tests: one list of cases, one loop. Very common in Go.
func TestCreateRejectsBadInput(t *testing.T) {
	h := newTestServer()

	cases := []struct {
		name string
		body string
		want int
	}{
		{"broken json", `{"title":`, http.StatusBadRequest},
		{"unknown field", `{"titel":"x","author":"y"}`, http.StatusBadRequest},
		{"missing title", `{"author":"y","year":2000}`, http.StatusUnprocessableEntity},
		{"future year", `{"title":"x","author":"y","year":3000}`, http.StatusUnprocessableEntity},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if rec := do(t, h, "POST", "/books", tc.body); rec.Code != tc.want {
				t.Errorf("want %d, got %d: %s", tc.want, rec.Code, rec.Body)
			}
		})
	}
}
