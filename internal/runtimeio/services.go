package runtimeio

import (
	"context"
	"errors"
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
		Timeout:   ReadURLTimeout,
		Transport: safeHTTPTransport(),
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

// ValidateURL checks URL shape and literal hosts before fetching. Hostname
// resolution and actual remote-address enforcement happen in the HTTP dialer.
func ValidateURL(u *url.URL) error {
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("readUrl: unsupported scheme %q (only http and https are allowed)", u.Scheme)
	}

	hostname := u.Hostname()
	if hostname == "" {
		return fmt.Errorf("readUrl: empty hostname")
	}

	if isBlockedHostname(hostname) {
		return fmt.Errorf("readUrl: access to %q is blocked", hostname)
	}

	if ip := parseHostIP(hostname); ip != nil && IsPrivateIP(ip) {
		return fmt.Errorf("readUrl: access to private/internal address %s is blocked", ip)
	}
	return nil
}

type ipResolver interface {
	LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error)
}

type dialContextFunc func(context.Context, string, string) (net.Conn, error)

type safeDialer struct {
	resolver    ipResolver
	dialContext dialContextFunc
}

func safeHTTPTransport() *http.Transport {
	base, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		base = &http.Transport{}
	}
	transport := base.Clone()
	transport.Proxy = nil
	transport.DialContext = newSafeDialer().DialContext
	return transport
}

func newSafeDialer() safeDialer {
	dialer := &net.Dialer{}
	return safeDialer{
		resolver:    net.DefaultResolver,
		dialContext: dialer.DialContext,
	}
}

func (d safeDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("readUrl: invalid dial address %q: %w", address, err)
	}
	if host == "" {
		return nil, fmt.Errorf("readUrl: empty hostname")
	}
	if isBlockedHostname(host) {
		return nil, fmt.Errorf("readUrl: access to %q is blocked", host)
	}

	addrs, err := d.resolve(ctx, host)
	if err != nil {
		return nil, err
	}
	for _, addr := range addrs {
		if addr.IP == nil {
			return nil, fmt.Errorf("readUrl: hostname %q resolved to an unverifiable address", host)
		}
		if IsPrivateIP(addr.IP) {
			return nil, fmt.Errorf("readUrl: hostname %q resolves to private address %s", host, addr.IP)
		}
	}

	dial := d.dialContext
	if dial == nil {
		dial = (&net.Dialer{}).DialContext
	}

	var lastErr error
	for _, addr := range addrs {
		conn, err := dial(ctx, network, net.JoinHostPort(addr.IP.String(), port))
		if err != nil {
			lastErr = err
			continue
		}
		if err := validateRemoteAddress(host, conn.RemoteAddr()); err != nil {
			_ = conn.Close()
			return nil, err
		}
		return conn, nil
	}
	if lastErr != nil {
		return nil, fmt.Errorf("readUrl: failed to connect to hostname %q: %w", host, lastErr)
	}
	return nil, fmt.Errorf("readUrl: hostname %q resolved to no addresses", host)
}

func (d safeDialer) resolve(ctx context.Context, host string) ([]net.IPAddr, error) {
	if ip := parseHostIP(host); ip != nil {
		return []net.IPAddr{{IP: ip}}, nil
	}
	resolver := d.resolver
	if resolver == nil {
		resolver = net.DefaultResolver
	}
	addrs, err := resolver.LookupIPAddr(ctx, host)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		return nil, fmt.Errorf("readUrl: failed to resolve hostname %q: %w", host, err)
	}
	if len(addrs) == 0 {
		return nil, fmt.Errorf("readUrl: hostname %q resolved to no addresses", host)
	}
	return addrs, nil
}

func validateRemoteAddress(host string, addr net.Addr) error {
	ip := remoteIP(addr)
	if ip == nil {
		return fmt.Errorf("readUrl: could not verify remote address %q for hostname %q", addr, host)
	}
	if IsPrivateIP(ip) {
		return fmt.Errorf("readUrl: connected to private/internal address %s for hostname %q", ip, host)
	}
	return nil
}

func remoteIP(addr net.Addr) net.IP {
	switch typed := addr.(type) {
	case *net.TCPAddr:
		return typed.IP
	case *net.UDPAddr:
		return typed.IP
	case *net.IPAddr:
		return typed.IP
	}
	if addr == nil {
		return nil
	}
	host, _, err := net.SplitHostPort(addr.String())
	if err != nil {
		host = addr.String()
	}
	return parseHostIP(host)
}

func isBlockedHostname(hostname string) bool {
	normalized := strings.TrimSuffix(strings.ToLower(hostname), ".")
	return normalized == "metadata.google.internal"
}

func parseHostIP(hostname string) net.IP {
	if ip := net.ParseIP(hostname); ip != nil {
		return ip
	}
	if zoneIndex := strings.LastIndex(hostname, "%"); zoneIndex > 0 {
		return net.ParseIP(hostname[:zoneIndex])
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
