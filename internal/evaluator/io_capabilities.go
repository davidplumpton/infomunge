package evaluator

import (
	"context"
	"fmt"
)

const (
	formatServiceContextKey  = "__format_service__"
	urlReadServiceContextKey = "__url_read_service__"
)

// FormatService is the explicit boundary between evaluator builtins and
// serialization adapters.
type FormatService interface {
	Read(content, mimeType string) (Value, error)
	ReadWithOptions(content, mimeType string, options Object) (Value, error)
	Write(value Value, mimeType string) (string, error)
	WriteWithOptions(value Value, mimeType string, options Object) (string, error)
}

// URLReadService is the optional environmental capability behind readUrl.
type URLReadService interface {
	ReadURL(ctx context.Context, rawURL, mimeType string) (Value, error)
}

type disabledURLReadService struct{}

func (disabledURLReadService) ReadURL(context.Context, string, string) (Value, error) {
	return nil, fmt.Errorf("readUrl: URL IO capability is disabled")
}

// WithFormatService installs a format service into an evaluation context.
func WithFormatService(ctx Context, service FormatService) Context {
	if ctx == nil {
		ctx = make(Context)
	}
	if service == nil {
		delete(ctx, formatServiceContextKey)
		return ctx
	}
	ctx[formatServiceContextKey] = service
	return ctx
}

// WithURLReadService installs a URL read service into an evaluation context.
func WithURLReadService(ctx Context, service URLReadService) Context {
	if ctx == nil {
		ctx = make(Context)
	}
	if service == nil {
		delete(ctx, urlReadServiceContextKey)
		return ctx
	}
	ctx[urlReadServiceContextKey] = service
	return ctx
}

// WithURLReadDisabled installs an explicit disabled readUrl capability.
func WithURLReadDisabled(ctx Context) Context {
	return WithURLReadService(ctx, disabledURLReadService{})
}

// GetFormatService returns the format service installed in an evaluation context.
func GetFormatService(ctx Context) (FormatService, bool) {
	if ctx == nil {
		return nil, false
	}
	service, ok := ctx[formatServiceContextKey].(FormatService)
	return service, ok
}

// GetURLReadService returns the URL read service installed in an evaluation context.
func GetURLReadService(ctx Context) (URLReadService, bool) {
	if ctx == nil {
		return nil, false
	}
	service, ok := ctx[urlReadServiceContextKey].(URLReadService)
	return service, ok
}

func requireFormatService(ctx Context) (FormatService, error) {
	service, ok := GetFormatService(ctx)
	if !ok {
		return nil, fmt.Errorf("format service is unavailable")
	}
	return service, nil
}

func requireURLReadService(ctx Context) (URLReadService, error) {
	service, ok := GetURLReadService(ctx)
	if !ok {
		return nil, fmt.Errorf("readUrl: URL IO capability is unavailable")
	}
	return service, nil
}

func copyEvaluationCapabilities(dst, src Context) {
	if dst == nil || src == nil {
		return
	}
	for _, key := range []string{formatServiceContextKey, urlReadServiceContextKey} {
		if value, ok := src[key]; ok {
			dst[key] = value
		}
	}
}
