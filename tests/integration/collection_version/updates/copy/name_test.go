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

package copy

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestColVersionUpdateCopyName(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						fullName: String
					}
				`,
			},
			&action.AddCollection{
				SDL: `
					type Books {}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Books/Fields/-", "value": {"Name": "name", "Kind": "String"} },
						{
							"op": "copy",
							"from": "/Users/Fields/1/Name",
							"path": "/Books/Fields/1/Name"
						},
						{ "op": "replace", "path": "/Users/IsActive", "value": false }
					]
				`,
			},
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetCollectionName("Books"),
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Books",
						IsActive:       true,
						IsMaterialized: true,
						PreviousVersion: immutable.Some(client.CollectionSource{
							SourceCollectionID: "bafyreia5nnmpybnn3yodhafe5cydbov3zhhyaz2ovez35uweglxkp4kopm",
						}),
						Fields: []client.CollectionFieldDescription{
							{
								Name: request.DocIDFieldName,
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "fullName",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
