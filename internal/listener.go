package internal

import (
	"fmt"
	"net"
	"strconv"
)

const maxPortAttempts = 100

// listenOn binds a TCP listener on the given bind host. An empty host keeps the
// standalone CLI policy of listening on all interfaces; the managed preview
// path passes "127.0.0.1" so the App-owned service really is loopback-only.
// Port fallback and strict behavior are unchanged.
func listenOn(bind string, port int, strict bool) (net.Listener, int, error) {
	listen := func(candidate int) (net.Listener, error) {
		return net.Listen("tcp", net.JoinHostPort(bind, strconv.Itoa(candidate)))
	}

	listener, err := listen(port)
	if err == nil {
		return listener, listenerPort(listener), nil
	}
	if strict {
		return nil, 0, fmt.Errorf("listen on port %d: %w", port, err)
	}

	for candidate := port + 1; candidate < port+maxPortAttempts; candidate++ {
		listener, err := listen(candidate)
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
