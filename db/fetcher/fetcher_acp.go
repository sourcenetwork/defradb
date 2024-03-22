// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package fetcher

import (
	"context"

	"github.com/sourcenetwork/defradb/acp"
	"github.com/sourcenetwork/defradb/db/permission"
)

// runDocReadPermissionCheck handles the checking (while fetching) if the document has read access
// or not, according to our access logic based on weather (1) the request is permissioned,
// (2) the collection is permissioned (has a policy), (3) acp module exists.
func (df *DocumentFetcher) runDocReadPermissionCheck(ctx context.Context) error {
	hasPermission, err := permission.CheckAccessOfDocOnCollectionWithACP(
		ctx,
		df.identity,
		df.acp,
		df.col,
		acp.ReadPermission,
		df.kv.Key.DocID,
	)

	if err != nil {
		df.passedPermissionCheck = false
		return err
	}

	df.passedPermissionCheck = hasPermission
	return nil
}
