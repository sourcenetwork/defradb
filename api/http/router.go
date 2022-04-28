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
	"path"

	"github.com/go-chi/chi"
)

const version = "/api/v1"

func setRoutes(h *Handler) *Handler {
	h.Mux = chi.NewRouter()

	// setup logger middleware
	h.Use(loggerMiddleware)

	// define routes
	h.Get(apiPath("/"), h.handle(root))
	h.Get(apiPath("/ping"), h.handle(ping))
	h.Get(apiPath("/dump"), h.handle(dump))
	h.Get(apiPath("/blocks/get/{cid}"), h.handle(getBlock))
	h.Get(apiPath("/graphql"), h.handle(execGQL))
	h.Post(apiPath("/graphql"), h.handle(execGQL))
	h.Post(apiPath("/schema/load"), h.handle(loadSchema))

	return h
}

func apiPath(pattern string) string {
	return path.Join(version, pattern)
}
