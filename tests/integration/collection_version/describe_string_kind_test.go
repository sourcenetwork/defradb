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

package collection_version

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// The `collection describe` / GET /collections output renders each field's Kind and Typ as
// human-readable strings instead of numeric IDs (issue #4816). That string form must
// round-trip losslessly back into []client.CollectionVersion across every client type.
//
// One schema exercises every Kind encoding the renderer produces:
//   - scalar (String) and scalar-array ([Int!]) fields,
//   - a *CollectionKind relation (Dog.owner -> User), encoded as an object carrying a
//     CollectionID,
//   - a *SelfKind relation (User.boss/minion -> User), encoded as an object carrying a
//     RelativeID,
//
// each of which the unmarshal path must rebuild as its original concrete type rather than a
// NamedKind.
//
// The go client compares in-memory structs; the http and cli clients reconstruct the
// versions from the stringified describe wire form, so running this under
// DEFRA_CLIENT_TYPE=http and =cli is what actually guards the round-trip.
func TestGetCollectionVersion_RendersAndRoundTripsStringifiedKinds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						boss: User @primary
						minion: User
					}
					type Dog {
						name: String
						tags: [Int!]
						owner: User @primary
					}
				`,
			},
			// Scalar, scalar-array, and the cross-collection *CollectionKind relation.
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetCollectionName("Dog"),
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Dog",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: request.DocIDFieldName,
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "_ownerID",
								Kind:         client.FieldKind_DocID,
								Typ:          client.LWW_REGISTER,
								RelationName: immutable.Some("dog_user"),
								IsPrimary:    true,
							},
							{
								Name: "name",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								// *CollectionKind: relation to a distinct collection, encoded
								// as an object carrying the related CollectionID.
								Name:         "owner",
								Kind:         client.NewCollectionKind("bafyreicuxpdrri4wwdknhbchhdii6tu4myqlhspv3s2c3pci7jt7qc3zua", false),
								RelationName: immutable.Some("dog_user"),
								IsPrimary:    true,
							},
							{
								Name: "tags",
								Kind: client.FieldKind_INT_ARRAY,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
				},
			},
			// The *SelfKind relation (a circular relation within the same collection set).
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetCollectionName("User"),
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "User",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: request.DocIDFieldName,
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "_bossID",
								Kind:         client.FieldKind_DocID,
								Typ:          client.LWW_REGISTER,
								RelationName: immutable.Some("user_user"),
								IsPrimary:    true,
							},
							{
								Name:         "_minionID",
								Kind:         client.FieldKind_DocID,
								Typ:          client.LWW_REGISTER,
								RelationName: immutable.Some("user_user"),
							},
							{
								// *SelfKind: relation back into the same collection set,
								// encoded as an object carrying a RelativeID.
								Name:         "boss",
								Kind:         client.NewSelfKind("", false),
								RelationName: immutable.Some("user_user"),
								IsPrimary:    true,
							},
							{
								Name:         "minion",
								Kind:         client.NewSelfKind("", false),
								RelationName: immutable.Some("user_user"),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAddCollection_WithStringifiedKinds_RoundTripsToOriginalVersions(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						boss: User @primary
						minion: User
					}
					type Dog {
						name: String
						tags: [Int!]
						owner: User @primary
					}
				`,
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "User",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: request.DocIDFieldName,
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "_bossID",
								Kind:         client.FieldKind_DocID,
								Typ:          client.LWW_REGISTER,
								RelationName: immutable.Some("user_user"),
								IsPrimary:    true,
							},
							{
								Name:         "_minionID",
								Kind:         client.FieldKind_DocID,
								Typ:          client.LWW_REGISTER,
								RelationName: immutable.Some("user_user"),
							},
							{
								// *SelfKind: relation back into the same collection set,
								// encoded as an object carrying a RelativeID.
								Name:         "boss",
								Kind:         client.NewSelfKind("", false),
								RelationName: immutable.Some("user_user"),
								IsPrimary:    true,
							},
							{
								Name:         "minion",
								Kind:         client.NewSelfKind("", false),
								RelationName: immutable.Some("user_user"),
							},
						},
					},
					{
						Name:           "Dog",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: request.DocIDFieldName,
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "_ownerID",
								Kind:         client.FieldKind_DocID,
								Typ:          client.LWW_REGISTER,
								RelationName: immutable.Some("dog_user"),
								IsPrimary:    true,
							},
							{
								Name: "name",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								// *CollectionKind: relation to a distinct collection, encoded
								// as an object carrying the related CollectionID.
								Name:         "owner",
								Kind:         client.NewCollectionKind("bafyreicuxpdrri4wwdknhbchhdii6tu4myqlhspv3s2c3pci7jt7qc3zua", false),
								RelationName: immutable.Some("dog_user"),
								IsPrimary:    true,
							},
							{
								Name: "tags",
								Kind: client.FieldKind_INT_ARRAY,
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
