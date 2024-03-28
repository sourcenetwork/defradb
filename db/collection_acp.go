// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"context"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp"
	"github.com/sourcenetwork/defradb/db/permission"
)

func (c *collection) registerDocCreation(
	ctx context.Context,
	identity immutable.Option[string],
	docID string,
) error {
	return permission.RegisterDocCreationOnCollection(
		ctx,
		identity,
		c.db.ACPModule(),
		c,
		docID,
	)
}

func (c *collection) checkDocPermissionedAccess(
	ctx context.Context,
	identity immutable.Option[string],
	dpiPermission acp.DPIPermission,
	docID string,
) (bool, error) {
	return permission.CheckDocPermissionedAccessOnCollection(
		ctx,
		identity,
		c.db.ACPModule(),
		c,
		dpiPermission,
		docID,
	)
}
