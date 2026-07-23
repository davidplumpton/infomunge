package cli

import (
	"strings"
	"testing"
)

func TestEffectiveListenAddrDefaultsToLoopback(t *testing.T) {
	if got := effectiveListenAddr(""); got != "127.0.0.1:8080" {
		t.Fatalf("effectiveListenAddr(\"\") = %q, want loopback default", got)
	}
}

func TestValidateServerExposure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		addr    string
		apiKey  string
		wantErr string
	}{
		{name: "IPv4 loopback", addr: "127.0.0.1:8080"},
		{name: "IPv4 loopback range", addr: "127.12.34.56:8080"},
		{name: "IPv6 loopback", addr: "[::1]:8080"},
		{name: "localhost", addr: "localhost:8080"},
		{
			name:    "wildcard host",
			addr:    ":8080",
			wantErr: "without --api-key must listen on a loopback address",
		},
		{
			name:    "IPv4 wildcard",
			addr:    "0.0.0.0:8080",
			wantErr: "without --api-key must listen on a loopback address",
		},
		{
			name:    "IPv6 wildcard",
			addr:    "[::]:8080",
			wantErr: "without --api-key must listen on a loopback address",
		},
		{
			name:    "LAN address",
			addr:    "192.168.1.20:8080",
			wantErr: "without --api-key must listen on a loopback address",
		},
		{
			name:    "non-loopback hostname",
			addr:    "example.com:8080",
			wantErr: "without --api-key must listen on a loopback address",
		},
		{
			name:    "missing port",
			addr:    "127.0.0.1",
			wantErr: "invalid server listen address",
		},
		{
			name:   "API key permits network exposure",
			addr:   "0.0.0.0:8080",
			apiKey: "secret-token",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := validateServerExposure(test.addr, test.apiKey)
			if test.wantErr == "" {
				if err != nil {
					t.Fatalf("validateServerExposure(%q, key present=%t) error = %v", test.addr, test.apiKey != "", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("validateServerExposure(%q, key present=%t) succeeded, want error containing %q", test.addr, test.apiKey != "", test.wantErr)
			}
			if !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("validateServerExposure(%q) error = %q, want substring %q", test.addr, err, test.wantErr)
			}
		})
	}
}
