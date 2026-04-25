package runtimeio

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"infomunge/internal/evaluator"
	"infomunge/pkg/formats"
)

const (
	// ReadURLTimeout is the maximum time allowed for a readUrl HTTP request.
	ReadURLTimeout = 30 * time.Second
	// MaxResponseBytes is the maximum response body size readUrl will consume.
	MaxResponseBytes = 10 * 1024 * 1024
)

// FormatService adapts pkg/formats to the evaluator's explicit format boundary.
type FormatService struct{}

func (FormatService) Read(content, mimeType string) (evaluator.Value, error) {
	return formats.Read(content, mimeType)
}

func (FormatService) ReadWithOptions(content, mimeType string, options evaluator.Object) (evaluator.Value, error) {
	return formats.ReadWithOptions(content, mimeType, options)
}

func (FormatService) Write(value evaluator.Value, mimeType string) (string, error) {
	return formats.Format(value, mimeType)
}

func (FormatService) WriteWithOptions(value evaluator.Value, mimeType string, options evaluator.Object) (string, error) {
	return formats.FormatWithOptions(value, mimeType, options)
}

// URLReadService adapts HTTP fetching to the evaluator's readUrl capability.
type URLReadService struct {
	Client        *http.Client
	FormatService evaluator.FormatService
}

func NewURLReadService(formatService evaluator.FormatService) *URLReadService {
	if formatService == nil {
		formatService = FormatService{}
	}
	return &URLReadService{
		Client:        defaultHTTPClient(),
		FormatService: formatService,
	}
}

func defaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout: ReadURLTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return ValidateURL(req.URL)
		},
	}
}

func (s *URLReadService) ReadURL(ctx context.Context, rawURL, mimeType string) (evaluator.Value, error) {
	if s == nil {
		return nil, fmt.Errorf("readUrl: URL IO capability is unavailable")
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("readUrl: invalid URL: %v", err)
	}
	if err := ValidateURL(parsed); err != nil {
		return nil, err
	}

	client := s.Client
	if client == nil {
		client = defaultHTTPClient()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("readUrl: failed to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("readUrl: failed to fetch URL: %v", err)
	}
	defer resp.Body.Close()

	limited := io.LimitReader(resp.Body, MaxResponseBytes+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("readUrl: failed to read response: %v", err)
	}
	if int64(len(body)) > MaxResponseBytes {
		return nil, fmt.Errorf("readUrl: response exceeds maximum size of %d bytes", MaxResponseBytes)
	}

	formatService := s.FormatService
	if formatService == nil {
		formatService = FormatService{}
	}
	result, err := formatService.Read(string(body), mimeType)
	if err != nil {
		return nil, fmt.Errorf("readUrl: failed to parse content: %v", err)
	}
	return result, nil
}

// ValidateURL checks that a URL is safe to fetch.
func ValidateURL(u *url.URL) error {
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("readUrl: unsupported scheme %q (only http and https are allowed)", u.Scheme)
	}

	hostname := u.Hostname()
	if hostname == "" {
		return fmt.Errorf("readUrl: empty hostname")
	}

	if hostname == "metadata.google.internal" {
		return fmt.Errorf("readUrl: access to %q is blocked", hostname)
	}

	ip := net.ParseIP(hostname)
	if ip == nil {
		addrs, err := net.LookupHost(hostname)
		if err != nil {
			return nil
		}
		for _, addr := range addrs {
			if resolved := net.ParseIP(addr); resolved != nil && IsPrivateIP(resolved) {
				return fmt.Errorf("readUrl: hostname %q resolves to private address %s", hostname, addr)
			}
		}
		return nil
	}

	if IsPrivateIP(ip) {
		return fmt.Errorf("readUrl: access to private/internal address %s is blocked", ip)
	}
	return nil
}

// IsPrivateIP returns true for loopback, private, link-local, and unspecified addresses.
func IsPrivateIP(ip net.IP) bool {
	return ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified()
}
