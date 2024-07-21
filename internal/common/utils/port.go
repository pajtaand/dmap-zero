package utils

import (
	"fmt"
	"net"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	requestTimeout = 250 * time.Microsecond
)

func TCPPortAvailable(port int) bool {
	conn, _ := net.DialTimeout("tcp", fmt.Sprintf(":%d", port), requestTimeout)
	if conn != nil {
		conn.Close()
	}
	return conn == nil
}

func FirstAvailablePort(port int) int {
	for !TCPPortAvailable(port) {
		log.Debug().Msgf("Port %d is not available", port)
		port += 1
		log.Debug().Msgf("Trying port: %d...", port)
	}
	return port
}
