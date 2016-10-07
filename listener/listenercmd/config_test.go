package listenercmd

import (
	"os"
	"testing"

	"github.com/spf13/pflag"
)

func TestConfig(t *testing.T) {
	addr := os.Getenv("LISTEN_ADDR")
	os.Setenv("LISTEN_ADDR", ":8888")
	defer os.Setenv("LISTEN_ADDR", addr)
	c := NewConfig()
	if c.ListenAddr != ":8888" {
		t.Fatalf("unexpected listen addr: want :8888, have %q", c.ListenAddr)
	}
}

func TestConfigFlags(t *testing.T) {
	c := NewConfig()
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	c.AddFlags(fs)
	err := fs.Parse([]string{"--listen-addr=:9999"})
	if err != nil {
		t.Fatal(err)
	}
	v, err := fs.GetString("listen-addr")
	if err != nil {
		t.Fatal(err)
	}
	if v != ":9999" {
		t.Fatalf("unexpected listen addr: want :8888, have %q", v)
	}
}

func TestConfigOptions(t *testing.T) {
	c := NewConfig()
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	c.AddFlags(fs)
	err := fs.Parse([]string{
		"--tcp-naggle",
		"--tcp-fast-open",
		"--tls",
		"--tls-cert-file=../testdata/cert.pem",
		"--tls-key-file=../testdata/key.pem",
		"--tls-client-auth=VerifyClientCertIfGiven",
		"--letsencrypt",
	})
	if err != nil {
		t.Fatal(err)
	}
	o := c.Options()
	if len(o) != 5 {
		t.Fatalf("unexpected opts count: want 5, have %d", len(o))
	}
}