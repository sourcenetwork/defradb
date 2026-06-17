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

// The `collection describe` / GET /collections output renders Kind and Typ as
// human-readable strings (see issue #4816). The string form must round-trip back into
// []client.CollectionVersion across all client types (go/http/cli). This test exercises
// that path for a scalar field, a scalar-array field, and the relation kinds, by
// asserting the rebuilt structs match through the GetCollections action.
//
// The go client reads in-memory structs; the http and cli clients reconstruct the
// versions from the stringified describe wire form, so running this under
// DEFRA_CLIENT_TYPE=http and =cli is what guards the round-trip.
func TestGetCollectionVersion_RoundTripsStringifiedScalarAndArrayKinds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Book {
						title: String
						pages: [Int!]
					}
				`,
			},
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetCollectionName("Book"),
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Book",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: request.DocIDFieldName,
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "pages",
								Kind: client.FieldKind_INT_ARRAY,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "title",
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

// TestGetCollectionVersion_RoundTripsStringifiedCollectionKind asserts that a
// cross-collection relation (a *CollectionKind, as opposed to the *SelfKind exercised
// below) keeps its object shape in the describe output and rebuilds as a *CollectionKind
// (NOT a NamedKind) after the round-trip. The SelfKind case below covers circular
// relations; this covers the distinct-collection case, which encodes a non-empty
// CollectionID rather than a RelativeID.
//
// As above, the http and cli clients reconstruct the version from the stringified
// describe wire form, so running this under DEFRA_CLIENT_TYPE=http and =cli is what
// guards the round-trip.
func TestGetCollectionVersion_RoundTripsStringifiedCollectionKind(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						dogs: [Dog]
					}
					type Dog {
						name: String
						owner: User @primary
					}
				`,
			},
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
								Name:         "owner",
								Kind:         client.NewCollectionKind("bafyreibhpgygzsmki22sql5ejzcojrrxbc5iuhpydhdzxul5w2znc7zrgu", false),
								RelationName: immutable.Some("dog_user"),
								IsPrimary:    true,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// TestGetCollectionVersion_RoundTripsStringifiedRelationKinds asserts that relation kinds
// keep their object shape in the describe output and rebuild as the correct
// *SelfKind/*CollectionKind (NOT a NamedKind) after the round-trip.
func TestGetCollectionVersion_RoundTripsStringifiedRelationKinds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						boss: User @primary
						minion: User
					}
				`,
			},
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
