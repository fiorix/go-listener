package listener

import (
	"crypto/tls"
	"io"
	"net"
	"strconv"
	"testing"
	"time"
)

func TestFastOpen(t *testing.T) {
	c := &config{}
	err := FastOpen()(c)
	if err != nil {
		t.Fatal(err)
	}
	if !c.fastOpen {
		t.Fatal("fast open not enabled")
	}
}

func TestNaggle(t *testing.T) {
	c := &config{}
	err := Naggle()(c)
	if err != nil {
		t.Fatal(err)
	}
	if !c.naggle {
		t.Fatal("naggle not enabled")
	}
}

func TestTLS(t *testing.T) {
	testCases := []struct {
		keyFile, certFile string
		err               bool
	}{
		{
			certFile: "",
			err:      true,
		},
		{
			certFile: "testdata/cert.pem",
			keyFile:  "testdata/key.pem",
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			c := &config{}
			err := TLS(tc.certFile, tc.keyFile)(c)
			if err != nil {
				if !tc.err {
					t.Fatal(err)
				}
			}
			if !tc.err && c.tls == nil {
				t.Fatal("tls not configured")
			}
		})
	}
}

func TestLetsEncrypt(t *testing.T) {
	c := &config{}
	err := LetsEncrypt("", "")(c)
	if err == nil {
		t.Fatal("instance created without host config")
	}
	err = LetsEncrypt("testdata/", "root@localhost", "localhost")(c)
	if err != nil {
		t.Fatal(err)
	}
	if c.tls == nil {
		t.Fatal("tls not configured")
	}
}

func TestTLSClientAuth(t *testing.T) {
	testCases := []struct {
		cacertFile string
		authType   tls.ClientAuthType
		err        bool
	}{
		{
			cacertFile: "",
			err:        true,
		},
		{
			cacertFile: "testdata/cacert.pem",
			authType:   0,
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			c := &config{}
			err := TLSClientAuth(tc.cacertFile, tc.authType)(c)
			if err != nil {
				if !tc.err {
					t.Fatal(err)
				}
			}
			if !tc.err && c.tls == nil {
				t.Fatal("tls not configured")
			}
		})
	}
}

func TestListener(t *testing.T) {
	ln, err := New("")
	if err != nil {
		t.Fatal(err)
	}
	ln.Close()
}

func TestSecureListener(t *testing.T) {
	ln, err := New("", TLS("testdata/cert.pem", "testdata/key.pem"))
	if err != nil {
		t.Fatal(err)
	}
	ln.Close()
}

func TestListenerOptionError(t *testing.T) {
	ln, err := New("", TLS("", ""))
	if err == nil {
		defer ln.Close()
		t.Fatalf("unexpected listener on %s", ln.Addr())
	}
}

func TestListenerError(t *testing.T) {
	ln, err := New(":fail")
	if err == nil {
		defer ln.Close()
		t.Fatalf("unexpected listener on %s", ln.Addr())
	}
}

func TestListenerAccept(t *testing.T) {
	ln, err := New("")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	errc := make(chan error, 1)
	go func() {
		cli, err := ln.Accept()
		if err != nil {
			errc <- err
		}
		cli.Close()
	}()
	cn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	b := make([]byte, 1024)
	cn.SetReadDeadline(time.Now().Add(time.Second))
	_, err = cn.Read(b)
	if err != io.EOF {
		t.Fatal(err)
	}
	select {
	case err = <-errc:
		t.Fatal("accept failed:", err)
	default:
	}
}
