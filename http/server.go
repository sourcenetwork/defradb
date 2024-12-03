// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package http

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// PlaygroundEnabled is used to detect if the playground is enabled
// on the current http server instance.
var PlaygroundEnabled = false

// We only allow cipher suites that are marked secure
// by ssllabs
var tlsCipherSuites = []uint16{
	tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
}

type ServerOptions struct {
	// Address is the bind address the server listens on.
	Address string
	// AllowedOrigins is the list of allowed origins for CORS.
	AllowedOrigins []string
	// TLSCertPath is the path to the TLS certificate.
	TLSCertPath string
	// TLSKeyPath is the path to the TLS key.
	TLSKeyPath string
	// ReadTimeout is the read timeout for connections.
	ReadTimeout time.Duration
	// WriteTimeout is the write timeout for connections.
	WriteTimeout time.Duration
	// IdleTimeout is the idle timeout for connections.
	IdleTimeout time.Duration
}

// DefaultServerOptions returns the default options for the server.
func DefaultServerOptions() *ServerOptions {
	return &ServerOptions{
		Address: "127.0.0.1:9181",
	}
}

// ServerOpt is a function that configures server options.
type ServerOpt func(*ServerOptions)

// WithAllowedOrigins sets the allowed origins for CORS.
func WithAllowedOrigins(origins ...string) ServerOpt {
	return func(opts *ServerOptions) {
		opts.AllowedOrigins = origins
	}
}

// WithAddress sets the bind address for the server.
func WithAddress(addr string) ServerOpt {
	return func(opts *ServerOptions) {
		opts.Address = addr
	}
}

// WithReadTimeout sets the server read timeout.
func WithReadTimeout(timeout time.Duration) ServerOpt {
	return func(opts *ServerOptions) {
		opts.ReadTimeout = timeout
	}
}

// WithWriteTimeout sets the server write timeout.
func WithWriteTimeout(timeout time.Duration) ServerOpt {
	return func(opts *ServerOptions) {
		opts.WriteTimeout = timeout
	}
}

// WithIdleTimeout sets the server idle timeout.
func WithIdleTimeout(timeout time.Duration) ServerOpt {
	return func(opts *ServerOptions) {
		opts.IdleTimeout = timeout
	}
}

// WithTLSCertPath sets the server TLS certificate path.
func WithTLSCertPath(path string) ServerOpt {
	return func(opts *ServerOptions) {
		opts.TLSCertPath = path
	}
}

// WithTLSKeyPath sets the server TLS private key path.
func WithTLSKeyPath(path string) ServerOpt {
	return func(opts *ServerOptions) {
		opts.TLSKeyPath = path
	}
}

// Server struct holds the Handler for the HTTP API.
type Server struct {
	options  *ServerOptions
	server   *http.Server
	listener net.Listener
	isTLS    bool
}

// NewServer instantiates a new server with the given http.Handler.
func NewServer(handler http.Handler, opts ...ServerOpt) (*Server, error) {
	options := DefaultServerOptions()
	for _, opt := range opts {
		opt(options)
	}

	// setup a mux with the default middleware stack
	mux := chi.NewMux()
	mux.Use(
		middleware.RequestLogger(&logFormatter{}),
		middleware.Recoverer,
		CorsMiddleware(options.AllowedOrigins),
	)
	mux.Handle("/*", handler)

	server := &http.Server{
		ReadTimeout:  options.ReadTimeout,
		WriteTimeout: options.WriteTimeout,
		IdleTimeout:  options.IdleTimeout,
		Handler:      mux,
	}

	return &Server{
		options: options,
		server:  server,
	}, nil
}

// Shutdown gracefully shuts down the server without interrupting any active connections.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// SetListener sets a new listener on the Server.
func (s *Server) SetListener() (err error) {
	s.listener, err = net.Listen("tcp", s.options.Address)
	return err
}

// Serve serves incoming connections.
func (s *Server) Serve() error {
	if s.options.TLSCertPath == "" && s.options.TLSKeyPath == "" {
		return s.serve()
	}
	s.isTLS = true
	return s.serveTLS()
}

// serve serves http connections.
func (s *Server) serve() error {
	if s.listener == nil {
		return ErrNoListener
	}
	return s.server.Serve(s.listener)
}

// serveTLS serves https connections.
func (s *Server) serveTLS() error {
	if s.listener == nil {
		return ErrNoListener
	}
	cert, err := tls.LoadX509KeyPair(s.options.TLSCertPath, s.options.TLSKeyPath)
	if err != nil {
		return err
	}
	config := &tls.Config{
		ServerName:   "DefraDB",
		MinVersion:   tls.VersionTLS12,
		CipherSuites: tlsCipherSuites,
		Certificates: []tls.Certificate{cert},
	}
	return s.server.Serve(tls.NewListener(s.listener, config))
}

func (s *Server) Address() string {
	if s.isTLS {
		return "https://" + s.listener.Addr().String()
	}
	return "http://" + s.listener.Addr().String()
}
