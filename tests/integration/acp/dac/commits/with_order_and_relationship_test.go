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

package test_acp_dac_commits

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// Private doc owned by identity 1. Identity 2 is granted reader, then the
// reader relationship is revoked. After revocation identity 2 should no
// longer see any commits.
func TestACP_QueryCommitsOnPrivateDocAfterRelationshipRevocation_CanNotSeeCommits(t *testing.T) {
	uniqueCid := testUtils.NewUniqueValue()

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   usersPolicy,
			},

			&action.AddCollection{
				SDL: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,
			},

			&action.AddDoc{
				CollectionID: 0,
				Identity:     testUtils.ClientIdentity(1),
				Doc:          userDoc,
			},

			// Grant reader to identity 2.
			testUtils.AddDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2),
				CollectionID:      0,
				DocID:             0,
				Relation:          "reader",
				ExpectedExistence: false,
			},

			// Before revocation, identity 2 sees all 3 commits.
			&action.Request{
				Identity: testUtils.ClientIdentity(2),
				Request: `
					query {
						_commits {
							cid
						}
					}
				`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{"cid": uniqueCid},
						{"cid": uniqueCid},
						{"cid": uniqueCid},
					},
				},
				NonOrderedResults: true,
			},

			// Revoke reader from identity 2.
			testUtils.DeleteDACActorRelationship{
				RequestorIdentity:  testUtils.ClientIdentity(1),
				TargetIdentity:     testUtils.ClientIdentity(2),
				CollectionID:       0,
				DocID:              0,
				Relation:           "reader",
				ExpectedRecordFound: true,
			},

			// After revocation, identity 2 sees nothing.
			&action.Request{
				Identity: testUtils.ClientIdentity(2),
				Request: `
					query {
						_commits {
							cid
						}
					}
				`,
				Results: map[string]any{
					"_commits": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Private doc + an update. _commits(order: {height: DESC}) without identity
// should NOT see any commits (per DAC).
func TestACP_QueryCommitsWithOrderOnPrivateDocWithoutIdentity_CanNotSeeCommits(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   usersPolicy,
			},

			&action.AddCollection{
				SDL: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,
			},

			&action.AddDoc{
				CollectionID: 0,
				Identity:     testUtils.ClientIdentity(1),
				Doc:          userDoc,
			},

			&action.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Identity:     testUtils.ClientIdentity(1),
				Doc: `
					{
						"age": 29
					}
				`,
			},

			&action.Request{
				Request: `
					query {
						_commits(order: {height: DESC}) {
							cid
						}
					}
				`,
				Results: map[string]any{
					"_commits": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Private doc + an update. _commits(order: {height: DESC}) with the owner
// identity should return all 5 commits in descending height order.
func TestACP_QueryCommitsWithOrderOnPrivateDocWithOwnerIdentity_CanSeeCommits(t *testing.T) {
	uniqueCid := testUtils.NewUniqueValue()

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   usersPolicy,
			},

			&action.AddCollection{
				SDL: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,
			},

			&action.AddDoc{
				CollectionID: 0,
				Identity:     testUtils.ClientIdentity(1),
				Doc:          userDoc,
			},

			&action.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Identity:     testUtils.ClientIdentity(1),
				Doc: `
					{
						"age": 29
					}
				`,
			},

			&action.Request{
				Identity: testUtils.ClientIdentity(1),
				Request: `
					query {
						_commits(order: {height: DESC}) {
							cid
							height
						}
					}
				`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{"cid": uniqueCid, "height": 2},
						{"cid": uniqueCid, "height": 2},
						{"cid": uniqueCid, "height": 1},
						{"cid": uniqueCid, "height": 1},
						{"cid": uniqueCid, "height": 1},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
