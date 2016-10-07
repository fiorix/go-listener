package fastopen

import (
	"net"
	"testing"
)

func TestFastOpen(t *testing.T) {
	ln, err := net.Listen("tcp", "")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	err = Enable(ln.(*net.TCPListener))
	if err != nil {
		t.Fatal(err)
	}
}
