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
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
)

type Router struct {
	mux chi.Router
	oas *openapi3.T
}

func NewRouter() (*Router, error) {
	oas, err := NewOpenAPISpec()
	if err != nil {
		return nil, err
	}
	return &Router{chi.NewMux(), oas}, nil
}

// AddMiddleware adds middleware functions to the current route group.
func (r *Router) AddMiddleware(middlewares ...func(http.Handler) http.Handler) {
	r.mux.Use(middlewares...)
}

// RouteGroup adds handlers as a group.
func (r *Router) AddRouteGroup(group func(*Router)) {
	r.mux.Group(func(router chi.Router) {
		group(&Router{router, r.oas})
	})
}

// AddRoute adds a handler for the given route.
func (r *Router) AddRoute(pattern, method string, op *openapi3.Operation, handler http.HandlerFunc) {
	r.mux.MethodFunc(method, pattern, handler)
	r.oas.AddOperation(pattern, method, op)
}

// Validate returns an error if the OpenAPI specification is invalid.
func (r *Router) Validate(ctx context.Context) error {
	loader := openapi3.NewLoader()
	if err := loader.ResolveRefsIn(r.oas, nil); err != nil {
		return err
	}
	return r.oas.Validate(ctx)
}

// OpenAPI returns the OpenAPI specification.
func (r *Router) OpenAPI() *openapi3.T {
	return r.oas
}

func (r *Router) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(rw, req)
}
