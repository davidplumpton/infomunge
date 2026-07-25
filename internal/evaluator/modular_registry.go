package evaluator

import (
	"fmt"
	"go/ast"
	"go/token"
	"sort"
	"sync"
)

type SpecialBuiltinFunc = func(*ast.CallExpr, *Scope, int) (Value, error)
type RegularBuiltinFunc = func([]Value, *ast.CallExpr) (Value, error)

type BuiltinEvaluationMode string

const (
	BuiltinEvaluationEager   BuiltinEvaluationMode = "eager"
	BuiltinEvaluationSpecial BuiltinEvaluationMode = "special"

	BuiltinArityVariadic = -1
)

const (
	builtinCategoryArrays     = "arrays"
	builtinCategoryAssertions = "assertions"
	builtinCategoryCore       = "core"
	builtinCategoryDates      = "dates"
	builtinCategoryIO         = "io"
	builtinCategoryNumbers    = "numbers"
	builtinCategoryObjects    = "objects"
	builtinCategoryRuntime    = "runtime"
	builtinCategoryStrings    = "strings"
	builtinCategoryURLs       = "urls"
)

// BuiltinArity describes the accepted argument count for builtin metadata.
// Max == BuiltinArityVariadic means no upper bound.
type BuiltinArity struct {
	Min int
	Max int
}

// BuiltinSpec is the domain-owned descriptor used to build builtin dispatch
// tables and expose builtin metadata for documentation/tests.
type BuiltinSpec struct {
	Name     string
	Category string
	Module   string
	Arity    BuiltinArity
	Mode     BuiltinEvaluationMode
	Regular  RegularBuiltinFunc
	Special  SpecialBuiltinFunc
	Summary  string

	tooFewArgumentsMessage  string
	tooManyArgumentsMessage string
}

// builtinSpecialRegistry maps special builtin function names to their handlers.
// Special functions need unevaluated arguments and are called via GetBuiltinSpecial.
var builtinSpecialRegistry map[string]SpecialBuiltinFunc

// builtinFunctionRegistry maps regular builtin function names to their handlers.
// Regular functions receive evaluated arguments and are called via GetBuiltinFunction.
var builtinFunctionRegistry map[string]RegularBuiltinFunc

// builtinSpecRegistry keeps metadata for the currently installed builtins.
var builtinSpecRegistry map[string]BuiltinSpec

// builtinRegistryMu protects builtin registries for concurrent reads/writes.
var builtinRegistryMu sync.RWMutex

func init() {
	special, regular, specs, err := buildBuiltinRegistries(defaultBuiltinSpecs())
	if err != nil {
		panic(fmt.Sprintf("invalid builtin registry: %v", err))
	}
	builtinSpecialRegistry = special
	builtinFunctionRegistry = regular
	builtinSpecRegistry = specs
}

func defaultBuiltinSpecs() []BuiltinSpec {
	groups := []func() []BuiltinSpec{
		coreBuiltinSpecs,
		controlFlowBuiltinSpecs,
		collectionBuiltinSpecs,
		stringBuiltinSpecs,
		numberBuiltinSpecs,
		dateBuiltinSpecs,
		runtimeBuiltinSpecs,
		ioBuiltinSpecs,
		urlBuiltinSpecs,
		assertionBuiltinSpecs,
	}

	var specs []BuiltinSpec
	for _, group := range groups {
		specs = append(specs, group()...)
	}
	return specs
}

func exactArity(n int) BuiltinArity {
	return BuiltinArity{Min: n, Max: n}
}

func rangeArity(min, max int) BuiltinArity {
	return BuiltinArity{Min: min, Max: max}
}

func variadicArity(min int) BuiltinArity {
	return BuiltinArity{Min: min, Max: BuiltinArityVariadic}
}

func anyArity() BuiltinArity {
	return variadicArity(0)
}

func (a BuiltinArity) valid() bool {
	if a.Min < 0 {
		return false
	}
	return a.Max == BuiltinArityVariadic || a.Max >= a.Min
}

func (a BuiltinArity) accepts(count int) bool {
	return count >= a.Min && (a.Max == BuiltinArityVariadic || count <= a.Max)
}

func regularBuiltinSpec(name, category string, arity BuiltinArity, fn RegularBuiltinFunc, summary string) BuiltinSpec {
	return BuiltinSpec{
		Name:     name,
		Category: category,
		Module:   builtinModuleForCategory(category),
		Arity:    arity,
		Mode:     BuiltinEvaluationEager,
		Regular:  fn,
		Summary:  builtinSummary(name, summary),
	}
}

func specialBuiltinSpec(name, category string, arity BuiltinArity, fn SpecialBuiltinFunc, summary string) BuiltinSpec {
	return BuiltinSpec{
		Name:     name,
		Category: category,
		Module:   builtinModuleForCategory(category),
		Arity:    arity,
		Mode:     BuiltinEvaluationSpecial,
		Special:  fn,
		Summary:  builtinSummary(name, summary),
	}
}

func withArityMessages(spec BuiltinSpec, tooFew, tooMany string) BuiltinSpec {
	spec.tooFewArgumentsMessage = tooFew
	spec.tooManyArgumentsMessage = tooMany
	return spec
}

func builtinSummary(name, summary string) string {
	if summary != "" {
		return summary
	}
	return name
}

func builtinModuleForCategory(category string) string {
	switch category {
	case builtinCategoryArrays:
		return "dw::core::Arrays"
	case builtinCategoryDates:
		return "dw::core::Dates"
	case builtinCategoryNumbers:
		return "dw::core::Numbers"
	case builtinCategoryObjects:
		return "dw::core::Objects"
	case builtinCategoryStrings:
		return "dw::core::Strings"
	case builtinCategoryURLs:
		return "dw::core::URL"
	default:
		return "internal::" + category
	}
}

func buildBuiltinRegistries(specs []BuiltinSpec) (
	map[string]SpecialBuiltinFunc,
	map[string]RegularBuiltinFunc,
	map[string]BuiltinSpec,
	error,
) {
	special := make(map[string]SpecialBuiltinFunc)
	regular := make(map[string]RegularBuiltinFunc)
	byName := make(map[string]BuiltinSpec)

	for _, spec := range specs {
		if err := validateBuiltinSpec(spec); err != nil {
			return nil, nil, nil, err
		}
		if _, exists := byName[spec.Name]; exists {
			return nil, nil, nil, fmt.Errorf("duplicate builtin %q", spec.Name)
		}

		registeredSpec := spec
		switch spec.Mode {
		case BuiltinEvaluationSpecial:
			registeredSpec.Special = enforceSpecialBuiltinArity(spec)
			special[spec.Name] = registeredSpec.Special
		case BuiltinEvaluationEager:
			registeredSpec.Regular = enforceRegularBuiltinArity(spec)
			regular[spec.Name] = registeredSpec.Regular
		}
		byName[spec.Name] = registeredSpec
	}

	return special, regular, byName, nil
}

func enforceRegularBuiltinArity(spec BuiltinSpec) RegularBuiltinFunc {
	return func(args []Value, e *ast.CallExpr) (Value, error) {
		if !spec.Arity.accepts(len(args)) {
			return nil, builtinArityError(spec, len(args), callExprPos(e))
		}
		return spec.Regular(args, e)
	}
}

func enforceSpecialBuiltinArity(spec BuiltinSpec) SpecialBuiltinFunc {
	return func(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
		count := 0
		if e != nil {
			count = len(e.Args)
		}
		if !spec.Arity.accepts(count) {
			return nil, builtinArityError(spec, count, callExprPos(e))
		}
		return spec.Special(e, scope, depth)
	}
}

func builtinArityError(spec BuiltinSpec, count int, pos token.Pos) error {
	if count < spec.Arity.Min && spec.tooFewArgumentsMessage != "" {
		return newPosError(spec.tooFewArgumentsMessage, pos)
	}
	if count > spec.Arity.Max && spec.Arity.Max != BuiltinArityVariadic && spec.tooManyArgumentsMessage != "" {
		return newPosError(spec.tooManyArgumentsMessage, pos)
	}

	name := spec.Name
	if spec.Category == builtinCategoryAssertions {
		name += "()"
	}

	var message string
	switch {
	case spec.Arity.Min == 0 && spec.Arity.Max == 0:
		message = fmt.Sprintf("%s takes no arguments", name)
	case spec.Arity.Min == spec.Arity.Max:
		message = fmt.Sprintf("%s requires exactly %d %s", name, spec.Arity.Min, argumentWord(spec.Arity.Min))
	case spec.Arity.Min == 0:
		message = fmt.Sprintf("%s accepts at most %d %s", name, spec.Arity.Max, argumentWord(spec.Arity.Max))
	case spec.Arity.Max == BuiltinArityVariadic:
		message = fmt.Sprintf("%s requires at least %d %s", name, spec.Arity.Min, argumentWord(spec.Arity.Min))
	default:
		message = fmt.Sprintf("%s requires %d or %d arguments", name, spec.Arity.Min, spec.Arity.Max)
	}
	return newPosError(message, pos)
}

func argumentWord(count int) string {
	if count == 1 {
		return "argument"
	}
	return "arguments"
}

func callExprPos(e *ast.CallExpr) token.Pos {
	if e == nil {
		return token.NoPos
	}
	return e.Pos()
}

func validateBuiltinSpec(spec BuiltinSpec) error {
	if spec.Name == "" {
		return fmt.Errorf("builtin spec has empty name")
	}
	if spec.Category == "" {
		return fmt.Errorf("builtin %q has empty category", spec.Name)
	}
	if spec.Module == "" {
		return fmt.Errorf("builtin %q has empty module", spec.Name)
	}
	if !spec.Arity.valid() {
		return fmt.Errorf("builtin %q has invalid arity", spec.Name)
	}
	if spec.Summary == "" {
		return fmt.Errorf("builtin %q has empty summary", spec.Name)
	}

	switch spec.Mode {
	case BuiltinEvaluationSpecial:
		if spec.Special == nil {
			return fmt.Errorf("special builtin %q has nil handler", spec.Name)
		}
		if spec.Regular != nil {
			return fmt.Errorf("special builtin %q also has a regular handler", spec.Name)
		}
	case BuiltinEvaluationEager:
		if spec.Regular == nil {
			return fmt.Errorf("eager builtin %q has nil handler", spec.Name)
		}
		if spec.Special != nil {
			return fmt.Errorf("eager builtin %q also has a special handler", spec.Name)
		}
	default:
		return fmt.Errorf("builtin %q has unknown evaluation mode %q", spec.Name, spec.Mode)
	}

	return nil
}

func cloneSpecialRegistry(src map[string]SpecialBuiltinFunc) map[string]SpecialBuiltinFunc {
	dst := make(map[string]SpecialBuiltinFunc, len(src))
	for name, fn := range src {
		dst[name] = fn
	}
	return dst
}

func cloneFunctionRegistry(src map[string]RegularBuiltinFunc) map[string]RegularBuiltinFunc {
	dst := make(map[string]RegularBuiltinFunc, len(src))
	for name, fn := range src {
		dst[name] = fn
	}
	return dst
}

func cloneSpecRegistry(src map[string]BuiltinSpec) map[string]BuiltinSpec {
	dst := make(map[string]BuiltinSpec, len(src))
	for name, spec := range src {
		dst[name] = spec
	}
	return dst
}

func specRegistryFromMaps(
	special map[string]SpecialBuiltinFunc,
	regular map[string]RegularBuiltinFunc,
) map[string]BuiltinSpec {
	specs := make(map[string]BuiltinSpec, len(special)+len(regular))
	for name, fn := range regular {
		specs[name] = regularBuiltinSpec(name, builtinCategoryRuntime, anyArity(), fn, "runtime registered builtin")
	}
	for name, fn := range special {
		specs[name] = specialBuiltinSpec(name, builtinCategoryRuntime, anyArity(), fn, "runtime registered special builtin")
	}
	return specs
}

// RegisterBuiltinSpecial registers or overrides a special builtin at runtime.
func RegisterBuiltinSpecial(name string, fn SpecialBuiltinFunc) {
	builtinRegistryMu.Lock()
	defer builtinRegistryMu.Unlock()
	builtinSpecialRegistry[name] = fn
	builtinSpecRegistry[name] = specialBuiltinSpec(name, builtinCategoryRuntime, anyArity(), fn, "runtime registered special builtin")
}

// RegisterBuiltinFunction registers or overrides a regular builtin at runtime.
func RegisterBuiltinFunction(name string, fn RegularBuiltinFunc) {
	builtinRegistryMu.Lock()
	defer builtinRegistryMu.Unlock()
	builtinFunctionRegistry[name] = fn
	if _, hasSpecial := builtinSpecialRegistry[name]; !hasSpecial {
		builtinSpecRegistry[name] = regularBuiltinSpec(name, builtinCategoryRuntime, anyArity(), fn, "runtime registered builtin")
	}
}

// SetBuiltinRegistriesForTesting swaps builtin registries and returns a restore func.
// The provided maps are cloned to avoid external mutation races.
func SetBuiltinRegistriesForTesting(
	special map[string]SpecialBuiltinFunc,
	regular map[string]RegularBuiltinFunc,
) func() {
	builtinRegistryMu.Lock()
	prevSpecial := cloneSpecialRegistry(builtinSpecialRegistry)
	prevRegular := cloneFunctionRegistry(builtinFunctionRegistry)
	prevSpecs := cloneSpecRegistry(builtinSpecRegistry)
	builtinSpecialRegistry = cloneSpecialRegistry(special)
	builtinFunctionRegistry = cloneFunctionRegistry(regular)
	builtinSpecRegistry = specRegistryFromMaps(builtinSpecialRegistry, builtinFunctionRegistry)
	builtinRegistryMu.Unlock()

	return func() {
		builtinRegistryMu.Lock()
		builtinSpecialRegistry = prevSpecial
		builtinFunctionRegistry = prevRegular
		builtinSpecRegistry = prevSpecs
		builtinRegistryMu.Unlock()
	}
}

// GetBuiltinSpecial returns a special function handler by name.
// This maintains compatibility with existing visitor code.
func GetBuiltinSpecial(name string) (SpecialBuiltinFunc, bool) {
	builtinRegistryMu.RLock()
	defer builtinRegistryMu.RUnlock()
	fn, ok := builtinSpecialRegistry[name]
	return fn, ok
}

// GetBuiltinFunction returns a regular function handler by name.
func GetBuiltinFunction(name string) (RegularBuiltinFunc, bool) {
	builtinRegistryMu.RLock()
	defer builtinRegistryMu.RUnlock()
	fn, ok := builtinFunctionRegistry[name]
	return fn, ok
}

func GetBuiltinSpec(name string) (BuiltinSpec, bool) {
	builtinRegistryMu.RLock()
	defer builtinRegistryMu.RUnlock()
	spec, ok := builtinSpecRegistry[name]
	return spec, ok
}

func ListBuiltinSpecs() []BuiltinSpec {
	builtinRegistryMu.RLock()
	defer builtinRegistryMu.RUnlock()

	specs := make([]BuiltinSpec, 0, len(builtinSpecRegistry))
	for _, spec := range builtinSpecRegistry {
		specs = append(specs, spec)
	}
	sortBuiltinSpecs(specs)
	return specs
}

func BuiltinSpecsByCategory() map[string][]BuiltinSpec {
	grouped := make(map[string][]BuiltinSpec)
	for _, spec := range ListBuiltinSpecs() {
		grouped[spec.Category] = append(grouped[spec.Category], spec)
	}
	return grouped
}

func sortBuiltinSpecs(specs []BuiltinSpec) {
	sort.Slice(specs, func(i, j int) bool {
		if specs[i].Category != specs[j].Category {
			return specs[i].Category < specs[j].Category
		}
		return specs[i].Name < specs[j].Name
	})
}
