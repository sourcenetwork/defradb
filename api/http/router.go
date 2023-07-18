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
	// Version is the current version of the HTTP API.
	Version          string = "v0"
	versionedAPIPath string = "/api/" + Version

	RootPath            string = versionedAPIPath + ""
	PingPath            string = versionedAPIPath + "/ping"
	DumpPath            string = versionedAPIPath + "/debug/dump"
	BlocksPath          string = versionedAPIPath + "/blocks"
	GraphQLPath         string = versionedAPIPath + "/graphql"
	SchemaPath          string = versionedAPIPath + "/schema"
	SchemaMigrationPath string = SchemaPath + "/migration"
	IndexPath           string = versionedAPIPath + "/index"
	PeerIDPath          string = versionedAPIPath + "/peerid"
)

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
	h.Get(SchemaPath, h.handle(listSchemaHandler))
	h.Post(SchemaPath, h.handle(loadSchemaHandler))
	h.Patch(SchemaPath, h.handle(patchSchemaHandler))
	h.Post(SchemaMigrationPath, h.handle(setMigrationHandler))
	h.Get(SchemaMigrationPath, h.handle(getMigrationHandler))
	h.Post(IndexPath, h.handle(createIndexHandler))
	h.Delete(IndexPath, h.handle(dropIndexHandler))
	h.Get(IndexPath, h.handle(listIndexHandler))
	h.Get(PeerIDPath, h.handle(peerIDHandler))

	return h
}

// JoinPaths takes a base path and any number of additional paths
// and combines them safely to form a full URL path.
// The base must start with a http or https.
func JoinPaths(base string, paths ...string) (*url.URL, error) {
	if !strings.HasPrefix(base, "http") {
		return nil, ErrSchema
	}

	u, err := url.Parse(base)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	u.Path = path.Join(u.Path, strings.Join(paths, "/"))

	return u, nil
}
