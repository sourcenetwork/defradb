// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package acp

import (
	"context"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/dac"
	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
)

// RegisterObject registers an object with the document acp system so that read access to it can be
// gated. The object is either a document (id = its docID) or a branchable collection's
// collection-level commit DAG (id = the collection id).
//
// The object is only registered if both of the following are true:
// (1) the collection is permissioned (has a policy),
// (2) the request is permissioned (has an identity, which becomes the object owner).
//
// Otherwise, nothing is registered with document acp and the object remains public.
func RegisterObject(
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	documentACP dac.DocumentACP,
	collection client.Collection,
	id string,
) error {
	// An identity exists and the collection has a policy.
	if policyID, resourceName, hasPolicy := IsPermissioned(collection); hasPolicy && identity.HasValue() {
		return documentACP.RegisterDocObject(
			ctx,
			identity.Value(),
			policyID,
			resourceName,
			id,
		)
	}

	return nil
}
