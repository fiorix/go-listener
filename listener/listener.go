// Package listener provides a TCP listener on roids.
package listener

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"time"

	"golang.org/x/crypto/acme/autocert"

	"github.com/fiorix/go-listener/listener/fastopen"
)

// Option type for listener options.
type Option func(*config) error

type config struct {
	naggle   bool
	fastOpen bool
	tls      *tls.Config
}

// FastOpen enables TCP fast open.
func FastOpen() Option {
	return func(c *config) error {
		c.fastOpen = true
		return nil
	}
}

// Naggle enables Naggle's algorithm - effectively setting nodelay=false.
// This might be useful when combined with fast open, to allow data on ack.
func Naggle() Option {
	return func(c *config) error {
		c.naggle = true
		return nil
	}
}

// TLS configures TLS with certificate and key files.
func TLS(certFile, keyFile string) Option {
	return func(c *config) error {
		pair, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return fmt.Errorf("listener: cert/key failed: %v", err)
		}
		if c.tls == nil {
			c.tls = &tls.Config{}
		}
		c.tls.Certificates = append(c.tls.Certificates, pair)
		return nil
	}
}

// LetsEncrypt configures automatic TLS certificates using letsencrypt.org.
//
// The cache dir is used to store certificates retrieved from LetsEncrypt
// and reuse on server restarts. If not specified, "." is used.
//
// The email parameter is optionally used for registration with LetsEncrypt
// to notify about certificate problems. If not set, certificates are
// obtained anonymously.
//
// The hosts parameter must define a list of allowed hostnames to obtain
// certificates for with LetsEncrypt.
//
// By calling this function you are accepting LetsEncrypt's TOS.
// https://letsencrypt.org/repository/
func LetsEncrypt(cacheDir, email string, hosts ...string) Option {
	return func(c *config) error {
		if cacheDir == "" {
			cacheDir = "."
		}
		if len(hosts) == 0 {
			return errors.New("listener: no hosts configured")
		}
		m := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			Email:      email,
			Cache:      autocert.DirCache(cacheDir),
			HostPolicy: autocert.HostWhitelist(hosts...),
		}
		if c.tls == nil {
			c.tls = &tls.Config{}
		}
		c.tls.GetCertificate = m.GetCertificate
		return nil
	}
}

// TLSClientAuth configures TLS client certificate authentication.
func TLSClientAuth(cacertFile string, authType tls.ClientAuthType) Option {
	return func(c *config) error {
		cacert, err := ioutil.ReadFile(cacertFile)
		if err != nil {
			return fmt.Errorf("listener: ca cert: %v", err)
		}
		certpool := x509.NewCertPool()
		certpool.AppendCertsFromPEM(cacert)
		if c.tls == nil {
			c.tls = &tls.Config{}
		}
		c.tls.RootCAs = certpool
		c.tls.ClientAuth = authType
		return nil
	}
}

// HTTP2 configures TLS listeners to set the "h2" option for net/http to
// to serve HTTP/2 over TLS. This option can only be used when combined
// with other TLS settings, either using local key/cert or LetsEncrypt.
func HTTP2() Option {
	return func(c *config) error {
		if c.tls == nil {
			c.tls = &tls.Config{}
		}
		c.tls.NextProtos = append(c.tls.NextProtos, "h2")
		return nil
	}
}

// TCPListener implements the net.Listener interface. Listens on a TCP port
// that is configured with options from this package. e.g. fast-open, etc.
type TCPListener struct {
	net.Listener
	tlsConfig *tls.Config
}

// New creates and initializes a new TCP listener that is ready to accept
// client connections.
func New(addr string, opts ...Option) (*TCPListener, error) {
	var err error
	c := &config{}
	for _, o := range opts {
		if err = o(c); err != nil {
			return nil, err
		}
	}
	ln, err := listen(addr, c.naggle, c.fastOpen)
	if err != nil {
		return nil, err
	}
	if c.tls == nil {
		return &TCPListener{ln, nil}, nil
	}
	return &TCPListener{tls.NewListener(ln, c.tls), c.tls}, nil
}

func listen(addr string, naggle, fastOpen bool) (net.Listener, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	tcpln := ln.(*net.TCPListener)
	if fastOpen {
		err = fastopen.Enable(tcpln)
		if err != nil {
			return nil, err
		}
	}
	ln = &tcpKeepAliveListener{
		TCPListener: tcpln,
		NoDelay:     !naggle,
	}
	return ln, nil
}

// TLSConfig returns the TLS configuration for the TCP listener.
func (ln *TCPListener) TLSConfig() *tls.Config {
	return ln.tlsConfig
}

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
//
// Copyright: copied and adapted from net/http.
type tcpKeepAliveListener struct {
	*net.TCPListener
	NoDelay bool
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetNoDelay(ln.NoDelay)
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}
