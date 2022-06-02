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
	"net/url"
	"path"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/pkg/errors"
)

const (
	version string = "/api/v0"

	RootPath       string = version + ""
	PingPath       string = version + "/ping"
	DumpPath       string = version + "/debug/dump"
	BlocksPath     string = version + "/blocks/get"
	GraphQLPath    string = version + "/graphql"
	SchemaLoadPath string = version + "/schema/load"
)

var schemeError = errors.New("base must start with the http or https scheme")

func setRoutes(h *handler) *handler {
	h.Mux = chi.NewRouter()

	// setup CORS
	if len(h.options.allowedOrigins) != 0 {
		h.Use(cors.Handler(cors.Options{
			AllowedOrigins: h.options.allowedOrigins,
			AllowedMethods: []string{"GET", "POST", "OPTIONS"},
			AllowedHeaders: []string{"Content-Type"},
			MaxAge:         300,
		}))
	}

	// setup logger middleware
	h.Use(loggerMiddleware)

	// define routes
	h.Get(RootPath, h.handle(rootHandler))
	h.Get(PingPath, h.handle(pingHandler))
	h.Get(DumpPath, h.handle(dumpHandler))
	h.Get(BlocksPath+"/{cid}", h.handle(getBlockHandler))
	h.Get(GraphQLPath, h.handle(execGQLHandler))
	h.Post(GraphQLPath, h.handle(execGQLHandler))
	h.Post(SchemaLoadPath, h.handle(loadSchemaHandler))

	return h
}

// JoinPaths takes a base path and any number of additionnal paths
// and combines them safely to form a full URL path.
// The base must start with a http or https.
func JoinPaths(base string, paths ...string) (*url.URL, error) {
	if !strings.HasPrefix(base, "http") {
		return nil, schemeError
	}

	u, err := url.Parse(base)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	u.Path = path.Join(u.Path, strings.Join(paths, "/"))

	return u, nil
}
