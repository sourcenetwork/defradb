// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cbindings

import "C"

import (
	"github.com/sourcenetwork/defradb/version"
)

func VersionGet(flagFull bool, flagJSON bool) GoCResult {
	// Call the version function
	dv, err := version.NewDefraVersion()
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	// Return either the JSON, the long string version, or the short string version
	if flagJSON {
		return marshalJSONToGoCResult(dv)
	}
	if flagFull {
		return returnGoC(0, "", dv.StringFull())
	}
	return returnGoC(0, "", dv.String())
}
