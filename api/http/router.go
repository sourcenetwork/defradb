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
	"github.com/go-chi/chi"
)

func (h *handler) setRoutes() {
	h.Mux = chi.NewRouter()

	// setup logger middleware
	h.Use(h.loggerMiddleware)

	// define routes
	h.Get("/", h.handle(root))
	h.Get("/ping", h.handle(ping))
	h.Get("/dump", h.handle(dump))
	h.Get("/blocks/get/{cid}", h.handle(getBlock))
	h.Get("/graphql", h.handle(execGQL))
	h.Post("/graphql", h.handle(execGQL))
	h.Post("/schema/load", h.handle(loadSchema))
}
