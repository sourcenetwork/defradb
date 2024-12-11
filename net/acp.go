// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package net

import (
	"context"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp"
	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
)

type ACP interface {
	acp.ACP
	// GetCollections returns the list of collections according to the given options.
	GetCollections(ctx context.Context, opts client.CollectionFetchOptions) ([]client.Collection, error)
	// GetIndentityToken returns an identity token for the given audience.
	GetIdentityToken(ctx context.Context, audience immutable.Option[string]) ([]byte, error)
	// GetNodeIdentity returns the node's public raw identity.
	GetNodeIdentity(ctx context.Context) (immutable.Option[identity.PublicRawIdentity], error)
}
