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

// DisabledURLReadService returns a URLReadService that rejects all URL IO.
func DisabledURLReadService() URLReadService {
	return disabledURLReadService{}
}

// WithFormatService installs a format service into an evaluation context for
// legacy callers. New evaluator code carries this capability on Scope.Runtime.
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

// WithURLReadService installs a URL read service into an evaluation context for
// legacy callers. New evaluator code carries this capability on Scope.Runtime.
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

// GetFormatService returns a legacy format service installed in an evaluation context.
func GetFormatService(ctx Context) (FormatService, bool) {
	if ctx == nil {
		return nil, false
	}
	service, ok := ctx[formatServiceContextKey].(FormatService)
	return service, ok
}

// GetURLReadService returns a legacy URL read service installed in an evaluation context.
func GetURLReadService(ctx Context) (URLReadService, bool) {
	if ctx == nil {
		return nil, false
	}
	service, ok := ctx[urlReadServiceContextKey].(URLReadService)
	return service, ok
}

func (s *Scope) SetFormatService(service FormatService) {
	if s == nil {
		return
	}
	s.ensure()
	s.Runtime.FormatService = service
}

func (s *Scope) SetURLReadService(service URLReadService) {
	if s == nil {
		return
	}
	s.ensure()
	s.Runtime.URLReadService = service
}

func (s *Scope) FormatService() (FormatService, bool) {
	if s == nil || s.Runtime.FormatService == nil {
		return nil, false
	}
	return s.Runtime.FormatService, true
}

func (s *Scope) URLReadService() (URLReadService, bool) {
	if s == nil || s.Runtime.URLReadService == nil {
		return nil, false
	}
	return s.Runtime.URLReadService, true
}

func requireFormatService(scope *Scope) (FormatService, error) {
	service, ok := scope.FormatService()
	if !ok {
		return nil, fmt.Errorf("format service is unavailable")
	}
	return service, nil
}

func requireURLReadService(scope *Scope) (URLReadService, error) {
	service, ok := scope.URLReadService()
	if !ok {
		return nil, fmt.Errorf("readUrl: URL IO capability is unavailable")
	}
	return service, nil
}
