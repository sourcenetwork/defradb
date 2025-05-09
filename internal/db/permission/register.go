// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package permission

import (
	"context"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/dac"
	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
)

// RegisterDocOnCollectionWithDocumentACP handles the registration of the document with document acp system.
//
// Since document acp will always exist when this is called we have these components to worry about:
// (1) the request is permissioned (has an identity signature),
// (2) the collection is permissioned (has a policy),
//
// The document is only registered if all (1) (2) are true.
//
// Otherwise, nothing is registered with document acp.
func RegisterDocOnCollectionWithDocumentACP(
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	documentACP dac.DocumentACP,
	collection client.Collection,
	docID string,
) error {
	// An identity exists and the collection has a policy.
	if policyID, resourceName, hasPolicy := IsPermissioned(collection); hasPolicy && identity.HasValue() {
		return documentACP.RegisterDocObject(
			ctx,
			identity.Value(),
			policyID,
			resourceName,
			docID,
		)
	}

	return nil
}
