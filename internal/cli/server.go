package cli

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"infomunge/internal/handlers"
	"infomunge/internal/runner"
)

const (
	serverReadHeaderTimeout = 10 * time.Second
	serverReadTimeout       = 30 * time.Second
	serverWriteTimeout      = 30 * time.Second
	serverIdleTimeout       = 60 * time.Second
	serverShutdownTimeout   = 5 * time.Second
	runRequestBodyMaxBytes  = 1 * 1024 * 1024
)

func (app *App) serve(config *Config) error {
	addr := config.Listen
	if addr == "" {
		addr = defaultListenAddr
	}

	mux := app.serverMux(config)

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: serverReadHeaderTimeout,
		ReadTimeout:       serverReadTimeout,
		WriteTimeout:      serverWriteTimeout,
		IdleTimeout:       serverIdleTimeout,
	}

	signalCtx, stopSignals := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stopSignals()

	errCh := make(chan error, 1)
	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err, ok := <-errCh:
		if !ok {
			return nil
		}
		return err
	case <-signalCtx.Done():
		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), serverShutdownTimeout)
		defer cancelShutdown()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return err
		}
		<-errCh
		return nil
	}
}

func (app *App) ServerHandler(config *Config) http.Handler {
	return app.serverMux(config)
}

func (app *App) serverMux(config *Config) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/run", app.handleRun(config))
	mux.HandleFunc("/", app.handlePlayground())
	mux.HandleFunc("/index.html", app.handlePlayground())
	return mux
}

func (app *App) handlePlayground() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if _, err := io.WriteString(w, playgroundHTML); err != nil {
			log.Printf("server write error: %v", err)
		}
	}
}

func (app *App) handleRun(config *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if !authorizedRunRequest(r, config.ServerAPIKey) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if err := r.Context().Err(); err != nil {
			http.Error(w, fmt.Sprintf("request canceled: %v", err), http.StatusRequestTimeout)
			return
		}

		requestBody := http.MaxBytesReader(w, r.Body, runRequestBodyMaxBytes)
		payload, err := handlers.DecodeRunRequest(requestBody)
		if err != nil {
			var maxBytesErr *http.MaxBytesError
			if errors.As(err, &maxBytesErr) {
				http.Error(w, fmt.Sprintf("request body exceeds %d bytes", runRequestBodyMaxBytes), http.StatusRequestEntityTooLarge)
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		evalContext, err := handlers.BuildRunContext(payload.Inputs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := r.Context().Err(); err != nil {
			http.Error(w, fmt.Sprintf("request canceled: %v", err), http.StatusRequestTimeout)
			return
		}

		opts := runner.RunnerOptions{
			Lazy: config.Lazy,
		}
		result, _, headerOutputMimeType, evalCtx, err := runner.RunStringWithGoContextAndOptionsWithOutput(r.Context(), payload.Script, evalContext, opts)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		outputMimeType, err := handlers.ResolveOutputMimeType(payload.Output, headerOutputMimeType)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		formatted, err := handlers.FormatRunResult(result, outputMimeType, evalCtx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := r.Context().Err(); err != nil {
			http.Error(w, fmt.Sprintf("request canceled: %v", err), http.StatusRequestTimeout)
			return
		}

		w.Header().Set("Content-Type", outputMimeType)
		if _, err := io.WriteString(w, formatted); err != nil {
			log.Printf("server write error: %v", err)
		}
	}
}

func authorizedRunRequest(r *http.Request, expectedAPIKey string) bool {
	if expectedAPIKey == "" {
		return true
	}
	if subtle.ConstantTimeCompare([]byte(r.Header.Get("X-API-Key")), []byte(expectedAPIKey)) == 1 {
		return true
	}
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if len(authHeader) <= len("Bearer ") || !strings.EqualFold(authHeader[:len("Bearer ")], "Bearer ") {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(strings.TrimSpace(authHeader[len("Bearer "):])), []byte(expectedAPIKey)) == 1
}
