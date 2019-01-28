package utils

import "net"

// GetPortAvailable finds an available port
func GetPortAvailable(host string) (int, error) {
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
