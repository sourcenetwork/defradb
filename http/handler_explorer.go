// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build explorer

package http

import (
	"io/fs"
	"net/http"

	"github.com/sourcenetwork/defradb/explorer"
)

func init() {
	sub, err := fs.Sub(explorer.Dist, "dist")
	if err != nil {
		panic(err)
	}
	explorerHandler = http.FileServer(http.FS(sub))
	ExplorerEnabled = true
}
