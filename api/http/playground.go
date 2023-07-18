// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build playground

package http

import (
	"io/fs"
	"net/http"

	"github.com/sourcenetwork/defradb/playground"
)

func init() {
	sub, err := fs.Sub(playground.Dist, "dist")
	if err != nil {
		panic(err)
	}
	router.Handle("/*", http.FileServer(http.FS(sub)))
}
