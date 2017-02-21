// Package listenercmd provides env and flag configuration for go-listener.
package listenercmd

import (
	"crypto/tls"
	"flag"
	"strings"

	"github.com/kelseyhightower/envconfig"

	"github.com/fiorix/go-listener/listener"
)

// Config is the listener configuration for command line tools.
type Config struct {
	ListenAddr           string `envconfig:"LISTEN_ADDR"`
	Naggle               bool   `envconfig:"TCP_NAGGLE"`
	FastOpen             bool   `envconfig:"TCP_FAST_OPEN"`
	TLS                  bool   `envconfig:"TLS"`
	TLSCACertFile        string `envconfig:"TLS_CA_CERT_FILE" default:"cacert.pem"`
	TLSClientAuth        string `envconfig:"TLS_CLIENT_AUTH"`
	TLSCertFile          string `envconfig:"TLS_CERT_FILE" default:"cert.pem"`
	TLSKeyFile           string `envconfig:"TLS_KEY_FILE" default:"key.pem"`
	LetsEncrypt          bool   `envconfig:"LETSENCRYPT"`
	LetsEncryptCacheFile string `envconfig:"LETSENCRYPT_CACHE_FILE" default:"letsencrypt.cache"`
	LetsEncryptEmail     string `envconfig:"LETSENCRYPT_EMAIL"`
	LetsEncryptHosts     string `envconfig:"LETSENCRYPT_HOSTS"`
}

// NewConfig creates a Config with values from environment variables.
func NewConfig(prefixes ...string) *Config {
	c := &Config{}
	envconfig.Process(strings.Join(prefixes, "_"), c)
	return c
}

// AddFlags adds Config options to the given FlagSet.
func (c *Config) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.ListenAddr, "listen-addr", c.ListenAddr, "address in form of [ip]:port to listen on")
	fs.BoolVar(&c.Naggle, "tcp-naggle", c.Naggle, "enable tcp nagle's algorithm")
	fs.BoolVar(&c.FastOpen, "tcp-fast-open", c.FastOpen, "enable tcp fast open")
	fs.BoolVar(&c.TLS, "tls", c.TLS, "enable tls (requires cert file and key file)")
	fs.StringVar(&c.TLSCACertFile, "tls-ca-cert-file", c.TLSCACertFile, "ca certificate file (for client auth)")
	fs.StringVar(&c.TLSClientAuth, "tls-client-auth", c.TLSClientAuth, "client auth policy: RequestClientCert, RequireAnyClientCert, VerifyClientCertIfGiven, RequireAndVerifyClientCert")
	fs.StringVar(&c.TLSCertFile, "tls-cert-file", c.TLSCertFile, "server certificate file")
	fs.StringVar(&c.TLSKeyFile, "tls-key-file", c.TLSKeyFile, "server key file")
	fs.BoolVar(&c.LetsEncrypt, "letsencrypt", c.LetsEncrypt, "enable automatic tls using letsencrypt.org")
	fs.StringVar(&c.LetsEncryptEmail, "letsencrypt-email", c.LetsEncryptEmail, "optional email to register with letsencrypt (default is anonymous)")
	fs.StringVar(&c.LetsEncryptHosts, "letsencrypt-hosts", c.LetsEncryptHosts, "comma separated list of hosts for the certificate (any otherwise)")
	fs.StringVar(&c.LetsEncryptCacheFile, "letsencrypt-cache-file", c.LetsEncryptCacheFile, "letsencrypt cache file (for storing cert data)")
}

var clientAuthType = map[string]tls.ClientAuthType{
	"requestclientcert":          tls.RequestClientCert,
	"requireanyclientcert":       tls.RequireAnyClientCert,
	"verifyclientcertifgiven":    tls.VerifyClientCertIfGiven,
	"requireandverifyclientcert": tls.RequireAndVerifyClientCert,
}

// Options returns listener options from the Config.
func (c *Config) Options() []listener.Option {
	var o []listener.Option
	if c.Naggle {
		o = append(o, listener.Naggle())
	}
	if c.FastOpen {
		o = append(o, listener.FastOpen())
	}
	if c.TLS {
		o = append(o, listener.TLS(c.TLSCertFile, c.TLSKeyFile))
	}
	if c.TLSClientAuth != "" {
		opt := clientAuthType[strings.ToLower(c.TLSClientAuth)]
		o = append(o, listener.TLSClientAuth(c.TLSCACertFile, opt))
	}
	if c.LetsEncrypt {
		o = append(o, listener.LetsEncrypt(
			c.LetsEncryptCacheFile,
			c.LetsEncryptEmail,
			strings.Split(c.LetsEncryptHosts, ",")...,
		))
	}
	return o
}
