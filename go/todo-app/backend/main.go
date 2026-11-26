package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

// Todo represents a single todo item.
type Todo struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Store is a PostgreSQL-backed, thread-safe store of todos.
type Store struct {
	db  *sql.DB
	mu  sync.Mutex // serializes writes; reads go straight to the DB
}

// NewStore connects to PostgreSQL and ensures the schema exists.
func NewStore(dsn string) (*Store, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Reasonable connection pool settings.
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Minute * 5)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	// Ensure schema (idempotent).
	if _, err := db.Exec(createTableSQL); err != nil {
		db.Close()
		return nil, err
	}

	return &Store{db: db}, nil
}

// Close releases the database pool.
func (s *Store) Close() error {
	return s.db.Close()
}

const createTableSQL = `
CREATE TABLE IF NOT EXISTS todos (
    id          SERIAL PRIMARY KEY,
    title       TEXT NOT NULL,
    completed   BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_todos_completed ON todos(completed);
`

// List returns all todos ordered by most recently updated first.
func (s *Store) List() ([]*Todo, error) {
	rows, err := s.db.Query(`
		SELECT id, title, completed, created_at, updated_at
		FROM todos
		ORDER BY updated_at DESC, id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*Todo
	for rows.Next() {
		t := &Todo{}
		if err := rows.Scan(&t.ID, &t.Title, &t.Completed, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if result == nil {
		result = []*Todo{}
	}
	return result, nil
}

// Create inserts a new todo and returns it with generated fields.
func (s *Store) Create(title string) (*Todo, error) {
	t := &Todo{}
	err := s.db.QueryRow(`
		INSERT INTO todos (title, completed)
		VALUES ($1, FALSE)
		RETURNING id, title, completed, created_at, updated_at
	`, title).Scan(&t.ID, &t.Title, &t.Completed, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// Get returns a todo by ID, or (nil, sql.ErrNoRows) if not found.
func (s *Store) Get(id int) (*Todo, error) {
	t := &Todo{}
	err := s.db.QueryRow(`
		SELECT id, title, completed, created_at, updated_at
		FROM todos
		WHERE id = $1
	`, id).Scan(&t.ID, &t.Title, &t.Completed, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

// Update modifies a todo's title and/or completed state.
// Only non-nil fields are changed. Returns (nil, sql.ErrNoRows) if not found.
func (s *Store) Update(id int, title *string, completed *bool) (*Todo, error) {
	// Build dynamic query so we only update provided fields.
	args := []any{id}
	sets := []string{}
	argIdx := 2

	if title != nil {
		sets = append(sets, `title = $`+strconv.Itoa(argIdx))
		args = append(args, *title)
		argIdx++
	}
	if completed != nil {
		sets = append(sets, `completed = $`+strconv.Itoa(argIdx))
		args = append(args, *completed)
		argIdx++
	}

	if len(sets) == 0 {
		return s.Get(id)
	}

	sets = append(sets, `updated_at = NOW()`)
	query := `UPDATE todos SET ` + strings.Join(sets, ", ") +
		` WHERE id = $1 RETURNING id, title, completed, created_at, updated_at`

	t := &Todo{}
	err := s.db.QueryRow(query, args...).Scan(&t.ID, &t.Title, &t.Completed, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

// Delete removes a todo by ID. Returns (false, nil) if not found, (true, nil) on success.
func (s *Store) Delete(id int) (bool, error) {
	res, err := s.db.Exec(`DELETE FROM todos WHERE id = $1`, id)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

var store *Store

// enableCORS adds CORS headers so the Next.js frontend (localhost:3000) can call the API.
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// todosHandler handles GET /api/todos (list) and POST /api/todos (create).
func todosHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		todos, err := store.List()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, todos)
		return
	case http.MethodPost:
		var req struct {
			Title string `json:"title"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}
		if strings.TrimSpace(req.Title) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "title is required"})
			return
		}
		todo, err := store.Create(req.Title)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, todo)
		return
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

// todoHandler handles GET/PUT/DELETE /api/todos/{id}.
func todoHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	// Expected: ["api", "todos", "<id>"]
	if len(parts) < 3 || parts[0] != "api" || parts[1] != "todos" {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	id, err := strconv.Atoi(parts[2])
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid todo id"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		todo, err := store.Get(id)
		if err != nil {
			if err == sql.ErrNoRows {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "todo not found"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, todo)

	case http.MethodPut:
		var req struct {
			Title     *string `json:"title"`
			Completed *bool   `json:"completed"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}
		todo, err := store.Update(id, req.Title, req.Completed)
		if err != nil {
			if err == sql.ErrNoRows {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "todo not found"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, todo)

	case http.MethodDelete:
		ok, err := store.Delete(id)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "todo not found"})
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func main() {
	// Build the DSN from env with sensible defaults matching the todo-db container.
	host := getEnv("PGHOST", "localhost")
	port := getEnv("PGPORT", "5432")
	user := getEnv("PGUSER", "todo_user")
	password := getEnv("PGPASSWORD", "todo_password")
	dbname := getEnv("PGDATABASE", "todo_db")
	sslmode := getEnv("PGSSLMODE", "disable")

	dsn := "host=" + host + " port=" + port + " user=" + user +
		" password=" + password + " dbname=" + dbname + " sslmode=" + sslmode

	var err error
	store, err = NewStore(dsn)
	if err != nil {
		log.Fatalf("connecting to postgres: %v", err)
	}
	defer store.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/todos", todosHandler)
	mux.HandleFunc("/api/todos/", todoHandler)

	addr := getEnv("LISTEN_ADDR", ":8080")
	log.Printf("Todo API listening on http://localhost%s (DB: %s:%s/%s)", addr, host, port, dbname)
	if err := http.ListenAndServe(addr, enableCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}
