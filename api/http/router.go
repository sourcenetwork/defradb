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
	"github.com/pkg/errors"
)

const (
	// Version is the current version of the HTTP API.
	Version          string = "v0"
	versionedAPIPath string = "/api/" + Version

	RootPath    string = versionedAPIPath + ""
	PingPath    string = versionedAPIPath + "/ping"
	DumpPath    string = versionedAPIPath + "/debug/dump"
	BlocksPath  string = versionedAPIPath + "/blocks"
	GraphQLPath string = versionedAPIPath + "/graphql"
	SchemaPath  string = versionedAPIPath + "/schema"
	IndexPath   string = versionedAPIPath + "/index"
	PeerIDPath  string = versionedAPIPath + "/peerid"
)

var router = chi.NewRouter()

func init() {
	router.Get(RootPath, rootHandler)
	router.Get(PingPath, pingHandler)
	router.Get(DumpPath, dumpHandler)
	router.Get(BlocksPath+"/{cid}", getBlockHandler)
	router.Get(GraphQLPath, execGQLHandler)
	router.Post(GraphQLPath, execGQLHandler)
	router.Get(SchemaPath, listSchemaHandler)
	router.Post(SchemaPath, loadSchemaHandler)
	router.Patch(SchemaPath, patchSchemaHandler)
	router.Post(IndexPath, createIndexHandler)
	router.Delete(IndexPath, dropIndexHandler)
	router.Get(IndexPath, listIndexHandler)
	router.Get(PeerIDPath, peerIDHandler)
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
