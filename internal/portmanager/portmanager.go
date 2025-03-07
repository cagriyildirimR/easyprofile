package portmanager

import (
	"fmt"
	"net"
	"sync"
)

type PortManager struct {
	mu          sync.Mutex
	basePort    int
	maxAttempts int
}

func New(basePort int) *PortManager {
	return &PortManager{
		basePort:    basePort,
		maxAttempts: 20,
	}
}

func (pm *PortManager) GetPort() (int, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for attempt := 0; attempt < pm.maxAttempts; attempt++ {
		port := pm.basePort + attempt

		addr := fmt.Sprintf(":%d", port)
		listener, err := net.Listen("tcp", addr)
		if err == nil {
			listener.Close()
			pm.basePort = port + 1
			return port, nil
		}
	}

	return 0, fmt.Errorf("no available ports found after %d attempts starting from port %d", pm.maxAttempts, pm.basePort)
}
