// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package api

import (
	"github.com/sourcenetwork/defradb/api/http"
	"github.com/sourcenetwork/defradb/client"
)

// NewHTTPServer returns a Server loaded with the HTTP API Handler.
func NewHTTPServer(db client.DB, c ...*http.HandlerConfig) *http.Server {
	s := http.NewServer()

	if len(c) > 0 {
		s.Handler = http.NewHandler(db, c[0])
	}

	s.Handler = http.NewHandler(db, nil)

	return s
}
