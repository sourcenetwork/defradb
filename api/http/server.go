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
)

// The Server struct holds the Handler for the HTTP API
type Server struct {
	Handler http.Handler
}

// NewServer instantiated a new server.
func NewServer() *Server {
	return &Server{}
}

// Listen calls ListenAndServe with our router.
func (s *Server) Listen(addr string) error {
	return http.ListenAndServe(addr, s.Handler)
}
