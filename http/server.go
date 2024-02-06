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
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

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

// DefaultOpts returns the default options for the server.
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
	// address is the assigned listen address for the server.
	//
	// The value is atomic to avoid a race condition between
	// the listener starting and calling AssignedAddr.
	address atomic.Value
	options *ServerOptions
	server  *http.Server
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

	var address atomic.Value
	address.Store("")

	return &Server{
		address: address,
		options: options,
		server:  server,
	}, nil
}

// AssignedAddr returns the address that was assigned to the server on calls to listen.
func (s *Server) AssignedAddr() string {
	return s.address.Load().(string)
}

// Shutdown gracefully shuts down the server without interrupting any active connections.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// ListenAndServe listens for and serves incoming connections.
func (s *Server) ListenAndServe() error {
	if s.options.TLSCertPath == "" && s.options.TLSKeyPath == "" {
		return s.listenAndServe()
	}
	return s.listenAndServeTLS()
}

// listenAndServe listens for and serves http connections.
func (s *Server) listenAndServe() error {
	listener, err := net.Listen("tcp", s.options.Address)
	if err != nil {
		return err
	}
	s.address.Store(listener.Addr().String())
	return s.server.Serve(listener)
}

// listenAndServeTLS listens for and serves https connections.
func (s *Server) listenAndServeTLS() error {
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
	listener, err := net.Listen("tcp", s.options.Address)
	if err != nil {
		return err
	}
	s.address.Store(listener.Addr().String())
	return s.server.Serve(tls.NewListener(listener, config))
}
