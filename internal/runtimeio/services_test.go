package runtimeio

import (
	"context"
	"errors"
	"net"
	"strings"
	"testing"
)

type resolverFunc func(context.Context, string) ([]net.IPAddr, error)

func (f resolverFunc) LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error) {
	return f(ctx, host)
}

func TestSafeDialerBlocksPrivateResolvedAddress(t *testing.T) {
	dialed := false
	dialer := safeDialer{
		resolver: resolverFunc(func(context.Context, string) ([]net.IPAddr, error) {
			return []net.IPAddr{{IP: net.ParseIP("127.0.0.1")}}, nil
		}),
		dialContext: func(context.Context, string, string) (net.Conn, error) {
			dialed = true
			return nil, errors.New("unexpected dial")
		},
	}

	_, err := dialer.DialContext(context.Background(), "tcp", "rebind.example:80")
	if err == nil || !strings.Contains(err.Error(), "resolves to private address") {
		t.Fatalf("expected private-resolution error, got %v", err)
	}
	if dialed {
		t.Fatal("dialer should block before dialing private resolved addresses")
	}
}

func TestSafeDialerFailsClosedOnResolutionError(t *testing.T) {
	dialed := false
	dialer := safeDialer{
		resolver: resolverFunc(func(context.Context, string) ([]net.IPAddr, error) {
			return nil, errors.New("dns unavailable")
		}),
		dialContext: func(context.Context, string, string) (net.Conn, error) {
			dialed = true
			return nil, errors.New("unexpected dial")
		},
	}

	_, err := dialer.DialContext(context.Background(), "tcp", "missing.example:80")
	if err == nil || !strings.Contains(err.Error(), "failed to resolve hostname") {
		t.Fatalf("expected resolution failure, got %v", err)
	}
	if dialed {
		t.Fatal("dialer should not dial after a resolution error")
	}
}

func TestSafeDialerBlocksPrivateRemoteAddress(t *testing.T) {
	localConn, remoteConn := net.Pipe()
	defer remoteConn.Close()

	dialer := safeDialer{
		resolver: resolverFunc(func(context.Context, string) ([]net.IPAddr, error) {
			return []net.IPAddr{{IP: net.ParseIP("93.184.216.34")}}, nil
		}),
		dialContext: func(context.Context, string, string) (net.Conn, error) {
			return remoteAddrConn{
				Conn: localConn,
				addr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080},
			}, nil
		},
	}

	_, err := dialer.DialContext(context.Background(), "tcp", "public.example:80")
	if err == nil || !strings.Contains(err.Error(), "connected to private/internal address") {
		t.Fatalf("expected private remote-address error, got %v", err)
	}
}

type remoteAddrConn struct {
	net.Conn
	addr net.Addr
}

func (c remoteAddrConn) RemoteAddr() net.Addr {
	return c.addr
}
