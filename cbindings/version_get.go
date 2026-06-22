// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cbindings

/*
#include <stdlib.h>
#include "defra_structs.h"
*/
import "C"

import (
	"github.com/sourcenetwork/defradb/version"
)

//export GetVersion
func GetVersion(flagFull C.int, flagJSON C.int) C.Result {
	dv, err := version.NewDefraVersion()
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	if flagJSON != 0 {
		return returnC(marshalJSONToGoCResult(dv))
	}
	if flagFull != 0 {
		return returnC(returnGoC(0, "", dv.StringFull()))
	}
	return returnC(returnGoC(0, "", dv.String()))
}
