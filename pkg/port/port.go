package port

import (
	"fmt"
	"net"
)

func free() (int, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, fmt.Errorf("get free port: %w", err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func Get(debug bool, fixed int) (int, error) {
	if debug {
		return fixed, nil
	}
	return free()
}
