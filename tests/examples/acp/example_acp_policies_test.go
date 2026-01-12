// Copyright 2023 Democratized Data Foundation
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
	"os"
	"path/filepath"
	"testing"

	prototypes "github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/acp/local"
	acp_types "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/crypto"
)

var examplePolicyRelativeDir = "../../../examples/policy"

func Test_ExamplePolicies_PolicyIsValid(t *testing.T) {
	acp, err := local.NewLocalACP(t.TempDir(), "test")
	require.NoError(t, err)

	ctx := context.Background()
	err = acp.Start(ctx)
	require.NoError(t, err)
	defer acp.Close()

	id, err := identity.Generate(crypto.KeyTypeEd25519)
	require.NoError(t, err)

	entries, err := os.ReadDir(examplePolicyRelativeDir)
	require.NoError(t, err)

	for _, entry := range entries {
		fileName := entry.Name()
		path := filepath.Join(examplePolicyRelativeDir, fileName)
		policy, err := os.ReadFile(path)
		require.NoError(t, err)
		t.Run(fileName, func(t *testing.T) {
			_, err := acp.AddPolicy(ctx,
				id,
				string(policy),
				acp_types.PolicyMarshalType_YAML,
				prototypes.TimestampNow())
			require.NoError(t, err)
		})
	}
}
