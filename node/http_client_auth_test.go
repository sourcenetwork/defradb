// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build !js

package node

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/immutable"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/crypto"
	defrahttp "github.com/sourcenetwork/defradb/http"
	iIdentity "github.com/sourcenetwork/defradb/internal/identity"
)

func TestHTTPClientDefaultIdentityAuthorizesNodeACPRequest(t *testing.T) {
	ctx := context.Background()
	ident, err := acpIdentity.Generate(crypto.KeyTypeEd25519)
	require.NoError(t, err)

	n, err := New(
		ctx,
		options.Node().
			SetDisableP2P(true).
			Store().SetType(options.NodeMemoryStore).
			Node().
			DB().SetNodeIdentity(ident).
			Node().
			NodeACP().SetEnabled(true).
			Node().
			HTTP().SetAddress("127.0.0.1:0").
			Node(),
	)
	require.NoError(t, err)
	startCtx := iIdentity.WithContext(ctx, immutable.Some[acpIdentity.Identity](ident))
	require.NoError(t, n.Start(startCtx))
	t.Cleanup(func() {
		assert.NoError(t, n.Close(context.Background()))
	})

	unauthenticated, err := defrahttp.NewClient(n.APIURL)
	require.NoError(t, err)
	_, err = unauthenticated.GetNACStatus(ctx)
	require.Error(t, err)

	audience, err := defrahttp.AuthAudienceForURL(n.APIURL)
	require.NoError(t, err)
	token, err := n.DB.GetNodeIdentityToken(ctx, immutable.Some(audience))
	require.NoError(t, err)
	tokenIdentity, err := acpIdentity.FromToken(token)
	require.NoError(t, err)

	authenticated, err := defrahttp.NewClient(n.APIURL, defrahttp.WithIdentity(tokenIdentity))
	require.NoError(t, err)
	status, err := authenticated.GetNACStatus(ctx)
	require.NoError(t, err)
	assert.Equal(t, client.NACEnabled.String(), status.Status)
}
