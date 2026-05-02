package input

import (
	"os"
	"strings"

	unifiederrors "infomunge/internal/errors"
	"infomunge/internal/evaluator"
	"infomunge/internal/readlimit"
	"infomunge/pkg/formats"
)

const (
	MaxStdinBytes     = 10 * 1024 * 1024
	MaxFileInputBytes = 10 * 1024 * 1024
)

// Source represents a data source with its content and MIME type
type Source struct {
	Content  []byte
	MimeType string
}

// Parser handles parsing input sources from various origins
type Parser struct{}

// NewParser creates a new input parser
func NewParser() *Parser {
	return &Parser{}
}

// ParseSource reads content from a source (file path or :format for stdin)
func (p *Parser) ParseSource(source string) (*Source, error) {
	if strings.HasPrefix(source, ":") {
		return p.parseStdin(source[1:])
	}
	return p.parseFile(source)
}

// parseStdin reads from stdin with expected format
func (p *Parser) parseStdin(format string) (*Source, error) {
	mimeType := formats.MimeTypeForFormat(format)
	if mimeType == "" {
		return nil, unifiederrors.ValidationErrorf("unknown format: %s", format)
	}

	content, tooLarge, err := readlimit.ReadAll(os.Stdin, MaxStdinBytes)
	if err != nil {
		return nil, unifiederrors.WrapIOf(err, "reading stdin")
	}
	if tooLarge {
		return nil, unifiederrors.ValidationErrorf("stdin input exceeds maximum size of %d bytes", MaxStdinBytes)
	}

	return &Source{
		Content:  content,
		MimeType: mimeType,
	}, nil
}

// parseFile reads from a file path
func (p *Parser) parseFile(filePath string) (*Source, error) {
	mimeType := formats.DetectMimeType(filePath)
	if mimeType == "" {
		return nil, unifiederrors.ValidationErrorf("unsupported file extension: %s", filePath)
	}

	content, tooLarge, err := readlimit.ReadFile(filePath, MaxFileInputBytes)
	if err != nil {
		return nil, unifiederrors.WrapIOf(err, "reading file %s", filePath)
	}
	if tooLarge {
		return nil, unifiederrors.ValidationErrorf("input file %s exceeds maximum size of %d bytes", filePath, MaxFileInputBytes)
	}

	return &Source{
		Content:  content,
		MimeType: mimeType,
	}, nil
}

// ParseInputs processes multiple input specifications into a context map
func (p *Parser) ParseInputs(inputs []string) (evaluator.Context, error) {
	context := make(evaluator.Context)
	stdinSources := 0

	for _, input := range inputs {
		parts := strings.SplitN(input, "=", 2)
		if len(parts) != 2 {
			return nil, unifiederrors.ValidationErrorf("invalid input format: %s (expected name=source)", input)
		}
		if strings.HasPrefix(parts[1], ":") {
			stdinSources++
		}
	}

	if stdinSources > 1 {
		return nil, unifiederrors.ValidationError("multiple stdin-backed inputs are not supported")
	}

	for _, input := range inputs {
		parts := strings.SplitN(input, "=", 2)
		if len(parts) != 2 {
			return nil, unifiederrors.ValidationErrorf("invalid input format: %s (expected name=source)", input)
		}

		name, err := NormalizeAndValidateInputName(parts[0])
		if err != nil {
			return nil, err
		}
		source := parts[1]
		if _, exists := context[name]; exists {
			return nil, unifiederrors.ValidationErrorf("duplicate input name %q", name)
		}

		src, err := p.ParseSource(source)
		if err != nil {
			return nil, err
		}

		data, err := formats.Read(string(src.Content), src.MimeType)
		if err != nil {
			return nil, unifiederrors.WrapParsef(err, "parsing %s", source)
		}

		context[name] = data
	}

	return context, nil
}

// StdinIsPiped returns true if stdin is piped (not a terminal)
func StdinIsPiped() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) == 0
}
