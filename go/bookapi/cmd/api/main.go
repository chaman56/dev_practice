// main is the entry point: it wires all the pieces together and starts the server.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/bookapi/internal/book"
	"github.com/example/bookapi/internal/config"
)

func main() {
	// Structured JSON logs are easy for log tools (Datadog, Loki, CloudWatch) to search.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := run(logger); err != nil {
		logger.Error("server stopped with error", "err", err)
		os.Exit(1)
	}
}

// run holds the real logic so main stays tiny and errors are handled in one place.
func run(logger *slog.Logger) error {
	cfg := config.Load()

	// 1. Build dependencies.
	store := book.NewMemoryStore()
	handler := book.NewHandler(store, logger)

	// 2. Build the router.
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok")) // load balancers ping this to see if we're alive
	})
	handler.Register(mux)

	// 3. Wrap the router with middleware (code that runs around every request).
	var h http.Handler = mux
	h = logRequests(logger, h)
	h = recoverPanics(logger, h)

	// 4. Configure the server with timeouts, so slow or stuck clients
	//    can't hold connections open forever.
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// 5. Start the server in the background.
	serverErr := make(chan error, 1)
	go func() {
		logger.Info("server starting", "port", cfg.Port)
		serverErr <- srv.ListenAndServe()
	}()

	// 6. Wait for Ctrl+C (SIGINT) or a stop signal from Docker/Kubernetes (SIGTERM).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErr:
		return err // server failed to start, e.g. port already in use
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	}

	// 7. Graceful shutdown: stop accepting new requests,
	//    let in-flight ones finish, but give up after the timeout.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	if err := <-serverErr; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	logger.Info("server stopped cleanly")
	return nil
}

// ---- middleware ----

// statusRecorder remembers the status code so we can log it.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// logRequests logs one line per request: method, path, status and duration.
func logRequests(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

// recoverPanics stops one buggy request from crashing the whole server.
func recoverPanics(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("panic recovered", "err", err, "path", r.URL.Path)
				http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
