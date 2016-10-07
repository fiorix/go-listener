// +build linux freebsd darwin

package fastopen

import (
	"net"
	"reflect"
	"syscall"
)

// Enable enables TCP fast open for the given listener.
func Enable(ln *net.TCPListener) error {
	fd := int(reflect.ValueOf(ln).Elem().FieldByName("fd").Elem().FieldByName("sysfd").Int())
	return syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, SocketOption, 1)
}
