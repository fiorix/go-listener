// +build !linux,!freebsd,!darwin

package fastopen

import "net"

// Enable enables TCP fast open for the given listener.
// Not supported on this platform.
func Enable(ln *net.TCPListener) error {
	return nil
}
