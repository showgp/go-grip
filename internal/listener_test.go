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
	listener, actualPort, err := listenOnPort(port, false)
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
	listener, _, err := listenOnPort(port, true)
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
