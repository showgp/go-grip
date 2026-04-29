package internal

import (
	"fmt"
	"net"
)

const maxPortAttempts = 100

func listenOnPort(port int, strict bool) (net.Listener, int, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err == nil {
		return listener, listenerPort(listener), nil
	}
	if strict {
		return nil, 0, fmt.Errorf("listen on port %d: %w", port, err)
	}

	for candidate := port + 1; candidate < port+maxPortAttempts; candidate++ {
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", candidate))
		if err == nil {
			return listener, listenerPort(listener), nil
		}
	}

	return nil, 0, fmt.Errorf("no available port found starting at %d", port)
}

func listenerPort(listener net.Listener) int {
	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return 0
	}
	return addr.Port
}
