// Package listener provides a TCP listener on roids.
package listener

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"time"

	"rsc.io/letsencrypt"

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
			return fmt.Errorf("cert/key: %v", err)
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
// The cache file is used to store the certificate retrieved from LetsEncrypt
// and reuse on server restarts. If not specified, "letsencrypt.cache" is used.
//
// If an email is not provided for registration with LetsEncrypt, it is
// registered anonymously.
//
// An optional list of hosts can be configured to limit the hosts served
// by the listener. All hosts are allowed otherwise.
//
// By calling this function you are accepting LetsEncrypt's TOS.
// https://letsencrypt.org/repository/
func LetsEncrypt(cacheFile, email string, hosts ...string) Option {
	return func(c *config) error {
		var err error
		var m letsencrypt.Manager
		m.SetHosts(hosts)
		if email != "" {
			err = m.Register(email, nil)
			if err != nil {
				return err
			}
		}
		if cacheFile == "" {
			cacheFile = "letsencrypt.cache"
		}
		if err = m.CacheFile(cacheFile); err != nil {
			return err
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
			return fmt.Errorf("ca cert: %v", err)
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

// New creates and initializes a new TCP listener.
func New(addr string, opts ...Option) (net.Listener, error) {
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
		return ln, nil
	}
	return tls.NewListener(ln, c.tls), nil
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
