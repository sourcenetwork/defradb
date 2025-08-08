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

import (
	"context"
	"fmt"
)

func AddSchema(n int, newSchema string, txnID uint64) GoCResult {
	ctx := context.Background()

	ctx, err := contextWithTransaction(n, ctx, txnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	collectionVersions, err := GetNode(n).DB.AddSchema(ctx, newSchema)
	if err != nil {
		return returnGoC(1, fmt.Sprintf(errAddingSchema, err), "")
	}
	return marshalJSONToGoCResult(collectionVersions)
}
