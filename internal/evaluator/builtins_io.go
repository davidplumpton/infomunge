package evaluator

import (
	"fmt"
	"go/ast"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"infomunge/pkg/formats"
)

const (
	// readUrlTimeout is the maximum time allowed for a readUrl HTTP request.
	readUrlTimeout = 30 * time.Second
	// maxResponseBytes is the maximum response body size readUrl will consume (10 MB).
	maxResponseBytes = 10 * 1024 * 1024
)

// readUrlClient is a shared HTTP client with timeout for readUrl calls.
var readUrlClient = &http.Client{
	Timeout: readUrlTimeout,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return fmt.Errorf("too many redirects")
		}
		// Validate each redirect target against SSRF rules
		if err := validateURL(req.URL); err != nil {
			return err
		}
		return nil
	},
}

// validateURL checks that a URL is safe to fetch (not an internal/private address).
func validateURL(u *url.URL) error {
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("readUrl: unsupported scheme %q (only http and https are allowed)", u.Scheme)
	}

	hostname := u.Hostname()
	if hostname == "" {
		return fmt.Errorf("readUrl: empty hostname")
	}

	// Block well-known metadata hostnames
	if hostname == "metadata.google.internal" {
		return fmt.Errorf("readUrl: access to %q is blocked", hostname)
	}

	ip := net.ParseIP(hostname)
	if ip == nil {
		// It's a hostname — resolve it to check for private IPs
		addrs, err := net.LookupHost(hostname)
		if err != nil {
			// Let the HTTP client handle DNS errors naturally
			return nil
		}
		for _, addr := range addrs {
			if resolved := net.ParseIP(addr); resolved != nil && isPrivateIP(resolved) {
				return fmt.Errorf("readUrl: hostname %q resolves to private address %s", hostname, addr)
			}
		}
		return nil
	}

	if isPrivateIP(ip) {
		return fmt.Errorf("readUrl: access to private/internal address %s is blocked", ip)
	}
	return nil
}

// isPrivateIP returns true for loopback, private, link-local, and metadata addresses.
func isPrivateIP(ip net.IP) bool {
	return ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified()
}

// callBuiltinRead implements the read(content, mimeType) function.
func callBuiltinRead(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) < 2 {
		return nil, newPosError("read function requires at least 2 arguments: content and mimeType", e.Pos())
	}
	content, contentIsString := args[0].(string)
	mimeType, mimeTypeIsString := args[1].(string)
	if !contentIsString || !mimeTypeIsString {
		return nil, newPosError("read function arguments must be strings", e.Pos())
	}
	res, err := formats.Read(content, mimeType)
	if err != nil {
		return nil, newPosError(err.Error(), e.Pos())
	}
	return res, nil
}

// callBuiltinReadUrl implements the readUrl(url, mimeType) function.
func callBuiltinReadUrl(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) != 2 {
		return nil, newPosError("readUrl requires exactly 2 arguments: url and mimeType", e.Pos())
	}

	rawURL, ok := args[0].(string)
	if !ok {
		return nil, newPosError(fmt.Sprintf("readUrl expects url to be a string, got %T", args[0]), e.Pos())
	}

	mimeType, ok := args[1].(string)
	if !ok {
		return nil, newPosError(fmt.Sprintf("readUrl expects mimeType to be a string, got %T", args[1]), e.Pos())
	}

	// Validate URL before fetching
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("readUrl: invalid URL: %v", err), e.Pos())
	}
	if err := validateURL(parsed); err != nil {
		return nil, newPosError(err.Error(), e.Pos())
	}

	// Fetch the URL with timeout-aware client
	resp, err := readUrlClient.Get(rawURL)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("readUrl: failed to fetch URL: %v", err), e.Pos())
	}
	defer resp.Body.Close()

	// Read the response body with size limit
	limited := io.LimitReader(resp.Body, maxResponseBytes+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("readUrl: failed to read response: %v", err), e.Pos())
	}
	if int64(len(body)) > maxResponseBytes {
		return nil, newPosError(fmt.Sprintf("readUrl: response exceeds maximum size of %d bytes", maxResponseBytes), e.Pos())
	}

	// Parse the content
	result, err := formats.Read(string(body), mimeType)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("readUrl: failed to parse content: %v", err), e.Pos())
	}

	return result, nil
}

// callBuiltinWrite implements the write(value, mimeType) function.
func callBuiltinWrite(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) != 2 {
		return nil, newPosError("write requires exactly 2 arguments: value and mimeType", e.Pos())
	}

	mimeType, ok := args[1].(string)
	if !ok {
		return nil, newPosError(fmt.Sprintf("write expects mimeType to be a string, got %T", args[1]), e.Pos())
	}

	result, err := formats.Format(args[0], mimeType)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("write error: %v", err), e.Pos())
	}

	return result, nil
}
