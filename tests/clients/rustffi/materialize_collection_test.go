// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

//go:build rust_ffi

package rustffi

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/lenses"
	lensmodel "github.com/sourcenetwork/lens/host-go/config/model"
)

func TestMaterializeCollectionEagerlyMigratesAndCachesDocuments(t *testing.T) {
	Init()

	node, err := NewNode(NodeOptions{InMemory: true})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, node.Close())
	})

	initialJSON, err := node.AddSchema("", `type Users { name: String }`)
	require.NoError(t, err)

	var initialVersions []client.CollectionVersion
	require.NoError(t, json.Unmarshal([]byte(initialJSON), &initialVersions))
	require.Len(t, initialVersions, 1)

	_, err = node.ExecRequest(
		"",
		`mutation { create_Users(input: {name: "John"}) { _docID } }`,
		"",
		"",
	)
	require.NoError(t, err)

	patchedJSON, err := node.PatchCollection(
		"",
		"Users",
		`[{"op":"add","path":"/Fields/-","value":{"Name":"verified","Kind":"Boolean"}}]`,
	)
	require.NoError(t, err)

	var patchedVersion client.CollectionVersion
	require.NoError(t, json.Unmarshal([]byte(patchedJSON), &patchedVersion))

	configJSON, err := json.Marshal(client.LensConfig{
		SourceCollectionVersionID:      initialVersions[0].VersionID,
		DestinationCollectionVersionID: patchedVersion.VersionID,
		Lens: lensmodel.Lens{
			Lenses: []lensmodel.LensModule{
				{
					Path: lenses.SetDefaultModulePath,
					Arguments: map[string]any{
						"dst":   "verified",
						"value": true,
					},
				},
			},
		},
	})
	require.NoError(t, err)
	_, err = node.SetMigration("", string(configJSON))
	require.NoError(t, err)

	count, err := node.MaterializeCollection("", "Users")
	require.NoError(t, err)
	require.Equal(t, 1, count)

	count, err = node.MaterializeCollection("", "Users")
	require.NoError(t, err)
	require.Zero(t, count)

	result, err := node.ExecRequest("", `{ Users { name verified } }`, "", "")
	require.NoError(t, err)
	require.JSONEq(t, `{"data":{"Users":[{"name":"John","verified":true}]}}`, result)
}
