package evaluator

import (
	"context"
	"sort"
	"time"
)

const deadlineContextKey = "__deadline"

var reservedBindingNames = map[string]struct{}{
	GoContextKey:             {},
	deadlineContextKey:       {},
	formatServiceContextKey:  {},
	urlReadServiceContextKey: {},
	outputMetadataContextKey: {},
	inputMimeContextKey:      {},
}

const (
	outputMetadataContextKey = "__output_metadata__"
	inputMimeContextKey      = "__input_mime__"
)

// Runtime carries non-user execution capabilities for evaluation.
type Runtime struct {
	GoContext          context.Context
	Deadline           time.Time
	HasDeadline        bool
	FormatService      FormatService
	URLReadService     URLReadService
	ExpressionCompiler ExpressionCompiler
}

// Scope separates user/module variables from runtime capabilities.
type Scope struct {
	Vars    Context
	Runtime Runtime
}

// IsReservedBindingName reports names that are reserved for runtime metadata.
func IsReservedBindingName(name string) bool {
	_, ok := reservedBindingNames[name]
	return ok
}

// ReservedBindingNames returns reserved runtime metadata names in stable order.
func ReservedBindingNames() []string {
	names := make([]string, 0, len(reservedBindingNames))
	for name := range reservedBindingNames {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// NewScope wraps variables in an evaluation scope and extracts legacy runtime
// context keys at the API boundary.
func NewScope(vars Context) *Scope {
	if vars == nil {
		vars = make(Context)
	}
	scope := &Scope{Vars: vars}
	scope.Runtime.GoContext = context.Background()
	scope.extractLegacyRuntime()
	return scope
}

// NewScopeWithRuntime creates a scope with explicit runtime capabilities.
func NewScopeWithRuntime(vars Context, runtime Runtime) *Scope {
	if vars == nil {
		vars = make(Context)
	}
	if runtime.GoContext == nil {
		runtime.GoContext = context.Background()
	}
	return &Scope{Vars: vars, Runtime: runtime}
}

func (s *Scope) ensure() {
	if s.Vars == nil {
		s.Vars = make(Context)
	}
	if s.Runtime.GoContext == nil {
		s.Runtime.GoContext = context.Background()
	}
}

// Copy creates a child scope with isolated variables and shared runtime
// capabilities.
func (s *Scope) Copy() *Scope {
	if s == nil {
		return NewScope(nil)
	}
	s.ensure()
	return NewScopeWithRuntime(copyContext(s.Vars), s.Runtime)
}

// WithGoContext sets the Go context used by lazy values, streams, and IO.
func (s *Scope) WithGoContext(goCtx context.Context) *Scope {
	if s == nil {
		s = NewScope(nil)
	}
	s.ensure()
	if goCtx == nil {
		goCtx = context.Background()
	}
	s.Runtime.GoContext = goCtx
	if deadline, ok := goCtx.Deadline(); ok {
		s.Runtime.Deadline = deadline
		s.Runtime.HasDeadline = true
	}
	return s
}

// GoContext returns the configured Go context or context.Background().
func (s *Scope) GoContext() context.Context {
	if s == nil || s.Runtime.GoContext == nil {
		return context.Background()
	}
	return s.Runtime.GoContext
}

// SetExpressionCompiler sets the compiler used by nested runtime constructs
// such as do blocks and update cases.
func (s *Scope) SetExpressionCompiler(compiler ExpressionCompiler) *Scope {
	if s == nil {
		s = NewScope(nil)
	}
	s.ensure()
	s.Runtime.ExpressionCompiler = compiler
	return s
}

// ExpressionCompiler returns the configured compiler or a parser-only fallback
// for direct evaluator tests that pass already rewritten expressions.
func (s *Scope) ExpressionCompiler() ExpressionCompiler {
	if s == nil || s.Runtime.ExpressionCompiler == nil {
		return parseExpressionCompiler{}
	}
	return s.Runtime.ExpressionCompiler
}

// LoopDeadline returns the active evaluation deadline or the default loop timeout.
func (s *Scope) LoopDeadline(startTime time.Time) time.Time {
	if s != nil && s.Runtime.HasDeadline {
		return s.Runtime.Deadline
	}
	return startTime.Add(DefaultLoopTimeout)
}

func (s *Scope) extractLegacyRuntime() {
	if s == nil || s.Vars == nil {
		return
	}
	if goCtx, ok := s.Vars[GoContextKey].(context.Context); ok {
		s.WithGoContext(goCtx)
		delete(s.Vars, GoContextKey)
	}
	if deadline, ok := s.Vars[deadlineContextKey].(time.Time); ok {
		s.Runtime.Deadline = deadline
		s.Runtime.HasDeadline = true
		delete(s.Vars, deadlineContextKey)
	}
	if service, ok := s.Vars[formatServiceContextKey].(FormatService); ok {
		s.Runtime.FormatService = service
		delete(s.Vars, formatServiceContextKey)
	}
	if service, ok := s.Vars[urlReadServiceContextKey].(URLReadService); ok {
		s.Runtime.URLReadService = service
		delete(s.Vars, urlReadServiceContextKey)
	}
}
