package cli

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"infomunge/internal/handlers"
	"infomunge/internal/runner"
)

func (app *App) serve(config *Config) error {
	addr := config.Listen
	if addr == "" {
		addr = ":8080"
	}

	mux := app.serverMux(config)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return server.ListenAndServe()
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
			fmt.Fprintf(os.Stderr, "server write error: %v\n", err)
		}
	}
}

func (app *App) handleRun(config *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		payload, err := handlers.DecodeRunRequest(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		context, err := handlers.BuildRunContext(payload.Inputs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		opts := runner.RunnerOptions{
			Lazy: config.Lazy,
		}
		result, _, headerOutputMimeType, evalCtx, err := runner.RunStringWithContextAndOptionsWithOutput(payload.Script, context, opts)
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

		w.Header().Set("Content-Type", outputMimeType)
		if _, err := io.WriteString(w, formatted); err != nil {
			fmt.Fprintf(os.Stderr, "server write error: %v\n", err)
		}
	}
}
