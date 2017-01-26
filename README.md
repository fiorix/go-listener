# TCP listener for Go

[![Build Status](https://secure.travis-ci.org/fiorix/go-listener.png)](http://travis-ci.org/fiorix/go-listener)
[![GoDoc](https://godoc.org/github.com/fiorix/go-listener?status.svg)](https://godoc.org/github.com/fiorix/go-listener)

This is an implementation of the [net.Listener](https://golang.org/pkg/net/#Listener) interface for Go with support for:

* TCP fast open [RFC 7413](https://tools.ietf.org/html/rfc7413)
* TLS configuration with optional client authentication
* Automatic TLS configuration using [letsencrypt.org](https://letsencrypt.org)
* Command line configuration via environment variables, using [envconfig](https://github.com/kelseyhightower/envconfig)
* Command line configuration for [cobra](https://github.com/spf13/cobra)
