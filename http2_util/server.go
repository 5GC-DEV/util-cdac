// Copyright 2019 Communication Service/Software Laboratory, National Chiao Tung University (free5gc.org)
//
// SPDX-License-Identifier: Apache-2.0

//go:build !debug
// +build !debug

package http2_util

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/net/http2"
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

func NewServer(bindAddr, preMasterSecretLogPath string, handler http.Handler) (*http.Server, error) {
	if handler == nil {
		return nil, fmt.Errorf("server needs handler to handle request")
	}

	// HTTP/2 server config
	h2s := &http2.Server{}

	// Base HTTP server (no h2c)
	server := &http.Server{
		Addr:    bindAddr,
		Handler: handler, // <<--- NO H2C WRAPPING HERE
	}
	// Enable TLS Key Logging if path provided
	if preMasterSecretLogPath != "" {
		keyLogFile, err := os.OpenFile(preMasterSecretLogPath,
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
		if err != nil {
			return nil, fmt.Errorf("failed to create keylog file: %w", err)
		}

		server.TLSConfig = &tls.Config{
			MinVersion:   tls.VersionTLS12,
			KeyLogWriter: keyLogFile,     // <<--- THIS IS THE IMPORTANT PART
			NextProtos:   []string{"h2"}, // Enable HTTP/2 over TLS
		}
		// Configure HTTP/2 support
		if err := http2.ConfigureServer(server, h2s); err != nil {
			return nil, fmt.Errorf("failed to configure http2: %w", err)
		}
	}

	return server, nil
}
