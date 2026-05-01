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
	"regexp"
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

var absolutePathPattern = regexp.MustCompile(`(?:[A-Za-z]:\\|/)[^\s:"]+`)

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
			writeRequestTimeout(w)
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
			writeSanitizedBadRequest(w, err)
			return
		}

		evalContext, err := handlers.BuildRunContext(payload.Inputs)
		if err != nil {
			writeSanitizedBadRequest(w, err)
			return
		}
		if err := r.Context().Err(); err != nil {
			writeRequestTimeout(w)
			return
		}

		opts := runner.RunnerOptions{
			Lazy: config.Lazy,
		}
		execution, err := runner.ExecuteStringWithGoContextAndOptions(r.Context(), payload.Script, evalContext, opts)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				writeRequestTimeout(w)
				return
			}
			writeSanitizedBadRequest(w, err)
			return
		}

		outputMimeType, err := handlers.ResolveOutputMimeType(payload.Output, execution.OutputMimeType)
		if err != nil {
			writeSanitizedBadRequest(w, err)
			return
		}

		execution.OutputMimeType = outputMimeType
		formatted, err := runner.FormatExecutionResult(execution)
		if err != nil {
			log.Printf("run endpoint format error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if err := r.Context().Err(); err != nil {
			writeRequestTimeout(w)
			return
		}

		w.Header().Set("Content-Type", outputMimeType)
		if _, err := io.WriteString(w, formatted); err != nil {
			log.Printf("server write error: %v", err)
		}
	}
}

func writeRequestTimeout(w http.ResponseWriter) {
	http.Error(w, "Request timeout", http.StatusRequestTimeout)
}

func writeSanitizedBadRequest(w http.ResponseWriter, err error) {
	http.Error(w, sanitizeClientErrorMessage(err), http.StatusBadRequest)
}

func sanitizeClientErrorMessage(err error) string {
	if err == nil {
		return "bad request"
	}
	message := strings.TrimSpace(err.Error())
	if message == "" {
		return "bad request"
	}

	message = strings.ReplaceAll(message, "\r", " ")
	message = strings.ReplaceAll(message, "\n", " ")
	message = strings.Join(strings.Fields(message), " ")
	message = absolutePathPattern.ReplaceAllString(message, "<path>")

	return strings.TrimSpace(message)
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
