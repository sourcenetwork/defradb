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

	"github.com/go-chi/chi"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/logging"
)

type Handler struct {
	db client.DB

	*chi.Mux
	*logger
}

// newHandler returns a handler with the router instantiated and configuration applied.
func newHandler(db client.DB, options ...func(*Handler)) *Handler {
	h := &Handler{
		db: db,
	}

	// apply options
	for _, o := range options {
		o(h)
	}

	// ensure we have a logger defined
	if h.logger == nil {
		h.logger = defaultLogger()
	}

	h.setRoutes()

	return h
}

// WithLogger returns an option loading function for logger.
func WithLogger(l logging.Logger) func(*Handler) {
	return func(h *Handler) {
		h.logger = &logger{l}
	}
}

type requestContext struct {
	res http.ResponseWriter
	req *http.Request
	db  client.DB
	log *logger
}

func (h *Handler) handle(f func(*requestContext)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f(&requestContext{
			res: w,
			req: r,
			db:  h.db,
			log: h.logger,
		})
	}
}
