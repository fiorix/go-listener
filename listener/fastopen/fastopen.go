// +build linux freebsd darwin

package fastopen

import (
	"net"
	"syscall"
)

// Enable enables TCP fast open for the given listener.
func Enable(ln *net.TCPListener) error {
	file, err := ln.File()
	if err != nil {
		return err
	}
	fd := int(file.Fd())
	return syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, SocketOption, 1)
}
