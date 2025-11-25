// Copyright 2019 Communication Service/Software Laboratory, National Chiao Tung University (free5gc.org)
//
// SPDX-License-Identifier: Apache-2.0

//go:build !debug
// +build !debug

package http2_util

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"

	// "time"

	"github.com/pkg/errors"
	// "golang.org/x/net/http2"
	// "golang.org/x/net/http2/h2c"
)

// NewServer returns a server instance with HTTP/2.0 and HTTP/2.0 cleartext support
// If this function cannot open or create the secret log file,
// **it still returns server instance** but without the secret log and error indication
/*func NewServer(bindAddr string, preMasterSecretLogPath string, handler http.Handler) (server *http.Server, err error) {
	if handler == nil {
		return nil, errors.New("server needs handler to handle request")
	}

	h2Server := &http2.Server{
		// TODO: extends the idle time after re-use openapi client
		IdleTimeout: 1 * time.Millisecond,
	}
	server = &http.Server{
		Addr:    bindAddr,
		Handler: h2c.NewHandler(handler, h2Server),
	}

	if preMasterSecretLogPath != "" {
		preMasterSecretFile, err := os.OpenFile(preMasterSecretLogPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
		if err != nil {
			return server, fmt.Errorf("create pre-master-secret log [%s] fail: %s", preMasterSecretLogPath, err)
		}
		server.TLSConfig = &tls.Config{
			KeyLogWriter: preMasterSecretFile,
		}
	}

	return
} */

func NewServer(bindAddr, keyLogPath, certPath, keyPath string, handler http.Handler) (*http.Server, error) {
	if handler == nil {
		return nil, errors.New("server needs handler")
	}

	// --- TLS KEY LOG FILE ---
	var keyLogWriter io.Writer
	if keyLogPath != "" {
		f, err := os.OpenFile(keyLogPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return nil, fmt.Errorf("open keylog file failed: %v", err)
		}
		keyLogWriter = f
	}

	// --- LOAD CERTIFICATE ---
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf("load cert failed: %v", err)
	}

	// --- TLS CONFIG ---
	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		KeyLogWriter: keyLogWriter,
		NextProtos:   []string{"h2"}, // HTTP/2 ALPN
		MinVersion:   tls.VersionTLS12,
	}

	// --- HTTP SERVER ---
	srv := &http.Server{
		Addr:      bindAddr,
		Handler:   handler,
		TLSConfig: tlsCfg,
	}

	return srv, nil
}
