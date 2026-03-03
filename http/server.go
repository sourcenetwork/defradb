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

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/utils"
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

const defaultHTTPAddress = "127.0.0.1:9181"

// Server struct holds the Handler for the HTTP API.
type Server struct {
	options   *options.NodeHTTPOptions
	server    *http.Server
	listener  net.Listener
	isTLS     bool
	ctxCancel context.CancelFunc
}

// NewServer instantiates a new server with the given http.Handler.
func NewServer(handler http.Handler, opts ...options.Enumerable[options.NodeHTTPOptions]) (*Server, error) {
	cfg := options.NodeHTTPOptions{
		Address: defaultHTTPAddress,
	}
	utils.ApplyOptions(&cfg, opts...)

	ctx, cancel := context.WithCancel(context.Background())
	// setup a mux with the default middleware stack
	mux := chi.NewMux()
	mux.Use(
		InjectServerContext(ctx),
		middleware.RequestLogger(&logFormatter{}),
		middleware.Recoverer,
		CorsMiddleware(cfg.AllowedOrigins),
	)
	mux.Handle("/*", handler)

	server := &http.Server{
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
		Handler:      mux,
	}

	return &Server{
		options:   &cfg,
		server:    server,
		ctxCancel: cancel,
	}, nil
}

// Shutdown gracefully shuts down the server without interrupting any active connections.
func (s *Server) Shutdown(ctx context.Context) error {
	s.ctxCancel()
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

// InjectServerContext sets the server context on each handler calls.
func InjectServerContext(serverCtx context.Context) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			ctx = context.WithValue(ctx, ctxContextKey, serverCtx)
			next.ServeHTTP(rw, req.WithContext(ctx))
		})
	}
}
