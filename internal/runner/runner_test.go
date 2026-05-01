package runner

import (
	"errors"
	"infomunge/internal/evaluator"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunString(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		context  evaluator.Context
		want     evaluator.Value
		wantErr  bool
		errMatch string
	}{
		{
			name:   "simple expression",
			script: "%im 0.1\noutput application/json\n---\n1 + 2",
			want:   3,
		},
		{
			name:   "with variable",
			script: "%im 0.1\noutput application/json\nvar x = 10\n---\nx * 2",
			want:   20,
		},
		{
			name:    "with context variable",
			script:  "%im 0.1\noutput application/json\n---\npayload.value",
			context: evaluator.Object{"payload": evaluator.Object{"value": 42}},
			want:    42,
		},
		{
			name:   "array map",
			script: "%im 0.1\noutput application/json\n---\n[1, 2, 3] map $ * 2",
			want:   evaluator.Array{2, 4, 6},
		},
		{
			name:   "function declaration",
			script: "%im 0.1\noutput application/json\nfun double(x) = x * 2\n---\ndouble(5)",
			want:   10,
		},
		{
			name:     "invalid expression",
			script:   "%im 0.1\noutput application/json\n---\n1 +",
			wantErr:  true,
			errMatch: "expected operand",
		},
		{
			name:   "simple string literal",
			script: "%im 0.1\noutput application/json\n---\n\"hello world\"",
			want:   "hello world",
		},
		{
			name:   "single-line script",
			script: `%im 0.1 var myObject = { user : "a" } output application/json --- { myObjectExample : myObject.user }`,
			want:   evaluator.Object{"myObjectExample": "a"},
		},
		{
			name:   "single-line with function",
			script: `%im 0.1 fun double(x) = x * 2 output application/json --- double(5)`,
			want:   10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RunString(tt.script, tt.context)
			if tt.wantErr {
				if err == nil {
					t.Errorf("RunString() expected error, got nil")
					return
				}
				if tt.errMatch != "" && !containsString(err.Error(), tt.errMatch) {
					t.Errorf("RunString() error = %v, want error containing %q", err, tt.errMatch)
				}
				return
			}
			if err != nil {
				t.Errorf("RunString() unexpected error: %v", err)
				return
			}
			if !deepEqual(got, tt.want) {
				t.Errorf("RunString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecuteString_LazyFlagUnsupported(t *testing.T) {
	script := "%im 0.1\noutput application/json\n---\n1 + 1"
	_, err := ExecuteString(t.Context(), script, nil, RunnerOptions{Lazy: true})
	if err == nil {
		t.Fatalf("expected error when Lazy option is enabled")
	}
	if !containsString(err.Error(), "--lazy is currently unsupported") {
		t.Fatalf("expected unsupported lazy flag error, got: %v", err)
	}
}

func TestRunStringWithOptions_DisablesURLIOInModuleDeclarations(t *testing.T) {
	tmpDir := t.TempDir()
	modulesDir := filepath.Join(tmpDir, "modules")
	if err := os.Mkdir(modulesDir, 0755); err != nil {
		t.Fatalf("failed to create modules dir: %v", err)
	}
	module := `%im 0.1
var remote = readUrl("http://example.com/data.json", "application/json")`
	if err := os.WriteFile(filepath.Join(modulesDir, "Remote.im"), []byte(module), 0644); err != nil {
		t.Fatalf("failed to write module: %v", err)
	}

	script := `%im 0.1
import modules::Remote
output application/json
---
Remote::remote`
	_, err := ExecuteString(
		t.Context(),
		script,
		nil,
		RunnerOptions{BaseDir: tmpDir, DisableURLReadService: true},
	)
	if err == nil || !errorChainContains(err, "URL IO capability is disabled") {
		t.Fatalf("expected disabled URL IO error, got %v", err)
	}
}

func errorChainContains(err error, expected string) bool {
	for err != nil {
		if strings.Contains(err.Error(), expected) {
			return true
		}
		err = errors.Unwrap(err)
	}
	return false
}

func TestParseVarDecl(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		trimmed  string
		context  evaluator.Context
		wantVal  evaluator.Value
		wantName string
		wantErr  bool
		errMatch string
	}{
		{
			name:     "simple integer",
			line:     "var x = 42",
			trimmed:  "var x = 42",
			context:  make(evaluator.Context),
			wantVal:  42,
			wantName: "x",
		},
		{
			name:     "string value",
			line:     "var name = \"hello\"",
			trimmed:  "var name = \"hello\"",
			context:  make(evaluator.Context),
			wantVal:  "hello",
			wantName: "name",
		},
		{
			name:     "expression value",
			line:     "var result = 10 + 5",
			trimmed:  "var result = 10 + 5",
			context:  make(evaluator.Context),
			wantVal:  15,
			wantName: "result",
		},
		{
			name:     "using context",
			line:     "var doubled = x * 2",
			trimmed:  "var doubled = x * 2",
			context:  evaluator.Object{"x": 5},
			wantVal:  10,
			wantName: "doubled",
		},
		{
			name:     "array value",
			line:     "var arr = [1, 2, 3]",
			trimmed:  "var arr = [1, 2, 3]",
			context:  make(evaluator.Context),
			wantVal:  evaluator.Array{1, 2, 3},
			wantName: "arr",
		},
		{
			name:     "missing equals sign",
			line:     "var x",
			trimmed:  "var x",
			context:  make(evaluator.Context),
			wantErr:  true,
			errMatch: "missing '='",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, name, err := parseVarDecl(tt.line, tt.trimmed, 0, tt.context, tt.line)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseVarDecl() expected error, got nil")
				}
				if tt.errMatch != "" && !containsString(err.Error(), tt.errMatch) {
					t.Errorf("parseVarDecl() error = %v, want error containing %q", err, tt.errMatch)
				}
				return
			}
			if err != nil {
				t.Errorf("parseVarDecl() unexpected error: %v", err)
				return
			}
			if name != tt.wantName {
				t.Errorf("parseVarDecl() name = %q, want %q", name, tt.wantName)
			}
			if tt.wantName != "" && !deepEqual(val, tt.wantVal) {
				t.Errorf("parseVarDecl() val = %v, want %v", val, tt.wantVal)
			}
		})
	}
}

func TestParseFunDecl(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		wantName string
		wantErr  bool
		errMatch string
	}{
		{
			name:     "simple function",
			line:     "fun double(x) = x * 2",
			wantName: "double",
		},
		{
			name:     "no params",
			line:     "fun getOne() = 1",
			wantName: "getOne",
		},
		{
			name:     "multiple params",
			line:     "fun add(a, b) = a + b",
			wantName: "add",
		},
		{
			name:     "complex body",
			line:     "fun transform(obj) = { name: obj.firstName }",
			wantName: "transform",
		},
		{
			name:     "missing params",
			line:     "fun test = 1",
			wantErr:  true,
			errMatch: "missing parameter list",
		},
		{
			name:     "missing name",
			line:     "fun () = 1",
			wantErr:  true,
			errMatch: "missing function name",
		},
		{
			name:     "missing equals",
			line:     "fun test() 1",
			wantErr:  true,
			errMatch: "missing '='",
		},
		{
			name:     "empty body",
			line:     "fun test() =",
			wantErr:  true,
			errMatch: "empty body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lambda, name, err := parseFunDecl(tt.line, nil)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseFunDecl() expected error, got nil")
					return
				}
				if tt.errMatch != "" && !containsString(err.Error(), tt.errMatch) {
					t.Errorf("parseFunDecl() error = %v, want error containing %q", err, tt.errMatch)
				}
				return
			}
			if err != nil {
				t.Errorf("parseFunDecl() unexpected error: %v", err)
				return
			}
			if name != tt.wantName {
				t.Errorf("parseFunDecl() name = %q, want %q", name, tt.wantName)
			}
			if lambda == nil {
				t.Errorf("parseFunDecl() lambda = nil, want non-nil")
			}
		})
	}
}

func TestModuleLoaderLoadArraysModule(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("filepath.Abs repo root: %v", err)
	}
	loader := NewModuleLoader(repoRoot)
	module, err := loader.Load("dw::core::Arrays")
	if err != nil {
		t.Fatalf("Load(dw::core::Arrays) unexpected error: %v", err)
	}
	if module == nil {
		t.Fatal("Load(dw::core::Arrays) returned nil module")
	}
	if module.Name != "Arrays" {
		t.Fatalf("Load(dw::core::Arrays) module name = %q, want %q", module.Name, "Arrays")
	}
	if _, ok := module.Namespace["outerJoin"]; !ok {
		t.Fatal("Load(dw::core::Arrays) missing outerJoin")
	}
}

func TestParseFunDeclFromLines_DoBlockSeparator(t *testing.T) {
	lines := []string{
		"fun addOne(x) = do {",
		"  var y = x + 1",
		"  ---",
		"  y",
		"}",
		"output application/json",
	}
	fn, name, consumed, err := parseFunDeclFromLines(lines, 0, nil)
	if err != nil {
		t.Fatalf("parseFunDeclFromLines() unexpected error: %v", err)
	}
	if name != "addOne" {
		t.Fatalf("parseFunDeclFromLines() name = %q, want %q", name, "addOne")
	}
	if consumed != 5 {
		t.Fatalf("parseFunDeclFromLines() consumed = %d, want %d", consumed, 5)
	}
	if fn == nil {
		t.Fatal("parseFunDeclFromLines() fn = nil, want non-nil")
	}
}

func TestParseTypeDecl(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		wantName  string
		wantBase  string
		wantProps evaluator.Object
		wantErr   bool
		errMatch  string
	}{
		{
			name:     "simple type",
			line:     "type MyString = String",
			wantName: "MyString",
			wantBase: "String",
		},
		{
			name:      "type with properties",
			line:      `type Currency = String { format: "##.00" }`,
			wantName:  "Currency",
			wantBase:  "String",
			wantProps: evaluator.Object{"format": "##.00"},
		},
		{
			name:     "missing equals",
			line:     "type MyType String",
			wantErr:  true,
			errMatch: "missing '='",
		},
		{
			name:     "missing name",
			line:     "type  = String",
			wantErr:  true,
			errMatch: "missing type name",
		},
		{
			name:     "missing base type",
			line:     "type MyType =",
			wantErr:  true,
			errMatch: "missing base type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typeDef, name, err := parseTypeDecl(tt.line)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseTypeDecl() expected error, got nil")
					return
				}
				if tt.errMatch != "" && !containsString(err.Error(), tt.errMatch) {
					t.Errorf("parseTypeDecl() error = %v, want error containing %q", err, tt.errMatch)
				}
				return
			}
			if err != nil {
				t.Errorf("parseTypeDecl() unexpected error: %v", err)
				return
			}
			if name != tt.wantName {
				t.Errorf("parseTypeDecl() name = %q, want %q", name, tt.wantName)
			}
			if typeDef.BaseType != tt.wantBase {
				t.Errorf("parseTypeDecl() baseType = %q, want %q", typeDef.BaseType, tt.wantBase)
			}
		})
	}
}

func TestParseNamespaceDecl(t *testing.T) {
	tests := []struct {
		name       string
		line       string
		wantPrefix string
		wantURI    string
		wantErr    bool
	}{
		{
			name:       "prefixed namespace",
			line:       "ns ns0 http://example.com",
			wantPrefix: "ns0",
			wantURI:    "http://example.com",
		},
		{
			name:       "default namespace",
			line:       "ns http://example.com",
			wantPrefix: "",
			wantURI:    "http://example.com",
		},
		{
			name:    "invalid - too many parts",
			line:    "ns a b c",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix, uri, err := parseNamespaceDecl(tt.line)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseNamespaceDecl() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("parseNamespaceDecl() unexpected error: %v", err)
				return
			}
			if prefix != tt.wantPrefix {
				t.Errorf("parseNamespaceDecl() prefix = %q, want %q", prefix, tt.wantPrefix)
			}
			if uri != tt.wantURI {
				t.Errorf("parseNamespaceDecl() uri = %q, want %q", uri, tt.wantURI)
			}
		})
	}
}

func TestModuleLoader(t *testing.T) {
	// Create a temporary directory with a module
	tmpDir, err := os.MkdirTemp("", "runner_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a simple module file
	moduleContent := `%im 0.1
var PI = 3.14159
fun double(x) = x * 2
`
	modulePath := filepath.Join(tmpDir, "MyModule.im")
	if err := os.WriteFile(modulePath, []byte(moduleContent), 0644); err != nil {
		t.Fatalf("Failed to write module file: %v", err)
	}

	t.Run("NewModuleLoader", func(t *testing.T) {
		loader := NewModuleLoader(tmpDir)
		if loader.BaseDir != tmpDir {
			t.Errorf("NewModuleLoader() BaseDir = %q, want %q", loader.BaseDir, tmpDir)
		}
		if len(loader.SearchPaths) < 1 {
			t.Errorf("NewModuleLoader() SearchPaths should have at least 1 entry")
		}
	})

	t.Run("Resolve", func(t *testing.T) {
		loader := NewModuleLoader(tmpDir)
		name, path, err := loader.Resolve("MyModule")
		if err != nil {
			t.Errorf("Resolve() unexpected error: %v", err)
			return
		}
		if name != "MyModule" {
			t.Errorf("Resolve() name = %q, want %q", name, "MyModule")
		}
		if path != modulePath {
			t.Errorf("Resolve() path = %q, want %q", path, modulePath)
		}
	})

	t.Run("Resolve standard module from arbitrary base dir", func(t *testing.T) {
		loader := NewModuleLoader(tmpDir)
		name, path, err := loader.Resolve("dw::core::Arrays")
		if err != nil {
			t.Errorf("Resolve() unexpected error: %v", err)
			return
		}
		if name != "Arrays" {
			t.Errorf("Resolve() name = %q, want %q", name, "Arrays")
		}
		if path != "dw::core::Arrays" {
			t.Errorf("Resolve() path = %q, want %q", path, "dw::core::Arrays")
		}
	})

	t.Run("Resolve not found", func(t *testing.T) {
		loader := NewModuleLoader(tmpDir)
		_, _, err := loader.Resolve("NonExistent")
		if err == nil {
			t.Errorf("Resolve() expected error for non-existent module")
		}
	})

	t.Run("Resolve rejects traversal", func(t *testing.T) {
		escapeDir := filepath.Join(filepath.Dir(tmpDir), "escape_mods")
		if err := os.MkdirAll(escapeDir, 0755); err != nil {
			t.Fatalf("Failed to create escape dir: %v", err)
		}
		t.Cleanup(func() { _ = os.RemoveAll(escapeDir) })

		escapePath := filepath.Join(escapeDir, "Evil.im")
		if err := os.WriteFile(escapePath, []byte("%im 0.1\nvar x = 42\n"), 0644); err != nil {
			t.Fatalf("Failed to write escape module file: %v", err)
		}

		loader := NewModuleLoader(tmpDir)
		_, _, err := loader.Resolve("..::escape_mods::Evil")
		if err == nil {
			t.Fatalf("Resolve() expected error for traversal module spec")
		}
		if !containsString(err.Error(), "invalid module spec") {
			t.Fatalf("Resolve() error = %q, expected invalid module spec error", err.Error())
		}
	})

	t.Run("Load", func(t *testing.T) {
		loader := NewModuleLoader(tmpDir)
		m, err := loader.Load("MyModule")
		if err != nil {
			t.Errorf("Load() unexpected error: %v", err)
			return
		}
		if m.Name != "MyModule" {
			t.Errorf("Load() module name = %q, want %q", m.Name, "MyModule")
		}
		if _, ok := m.Namespace["PI"]; !ok {
			t.Errorf("Load() module should have PI variable")
		}
		if _, ok := m.Namespace["double"]; !ok {
			t.Errorf("Load() module should have double function")
		}
	})

	t.Run("Load caching", func(t *testing.T) {
		loader := NewModuleLoader(tmpDir)
		m1, _ := loader.Load("MyModule")
		m2, _ := loader.Load("MyModule")
		if m1 != m2 {
			t.Errorf("Load() should return cached module on second call")
		}
	})

	t.Run("Load standard module from arbitrary base dir", func(t *testing.T) {
		loader := NewModuleLoader(tmpDir)
		m, err := loader.Load("dw::core::Arrays")
		if err != nil {
			t.Errorf("Load() unexpected error: %v", err)
			return
		}
		if m.Name != "Arrays" {
			t.Errorf("Load() module name = %q, want %q", m.Name, "Arrays")
		}
		if _, ok := m.Namespace["countBy"]; !ok {
			t.Errorf("Load() module should have countBy function")
		}
	})
}

func TestParseHeader(t *testing.T) {
	tests := []struct {
		name           string
		header         string
		hasHeader      bool
		wantMimeType   string
		wantContextKey string
		wantErr        bool
	}{
		{
			name:         "no header",
			header:       "",
			hasHeader:    false,
			wantMimeType: "application/json",
		},
		{
			name:         "output directive",
			header:       "output application/xml",
			hasHeader:    true,
			wantMimeType: "application/xml",
		},
		{
			name:           "variable declaration",
			header:         "var x = 42",
			hasHeader:      true,
			wantMimeType:   "application/json",
			wantContextKey: "x",
		},
		{
			name:           "function declaration",
			header:         "fun test() = 1",
			hasHeader:      true,
			wantMimeType:   "application/json",
			wantContextKey: "test",
		},
		{
			name:         "version declaration",
			header:       "%im 0.1\noutput application/json",
			hasHeader:    true,
			wantMimeType: "application/json",
		},
		{
			name:         "comment line",
			header:       "// note\noutput application/xml",
			hasHeader:    true,
			wantMimeType: "application/xml",
		},
		{
			name:         "input directive",
			header:       "input application/json\noutput application/json",
			hasHeader:    true,
			wantMimeType: "application/json",
		},
		{
			name:      "unknown directive",
			header:    "unknown directive",
			hasHeader: true,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewModuleLoader(".")
			scope, mimeType, _, err := parseHeader(tt.header, tt.hasHeader, tt.header, loader)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseHeader() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("parseHeader() unexpected error: %v", err)
				return
			}
			if mimeType != tt.wantMimeType {
				t.Errorf("parseHeader() mimeType = %q, want %q", mimeType, tt.wantMimeType)
			}
			if tt.wantContextKey != "" {
				if _, ok := scope.Vars[tt.wantContextKey]; !ok {
					t.Errorf("parseHeader() context should have key %q", tt.wantContextKey)
				}
			}
		})
	}
}

func TestParseHeaderDirectives(t *testing.T) {
	tests := []struct {
		name      string
		header    string
		hasHeader bool
		wantKinds []headerDirectiveKind
		wantErr   string
	}{
		{
			name:      "parse only keeps directive order without evaluating declarations",
			hasHeader: true,
			header: `%im 0.1
var broken = (
fun wrap(x) = x
output application/json`,
			wantKinds: []headerDirectiveKind{
				headerDirectiveVersion,
				headerDirectiveVar,
				headerDirectiveFun,
				headerDirectiveOutput,
			},
		},
		{
			name:      "invalid variable declaration still fails during parse phase",
			hasHeader: true,
			header: `%im 0.1
var broken
output application/json`,
			wantErr: "missing '='",
		},
		{
			name:      "invalid output option still fails during parse phase",
			hasHeader: true,
			header: `%im 0.1
output application/json badoption`,
			wantErr: "invalid output option",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			directives, err := parseHeaderDirectives(tt.header, tt.hasHeader, tt.header)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("parseHeaderDirectives() expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("parseHeaderDirectives() error = %v, want substring %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseHeaderDirectives() unexpected error: %v", err)
			}
			if len(directives) != len(tt.wantKinds) {
				t.Fatalf("parseHeaderDirectives() directive count = %d, want %d", len(directives), len(tt.wantKinds))
			}
			for i, kind := range tt.wantKinds {
				if directives[i].Kind != kind {
					t.Fatalf("parseHeaderDirectives()[%d].Kind = %q, want %q", i, directives[i].Kind, kind)
				}
			}
		})
	}
}

func TestParseHeaderDirectivesBuildsDeclarationIRWithoutEvaluation(t *testing.T) {
	header := `%im 0.1
var broken = missing + 1
fun invalid(x) = x +
type Label = String { format: "plain" }
output application/json`

	declarations, err := parseHeaderDirectives(header, true, header)
	if err != nil {
		t.Fatalf("parseHeaderDirectives() unexpected error: %v", err)
	}
	if len(declarations) != 5 {
		t.Fatalf("parseHeaderDirectives() declaration count = %d, want 5", len(declarations))
	}

	varDecl := declarations[1]
	if varDecl.Kind != DeclarationVar || varDecl.Var == nil {
		t.Fatalf("declarations[1] = %#v, want var declaration", varDecl)
	}
	if varDecl.Var.Name != "broken" || varDecl.Var.Expression != "missing + 1" {
		t.Fatalf("var declaration = %#v, want name broken and unevaluated expression", varDecl.Var)
	}
	if varDecl.Source.Span.Start != varDecl.Source.Offset || varDecl.Source.Span.End <= varDecl.Source.Span.Start {
		t.Fatalf("var declaration source span = %#v, offset %d", varDecl.Source.Span, varDecl.Source.Offset)
	}

	funDecl := declarations[2]
	if funDecl.Kind != DeclarationFun || funDecl.Function == nil {
		t.Fatalf("declarations[2] = %#v, want function declaration", funDecl)
	}
	if funDecl.Function.Name != "invalid" || funDecl.Function.Body != "x +" {
		t.Fatalf("function declaration = %#v, want invalid body preserved without AST parsing", funDecl.Function)
	}

	typeDecl := declarations[3]
	if typeDecl.Kind != DeclarationType || typeDecl.Type == nil {
		t.Fatalf("declarations[3] = %#v, want type declaration", typeDecl)
	}
	if typeDecl.Type.Name != "Label" || typeDecl.Type.BaseType != "String" {
		t.Fatalf("type declaration = %#v, want Label = String", typeDecl.Type)
	}
}

func TestScriptHeadersAndModulesShareDeclarationIR(t *testing.T) {
	source := `%im 0.1
import double from modules::MathUtils
var base = 5
fun scale(x) = double(x) + base
type Label = String`

	headerDeclarations, err := parseHeaderDirectives(source, true, source)
	if err != nil {
		t.Fatalf("parseHeaderDirectives() unexpected error: %v", err)
	}
	moduleDeclarations, err := parseModuleDirectives(source)
	if err != nil {
		t.Fatalf("parseModuleDirectives() unexpected error: %v", err)
	}
	if len(headerDeclarations) != len(moduleDeclarations) {
		t.Fatalf("declaration count mismatch: header=%d module=%d", len(headerDeclarations), len(moduleDeclarations))
	}

	for i := range headerDeclarations {
		if headerDeclarations[i].Kind != moduleDeclarations[i].Kind {
			t.Fatalf("declaration %d kind mismatch: header=%q module=%q", i, headerDeclarations[i].Kind, moduleDeclarations[i].Kind)
		}
	}
	if headerDeclarations[1].Import.ModuleSpec != moduleDeclarations[1].Import.ModuleSpec {
		t.Fatalf("import module mismatch: header=%#v module=%#v", headerDeclarations[1].Import, moduleDeclarations[1].Import)
	}
	if headerDeclarations[2].Var.Name != moduleDeclarations[2].Var.Name {
		t.Fatalf("var mismatch: header=%#v module=%#v", headerDeclarations[2].Var, moduleDeclarations[2].Var)
	}
	if headerDeclarations[3].Function.Name != moduleDeclarations[3].Function.Name {
		t.Fatalf("function mismatch: header=%#v module=%#v", headerDeclarations[3].Function, moduleDeclarations[3].Function)
	}
	if headerDeclarations[4].Type.Name != moduleDeclarations[4].Type.Name {
		t.Fatalf("type mismatch: header=%#v module=%#v", headerDeclarations[4].Type, moduleDeclarations[4].Type)
	}
}

func TestParseModuleContentRejectsBodySection(t *testing.T) {
	_, err := parseModuleContent("%im 0.1\nvar x = 1\n---\nx", NewModuleLoader("."))
	if err == nil {
		t.Fatal("parseModuleContent() expected error for module body separator, got nil")
	}
	for _, want := range []string{"modules cannot contain a body section", "line: ---"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("parseModuleContent() error = %v, want substring %q", err, want)
		}
	}
}

func TestSplitPropertyPairs(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "single pair",
			input: `format: "##.00"`,
			want:  []string{`format: "##.00"`},
		},
		{
			name:  "multiple pairs",
			input: `a: 1, b: 2`,
			want:  []string{"a: 1", " b: 2"},
		},
		{
			name:  "comma in string",
			input: `format: "a, b", other: 1`,
			want:  []string{`format: "a, b"`, ` other: 1`},
		},
		{
			name:  "empty",
			input: "",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitPropertyPairs(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("splitPropertyPairs() = %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("splitPropertyPairs()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestParsePropertyValue(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    evaluator.Value
		wantErr bool
	}{
		{name: "string", input: `"hello"`, want: "hello"},
		{name: "integer", input: "42", want: 42},
		{name: "float", input: "3.14", want: 3.14},
		{name: "true", input: "true", want: true},
		{name: "false", input: "false", want: false},
		{name: "null", input: "null", want: nil},
		{name: "invalid", input: "not_valid", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePropertyValue(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parsePropertyValue() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("parsePropertyValue() unexpected error: %v", err)
				return
			}
			if !deepEqual(got, tt.want) {
				t.Errorf("parsePropertyValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeHeader(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   string
	}{
		{
			name:   "already multi-line",
			header: "%im 0.1\noutput application/json",
			want:   "%im 0.1\noutput application/json",
		},
		{
			name:   "single-line with version and output",
			header: "%im 0.1 output application/json",
			want:   "%im 0.1\noutput application/json",
		},
		{
			name:   "single-line with var",
			header: `%im 0.1 var x = 10 output application/json`,
			want:   "%im 0.1\nvar x = 10\noutput application/json",
		},
		{
			name:   "single-line with var containing object",
			header: `%im 0.1 var myObject = { user : "a" } output application/json`,
			want:   "%im 0.1\nvar myObject = { user : \"a\" }\noutput application/json",
		},
		{
			name:   "single-line with function",
			header: `%im 0.1 fun double(x) = x * 2 output application/json`,
			want:   "%im 0.1\nfun double(x) = x * 2\noutput application/json",
		},
		{
			name:   "keyword in string should not split",
			header: `%im 0.1 var msg = "output this" output application/json`,
			want:   "%im 0.1\nvar msg = \"output this\"\noutput application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeHeader(tt.header)
			if got != tt.want {
				t.Errorf("normalizeHeader() = %q, want %q", got, tt.want)
			}
		})
	}
}

// Helper functions

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func deepEqual(a, b evaluator.Value) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	switch va := a.(type) {
	case evaluator.Array:
		vb, ok := b.(evaluator.Array)
		if !ok || len(va) != len(vb) {
			return false
		}
		for i := range va {
			if !deepEqual(va[i], vb[i]) {
				return false
			}
		}
		return true
	case evaluator.Object:
		vb, ok := b.(evaluator.Object)
		if !ok || len(va) != len(vb) {
			return false
		}
		for k, v := range va {
			if !deepEqual(v, vb[k]) {
				return false
			}
		}
		return true
	case int:
		switch vb := b.(type) {
		case int:
			return va == vb
		case float64:
			return float64(va) == vb
		}
		return false
	case float64:
		switch vb := b.(type) {
		case float64:
			return va == vb
		case int:
			return va == float64(vb)
		}
		return false
	default:
		return a == b
	}
}
