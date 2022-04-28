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
	"context"
	"net/url"
	"path"

	"github.com/go-chi/chi"
)

const (
	version string = "/api/v1"

	HomePath       string = version + "/"
	PingPath       string = version + "/ping"
	DumpPath       string = version + "/dump"
	BlocksPath     string = version + "/blocks/get"
	GraphQLPath    string = version + "/graphql"
	SchemaLoadPath string = version + "/schema/load"
)

func setRoutes(h *Handler) *Handler {
	h.Mux = chi.NewRouter()

	// setup logger middleware
	h.Use(loggerMiddleware)

	// define routes
	h.Get(HomePath, h.handle(root))
	h.Get(PingPath, h.handle(ping))
	h.Get(DumpPath, h.handle(dump))
	h.Get(BlocksPath+"/{cid}", h.handle(getBlock))
	h.Get(GraphQLPath, h.handle(execGQL))
	h.Post(GraphQLPath, h.handle(execGQL))
	h.Post(SchemaLoadPath, h.handle(loadSchema))

	return h
}

// JoinPaths takes a base path and any number of additionnal paths
// and combines them safely to form a full URL path or a simple path if
// the base parameter is not a valid URL starting with `http://` or `https://`.
func JoinPaths(base string, paths ...string) string {
	u, err := url.Parse(base)
	if err != nil {
		log.Error(context.Background(), err.Error())
		paths = append(([]string{base}), paths...)
		return path.Join(paths...)
	}

	paths = append(([]string{u.Path}), paths...)
	u.Path = path.Join(paths...)

	return u.String()
}
