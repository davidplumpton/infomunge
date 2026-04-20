package cli

import (
	"flag"
	"fmt"
	unifiederrors "infomunge/internal/errors"
	"infomunge/internal/evaluator"
	input "infomunge/internal/io"
	"infomunge/internal/runner"
	"infomunge/internal/version"
	"os"
	"path/filepath"
	"strings"
)

const defaultListenAddr = ":8080"

// Config holds the application configuration
type Config struct {
	ScriptFile   string
	StdinFormat  string
	Inputs       []string
	Script       string
	Lazy         bool
	Server       bool
	Listen       string
	ServerAPIKey string
	ShowVersion  bool
}

// App represents the CLI application
type App struct {
	parser *input.Parser
}

// NewApp creates a new CLI application
func NewApp() *App {
	return &App{
		parser: input.NewParser(),
	}
}

// Run executes the CLI application
func (app *App) Run(args []string) error {
	// Allow optional "run" subcommand for DataWeave CLI compatibility
	if len(args) > 0 && args[0] == "run" {
		args = args[1:]
	}

	config, err := app.parseFlags(args)
	if err != nil {
		return err
	}

	if config.ShowVersion {
		fmt.Println(version.String())
		return nil
	}

	if config.Server {
		if config.ScriptFile != "" || config.Script != "" || config.StdinFormat != "" || len(config.Inputs) > 0 {
			return unifiederrors.ValidationError("server mode does not accept scripts or inputs")
		}
		return app.serve(config)
	}

	context, err := app.prepareInputs(config)
	if err != nil {
		return err
	}

	return app.execute(config, context)
}

// parseFlags parses command line flags and arguments
func (app *App) parseFlags(args []string) (*Config, error) {
	var inputs inputFlags

	fs := flag.NewFlagSet("infomunge", flag.ExitOnError)

	scriptFile := fs.String("f", "", "script file to execute")
	stdinFormat := fs.String("s", "", "read stdin as format (e.g., json), shorthand for -i payload=:format for stdin")
	lazy := fs.Bool("lazy", false, "deprecated: lazy mode flag is unsupported (use lazy builtins directly)")
	serverMode := fs.Bool("server", false, "run as an HTTP server")
	listenAddr := fs.String("listen", defaultListenAddr, "address to listen on in server mode")
	serverAPIKey := fs.String("api-key", "", "shared API key required for /run requests in server mode")
	showVersion := fs.Bool("version", false, "show version information")
	fs.Var(&inputs, "i", "input source (name=file or name=:format for stdin)")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: infomunge [-s format] [-i name=source]... [-f <script_file>] [script]\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	// Auto-detect piped stdin and default to JSON if no -s flag (non-server mode only)
	processedInputs := inputs
	if !*serverMode {
		if *stdinFormat != "" {
			processedInputs = append([]string{"payload=:" + *stdinFormat}, inputs...)
		} else if input.StdinIsPiped() {
			processedInputs = append([]string{"payload=:json"}, inputs...)
		}
	}

	var script string
	if fs.NArg() > 0 {
		script = fs.Arg(0)
	}

	return &Config{
		ScriptFile:   *scriptFile,
		StdinFormat:  *stdinFormat,
		Inputs:       processedInputs,
		Script:       script,
		Lazy:         *lazy,
		Server:       *serverMode,
		Listen:       *listenAddr,
		ServerAPIKey: *serverAPIKey,
		ShowVersion:  *showVersion,
	}, nil
}

// prepareInputs processes all input specifications
func (app *App) prepareInputs(config *Config) (evaluator.Context, error) {
	return app.parser.ParseInputs(config.Inputs)
}

// execute runs the script with the given context
func (app *App) execute(config *Config, context evaluator.Context) error {
	opts := runner.RunnerOptions{
		Lazy: config.Lazy,
	}

	if config.ScriptFile != "" {
		content, err := os.ReadFile(config.ScriptFile)
		if err != nil {
			return unifiederrors.WrapIOf(err, "error reading script file: %s", config.ScriptFile)
		}
		absPath, err := filepath.Abs(config.ScriptFile)
		if err != nil {
			opts.BaseDir = filepath.Dir(config.ScriptFile)
		} else {
			opts.BaseDir = filepath.Dir(absPath)
		}
		return runner.RunFromStringWithContextAndOptions(string(content), context, opts)
	}

	if config.Script != "" {
		return runner.RunFromStringWithContextAndOptions(config.Script, context, opts)
	}

	return unifiederrors.ValidationError("no script provided")
}

// inputFlags collects multiple -i flag values
type inputFlags []string

func (i *inputFlags) String() string {
	return strings.Join(*i, ", ")
}

func (i *inputFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}
