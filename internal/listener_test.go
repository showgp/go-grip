package internal

import (
	"fmt"
	"net"
	"strings"
	"testing"
)

func TestListenOnPortFallsBackWhenNotStrict(t *testing.T) {
	t.Parallel()

	occupied, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("reserve port: %v", err)
	}
	defer func() {
		if err := occupied.Close(); err != nil {
			t.Errorf("close occupied listener: %v", err)
		}
	}()

	port := occupied.Addr().(*net.TCPAddr).Port
	listener, actualPort, err := listenOn("", port, false)
	if err != nil {
		t.Fatalf("listen with fallback: %v", err)
	}
	defer func() {
		if err := listener.Close(); err != nil {
			t.Errorf("close fallback listener: %v", err)
		}
	}()

	if actualPort == port {
		t.Fatalf("expected fallback port, got original occupied port %d", actualPort)
	}
}

func TestListenOnPortReportsStrictConflict(t *testing.T) {
	t.Parallel()

	occupied, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("reserve port: %v", err)
	}
	defer func() {
		if err := occupied.Close(); err != nil {
			t.Errorf("close occupied listener: %v", err)
		}
	}()

	port := occupied.Addr().(*net.TCPAddr).Port
	listener, _, err := listenOn("", port, true)
	if err == nil {
		if err := listener.Close(); err != nil {
			t.Errorf("close unexpected listener: %v", err)
		}
		t.Fatalf("expected strict conflict on port %d", port)
	}
	if got := err.Error(); got == "" || !strings.Contains(got, fmt.Sprintf("%d", port)) {
		t.Fatalf("expected error to mention port %d, got %q", port, got)
	}
}

// TestListenOnBindsLoopbackOnly pins the managed preview contract: with an
// explicit loopback bind the listener owns the real loopback address and an
// OS-assigned port, not a wildcard socket.
func TestListenOnBindsLoopbackOnly(t *testing.T) {
	t.Parallel()

	listener, port, err := listenOn("127.0.0.1", 0, true)
	if err != nil {
		t.Fatalf("listen on loopback: %v", err)
	}
	defer func() {
		if err := listener.Close(); err != nil {
			t.Errorf("close listener: %v", err)
		}
	}()

	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("expected TCP address, got %T", listener.Addr())
	}
	if !addr.IP.IsLoopback() {
		t.Fatalf("expected loopback bind, got %s", addr)
	}
	if port <= 0 || port != addr.Port {
		t.Fatalf("expected the OS-assigned port %d to match %d", addr.Port, port)
	}

	conn, err := net.Dial("tcp", addr.String())
	if err != nil {
		t.Fatalf("dial the loopback listener: %v", err)
	}
	if err := conn.Close(); err != nil {
		t.Errorf("close dialed connection: %v", err)
	}
}
