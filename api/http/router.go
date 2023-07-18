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
	"net/url"
	"path"
	"strings"

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

// playgroundHandler is set when building with the playground build tag
var playgroundHandler http.Handler

func setRoutes(h *handler) *handler {
	h.Get(RootPath, rootHandler)
	h.Get(PingPath, pingHandler)
	h.Get(DumpPath, dumpHandler)
	h.Get(BlocksPath+"/{cid}", getBlockHandler)
	h.Get(GraphQLPath, execGQLHandler)
	h.Post(GraphQLPath, execGQLHandler)
	h.Get(SchemaPath, listSchemaHandler)
	h.Post(SchemaPath, loadSchemaHandler)
	h.Patch(SchemaPath, patchSchemaHandler)
  h.Post(SchemaMigrationPath, setMigrationHandler)
	h.Get(SchemaMigrationPath, getMigrationHandler)
	h.Post(IndexPath, createIndexHandler)
	h.Delete(IndexPath, dropIndexHandler)
	h.Get(IndexPath, listIndexHandler)
	h.Get(PeerIDPath, peerIDHandler)
	h.Handle("/*", playgroundHandler)
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
