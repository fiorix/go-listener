package listener_test

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/fiorix/go-listener/listener"
)

func ExampleListener() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello, world")
	})
	ln, err := listener.New(":80", listener.FastOpen())
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	http.Serve(ln, nil)
}

func ExampleListenerTLS() {
	opts := []listener.Option{
		listener.FastOpen(),
		listener.TLS("cert.pem", "key.pem"),
		listener.TLSClientAuth("cacert.pem", tls.VerifyClientCertIfGiven),
	}
	ln, err := listener.New(":443", opts...)
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	// ...
}

func ExampleListenerLetsEncrypt() {
	opts := []listener.Option{
		listener.FastOpen(),
		listener.LetsEncrypt(
			"letsencrypt.cache", // cache file
			"me@example.com",    // optional email for registration
			"example.com",       // hosts...
			"api.example.com",
			"foobar.example.com",
		),
	}
	ln, err := listener.New(":443", opts...)
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	// ...
}
