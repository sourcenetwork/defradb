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
	"context"

	"github.com/sourcenetwork/defradb/client/options"
)

//export BasicExport
func BasicExport(
	nodePtr C.uintptr_t,
	filepath *C.char,
	collections *C.char,
	format *C.char,
	pretty C.int,
) C.Result {
	ctx := context.Background()

	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	exportOpt := options.BasicExport().
		SetCollections(splitCommaSeparatedString(C.GoString(collections))).
		SetFormat(C.GoString(format)).
		SetPretty(pretty != 0)
	err = store.BasicExport(ctx, C.GoString(filepath), exportOpt)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}
