// Copyright 2022 Democratized Data Foundation
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
	"net/http"

	"github.com/sourcenetwork/defradb/client"
)

// The Server struct holds the Handler for the HTTP API
type Server struct {
	serverOptions
	http.Server
}

type serverOptions struct {
	allowedOrigins []string
}

// NewServer instantiated a new server with the given http.Handler.
func NewServer(db client.DB, options ...func(*Server)) *Server {
	svr := &Server{}

	for _, opt := range append(options, DefaultOpts()) {
		opt(svr)
	}

	svr.Server.Handler = newHandler(db, svr.serverOptions)

	return svr
}

func DefaultOpts() func(*Server) {
	return func(s *Server) {
		if len(s.allowedOrigins) == 0 {
			s.allowedOrigins = []string{"https://*", "http://*"}
		}

		if s.Addr == "" {
			s.Addr = "localhost:9181"
		}
	}
}

func WithAllowedOrigins(origins ...string) func(*Server) {
	return func(s *Server) {
		s.allowedOrigins = append(s.allowedOrigins, origins...)
	}
}

func WithAddress(addr string) func(*Server) {
	return func(s *Server) {
		s.Addr = addr
	}
}

// Listen calls ListenAndServe with our router.
func (s *Server) Listen() error {
	return s.ListenAndServe()
}
