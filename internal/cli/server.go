package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	unifiederrors "infomunge/internal/errors"
	"infomunge/internal/runner"
	"infomunge/pkg/formats"
)

type runInput struct {
	Name    string `json:"name"`
	Format  string `json:"format,omitempty"`
	Content string `json:"content"`
}

type runRequest struct {
	Script string     `json:"script"`
	Output string     `json:"output"`
	Inputs []runInput `json:"inputs,omitempty"`
}

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

		payload, err := decodeRunRequest(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		context, err := buildRunContext(payload.Inputs)
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

		outputMimeType, err := resolveOutputMimeType(payload.Output, headerOutputMimeType)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		formatted, err := formatRunResult(result, outputMimeType, evalCtx)
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

func resolveOutputMimeType(requested string, headerOutput string) (string, error) {
	if strings.TrimSpace(requested) != "" {
		return normalizeMimeType(requested, "output")
	}
	return normalizeMimeType(headerOutput, "output")
}

func decodeRunRequest(body io.Reader) (*runRequest, error) {
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()

	var payload runRequest
	if err := decoder.Decode(&payload); err != nil {
		return nil, unifiederrors.ValidationErrorf("invalid JSON body: %v", err)
	}
	if err := ensureEOF(decoder); err != nil {
		return nil, err
	}

	payload.Script = strings.TrimSpace(payload.Script)
	if payload.Script == "" {
		return nil, unifiederrors.ValidationError("script is required")
	}

	payload.Output = strings.TrimSpace(payload.Output)

	return &payload, nil
}

func ensureEOF(decoder *json.Decoder) error {
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return unifiederrors.ValidationError("invalid JSON body: unexpected trailing data")
	}
	return nil
}

func normalizeMimeType(format string, label string) (string, error) {
	format = strings.TrimSpace(format)
	if format == "" {
		return "text/plain", nil
	}
	if strings.Contains(format, "/") {
		return format, nil
	}
	mimeType := formats.MimeTypeForFormat(format)
	if mimeType == "" {
		return "", unifiederrors.ValidationErrorf("unknown %s format: %s", label, format)
	}
	return mimeType, nil
}

func buildRunContext(inputs []runInput) (map[string]interface{}, error) {
	context := make(map[string]interface{})
	for _, input := range inputs {
		name := strings.TrimSpace(input.Name)
		if name == "" {
			return nil, unifiederrors.ValidationError("input name is required")
		}

		mimeType, err := normalizeMimeType(input.Format, "input")
		if err != nil {
			return nil, unifiederrors.ValidationErrorf("input %q: %v", name, err)
		}

		value, err := formats.Read(input.Content, mimeType)
		if err != nil {
			return nil, unifiederrors.ValidationErrorf("input %q: %v", name, err)
		}

		context[name] = value
	}
	return context, nil
}

func formatRunResult(result interface{}, mimeType string, evalCtx map[string]interface{}) (string, error) {
	if mimeType == "application/xml" {
		if namespaces, ok := evalCtx["__namespaces__"].(map[string]string); ok {
			return formats.FormatXMLWithNamespaces(result, namespaces)
		}
	}
	return formats.Format(result, mimeType)
}
