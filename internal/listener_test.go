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
	defer occupied.Close()

	port := occupied.Addr().(*net.TCPAddr).Port
	listener, actualPort, err := listenOnPort(port, false)
	if err != nil {
		t.Fatalf("listen with fallback: %v", err)
	}
	defer listener.Close()

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
	defer occupied.Close()

	port := occupied.Addr().(*net.TCPAddr).Port
	listener, _, err := listenOnPort(port, true)
	if err == nil {
		listener.Close()
		t.Fatalf("expected strict conflict on port %d", port)
	}
	if got := err.Error(); got == "" || !strings.Contains(got, fmt.Sprintf("%d", port)) {
		t.Fatalf("expected error to mention port %d, got %q", port, got)
	}
}
