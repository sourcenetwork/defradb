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

package crdt

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// A field's Typ (CRDT type) may be supplied to a collection patch as its string form (e.g.
// "pncounter") as well as the numeric form, mirroring how Kind already accepts strings. This
// is the read side of the same CType.UnmarshalJSON that lets the describe output round-trip.
// The stored field must end up with the corresponding numeric CType.
func TestCollectionVersionUpdates_AddFieldCRDTTypAsString_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "foo", "Kind": "Int", "Typ": "pncounter"} }
					]
				`,
			},
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetCollectionName("Users"),
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsActive:       true,
						IsMaterialized: true,
						PreviousVersion: immutable.Some(client.CollectionSource{
							SourceCollectionID: "bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu",
						}),
						Fields: []client.CollectionFieldDescription{
							{
								Name: request.DocIDFieldName,
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "name",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "foo",
								Kind: client.FieldKind_NILLABLE_INT,
								Typ:  client.PN_COUNTER,
							},
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

// An unknown Typ string is rejected rather than silently mis-parsed.
func TestCollectionVersionUpdates_AddFieldCRDTTypAsUnknownString_Error(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "foo", "Kind": "Int", "Typ": "notacrdt"} }
					]
				`,
				ExpectedError: "unknown crdt string representation",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
