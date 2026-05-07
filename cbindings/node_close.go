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
	"runtime/cgo"
)

//export CloseNode
func CloseNode(nodePtr C.uintptr_t) C.Result {
	node, err := getNodeFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	err = node.Close(context.Background())
	if err != nil {
		return returnC(GoCResult{1, err.Error(), ""})
	}
	cgo.Handle(nodePtr).Delete()
	return returnC(GoCResult{0, "", ""})
}
