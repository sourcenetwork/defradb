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

	"github.com/sourcenetwork/defradb/acp/identity"
	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/internal/db/permission"
)

// registerDocWithACP handles the registration of the document with acp.
// The registering is done at document creation on the collection.
//
// According to our access logic we have these components to worry about:
// (1) the request is permissioned (has an identity signature),
// (2) the collection is permissioned (has a policy),
// (3) acp is available (acp is enabled).
//
// The document is only registered if all (1) (2) and (3) are true.
//
// Otherwise, nothing is registered with the acp system.
func (c *collection) registerDocWithACP(
	ctx context.Context,
	docID string,
) error {
	// If document acp is not available, then no document is registered.
	if !c.db.documentACP.HasValue() {
		return nil
	}
	return permission.RegisterDocOnCollectionWithDocumentACP(
		ctx,
		identity.FromContext(ctx),
		c.db.documentACP.Value(),
		c,
		docID,
	)
}

func (c *collection) checkAccessOfDocWithACP(
	ctx context.Context,
	resourcePermission acpTypes.ResourceInterfacePermission,
	docID string,
) (bool, error) {
	// If document acp is not available, then we have unrestricted access.
	if !c.db.documentACP.HasValue() {
		return true, nil
	}
	ident := identity.FromContext(ctx)
	if ident.HasValue() && c.db.nodeIdentity.HasValue() && ident.Value().DID == c.db.nodeIdentity.Value().DID {
		return true, nil
	}
	return permission.CheckAccessOfDocOnCollectionWithACP(
		ctx,
		identity.FromContext(ctx),
		c.db.documentACP.Value(),
		c,
		resourcePermission,
		docID,
	)
}
