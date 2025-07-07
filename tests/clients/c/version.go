// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cwrap

/*
#include "defra_structs.h"
*/
import "C"

import (
	"github.com/sourcenetwork/defradb/version"
)

func VersionGet(cFlagFull C.int, cFlagJSON C.int) *C.Result {
	flagFull := cFlagFull != 0
	flagJSON := cFlagJSON != 0

	// Call the version function
	dv, err := version.NewDefraVersion()
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	// Return either the JSON, the long string version, or the short string version
	if flagJSON {
		return marshalJSONToCResult(dv)
	}
	if flagFull {
		return returnC(0, "", dv.StringFull())
	}
	return returnC(0, "", dv.String())
}
