// Copyright 2025 Democratized Data Foundation
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

	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	acpDB "github.com/sourcenetwork/defradb/internal/db/acp"
	"github.com/sourcenetwork/defradb/internal/identity"
)

// registerDoc handles the registration of the document with acp.
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
func (c *collection) registerDoc(
	ctx context.Context,
	docID string,
) error {
	// If document acp is not available, then no document is registered.
	if !c.db.documentACP.HasValue() {
		return nil
	}
	return acpDB.RegisterDocOnCollectionWithDocumentACP(
		ctx,
		identity.FromContext(ctx),
		c.db.documentACP.Value(),
		c,
		docID,
	)
}

// registerCollection handles the registration of the collection itself with document acp.
//
// This is only relevant for branchable collections, which maintain a collection-level commit DAG.
// Registering the collection as an acp object lets us gate read access to that collection-level
// commit DAG, the same way document registration gates access to a document.
//
// The registration is a no-op unless document acp is available, the collection is branchable, the
// collection has a policy, and the request carries an identity (see
// [acpDB.RegisterCollectionObject]).
func (c *collection) registerCollection(
	ctx context.Context,
) error {
	if !c.db.documentACP.HasValue() {
		return nil
	}
	return acpDB.RegisterCollectionObject(
		ctx,
		identity.FromContext(ctx),
		c.db.documentACP.Value(),
		c,
	)
}

func (c *collection) checkAccessOfDoc(
	ctx context.Context,
	resourcePermission acpTypes.ResourceInterfacePermission,
	docID string,
) (bool, error) {
	// If document acp is not available, then we have unrestricted access.
	if !c.db.documentACP.HasValue() {
		return true, nil
	}
	ident := identity.FromContext(ctx)
	if ident.HasValue() && c.db.nodeIdentity.HasValue() && ident.Value().DID() == c.db.nodeIdentity.Value().DID() {
		return true, nil
	}
	return acpDB.CheckAccessOfDocOnCollection(
		ctx,
		ident,
		c.db.nodeACP,
		c.db.documentACP.Value(),
		c,
		resourcePermission,
		docID,
	)
}
