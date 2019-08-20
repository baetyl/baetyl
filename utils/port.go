package utils

import "net"

// GetAvailablePort finds an available port
func GetAvailablePort(host string) (int, error) {
	address, err := net.ResolveTCPAddr("tcp", host+":0")
	if err != nil {
		return 0, err
	}

	listener, err := net.ListenTCP("tcp", address)
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

// CheckPortsInUse check if the ports are in use
func CheckPortsInUse(ports []string) ([]string, bool) {
	var usedPorts []string
	for _, port := range ports {
		conn, _ := net.Dial("tcp", net.JoinHostPort("", port))
		if conn != nil {
			conn.Close()
			usedPorts = append(usedPorts, port)
		}
	}
	if len(usedPorts) > 0 {
		return usedPorts, true
	}
	return usedPorts, false
}
