# Book API — a beginner-friendly, production-style CRUD app in Go

CRUD = **C**reate, **R**ead, **U**pdate, **D**elete. This API manages a list of books.
It uses only Go's standard library (no frameworks), so everything you see is plain Go.

## Run it

```bash
go run ./cmd/api          # starts on http://localhost:8080
go test -race ./...       # runs the tests
```

Try it from another terminal:

```bash
# Create
curl -X POST localhost:8080/books -d '{"title":"Dune","author":"Frank Herbert","year":1965}'
# List all
curl localhost:8080/books
# Get one (paste the id you got back)
curl localhost:8080/books/<id>
# Update
curl -X PUT localhost:8080/books/<id> -d '{"title":"Dune Messiah","author":"Frank Herbert","year":1969}'
# Delete
curl -X DELETE localhost:8080/books/<id>
```

| Method | Path          | What it does      | Success status |
|--------|---------------|-------------------|----------------|
| GET    | /books        | list all books    | 200 OK         |
| POST   | /books        | create a book     | 201 Created    |
| GET    | /books/{id}   | get one book      | 200 OK         |
| PUT    | /books/{id}   | replace a book    | 200 OK         |
| DELETE | /books/{id}   | delete a book     | 204 No Content |
| GET    | /healthz      | "am I alive?"     | 200 OK         |

## Folder layout — why it's split up this way

```
bookapi/
├── cmd/api/main.go          ← the program's starting point; wires everything together
├── internal/
│   ├── book/
│   │   ├── book.go          ← the data shape (Book) + input validation
│   │   ├── store.go         ← where books are saved (an interface + in-memory version)
│   │   ├── handler.go       ← turns HTTP requests into store calls
│   │   └── handler_test.go  ← automated tests
│   └── config/config.go     ← reads settings from environment variables
├── Dockerfile               ← packages the app into a container
├── Makefile                 ← short commands: make run, make test
└── go.mod                   ← the module name and Go version
```

- **`cmd/`** holds programs you can run. If you later add a CLI tool, it goes in `cmd/othertool/`.
- **`internal/`** is special in Go: code in here can't be imported by other projects. It keeps your private code private.
- Each file has **one job**. When something breaks, you know where to look.

## The layers (how a request flows)

```
HTTP request → middleware (log, recover) → handler → store → back out as JSON
```

1. **Model (`book.go`)** — `Book` is what we store. `Input` is what a client may send.
   They're separate on purpose: a client should never be able to set the `id` or `created_at`.
2. **Store (`store.go`)** — `Store` is an *interface*: a list of methods any storage must have.
   `MemoryStore` is one implementation using a map. Because handlers only know about the
   interface, you could write `PostgresStore` later and change one line in `main.go`.
3. **Handler (`handler.go`)** — reads the request, calls the store, writes a response.
   It knows HTTP; the store doesn't. That separation is called *separation of concerns*.
4. **main.go** — creates the store, the handler, the router and the server, then starts it.
   Passing things in like this (instead of using global variables) is *dependency injection*.
   It's what makes testing easy: tests just pass in a fresh store.

## The "production" practices, one by one

| Practice | Where | Why it matters |
|---|---|---|
| **Input validation** | `Input.Validate` | Never trust client data. Bad input gets a clear `422` with which fields are wrong. |
| **Reject unknown fields** | `DisallowUnknownFields` | A typo like `"titel"` becomes an error instead of silently doing nothing. |
| **Body size limit** | `MaxBytesReader` (1 MB) | Stops someone sending a 5 GB body to crash your server. |
| **Correct status codes** | handlers | 201 for created, 204 for deleted, 404 for missing, 400 for bad JSON, 422 for invalid data, 500 for our bugs. Clients rely on these. |
| **Consistent JSON errors** | `writeError` | Every error looks like `{"error": "..."}`, so clients handle them one way. |
| **Hide internal errors** | `serverError` | The real error goes to logs; the client only sees "internal server error". Leaking internals helps attackers. |
| **Named errors** | `ErrNotFound` + `errors.Is` | Lets the handler tell "not found" (404) apart from "database is down" (500). |
| **Thread safety** | `sync.RWMutex` | Go serves many requests at the same time. Without the lock, two writes to the map at once would crash the program. |
| **Context** | `ctx` in every store method | Carries cancellation: if the client disconnects, a real DB query can stop early. Unused by the map store but ready for a database. |
| **Structured logging** | `log/slog` JSON | Logs like `{"status":201,"path":"/books"}` can be searched and graphed by log tools. |
| **Request logging middleware** | `logRequests` | One log line per request with status and duration — your first debugging tool in production. |
| **Panic recovery** | `recoverPanics` | A bug in one request returns a 500 instead of killing the server for everyone. |
| **Server timeouts** | `http.Server{...Timeout}` | Go's default server has *no* timeouts; slow clients could hold connections forever. |
| **Graceful shutdown** | `signal.NotifyContext` + `srv.Shutdown` | On deploy, Kubernetes/Docker sends SIGTERM. We finish in-flight requests before exiting instead of cutting users off. |
| **Config from env vars** | `config.go` | Same binary everywhere; only `PORT` etc. change. (This is one of the "12-factor app" rules.) |
| **Health check** | `/healthz` | Load balancers and Kubernetes ping it to know whether to send traffic. |
| **Tests** | `handler_test.go` | `httptest` fakes requests without a real network. *Table-driven tests* list many cases in one slice. `-race` detects unsafe concurrent access. |
| **Small, secure container** | `Dockerfile` | Multi-stage build: compile in a big image, ship only the binary in a tiny one that runs as a non-root user. |

## What you'd add next for a real product

1. **A real database** — write `PostgresStore` (using `pgx` or `database/sql`) implementing the same `Store` interface.
2. **Pagination** on `GET /books` (`?limit=20&offset=40`) so lists don't grow unbounded.
3. **Authentication** (API keys or JWT) as another middleware.
4. **Request IDs** in logs so you can follow one request across services.
5. **Metrics** (Prometheus) and **CI** that runs `make lint test` on every pull request.
