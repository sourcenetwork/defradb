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

	"github.com/sourcenetwork/defradb/client"
)

// tryRegisterDocWithACP handles the registeration of the document with acp module,
// according to our registration logic based on weather (1) the request is permissioned,
// (2) the collection is permissioned (has a policy), (3) acp module exists.
//
// Note: we only register the document with ACP if all (1) (2) and (3) are true.
// In all other cases, nothing is registered with ACP.
//
// Moreover 8 states, upon document creation:
// - (SignatureRequest, PermissionedCollection, ModuleExists)    => Register with ACP
// - (SignatureRequest, PermissionedCollection, !ModuleExists)   => Normal/Public - Don't Register with ACP
// - (SignatureRequest, !PermissionedCollection, ModuleExists)   => Normal/Public - Don't Register with ACP
// - (SignatureRequest, !PermissionedCollection, !ModuleExists)  => Normal/Public - Don't Register with ACP
// - (!SignatureRequest, PermissionedCollection, ModuleExists)   => Normal/Public - Don't Register with ACP
// - (!SignatureRequest, !PermissionedCollection, ModuleExists)  => Normal/Public - Don't Register with ACP
// - (!SignatureRequest, PermissionedCollection, !ModuleExists)  => Normal/Public - Don't Register with ACP
// - (!SignatureRequest, !PermissionedCollection, !ModuleExists) => Normal/Public - Don't Register with ACP
func (c *collection) tryRegisterDocWithACP(ctx context.Context, doc *client.Document) error {
	// Check if acp module exists.
	if c.db.ACPModule().HasValue() {
		// Check if collection has policy.
		if policyID, resourceName, hasPolicy := client.IsPermissioned(c); hasPolicy {
			return c.db.ACPModule().Value().RegisterDocCreation(
				ctx,
				"cosmos1zzg43wdrhmmk89z3pmejwete2kkd4a3vn7w969", // TODO-ACP: Replace with signature identity
				policyID,
				resourceName,
				doc.ID().String(),
			)
		}
	}

	return nil
}
