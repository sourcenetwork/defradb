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
	"github.com/sourcenetwork/defradb/tests/multiplier"
)

// Test that a query with multiple CIDs returns only those the identity has permission to view.
func TestACP_QueryCommitsWithMultipleCIDsAcrossPrivateDocs_CanOnlySeeOwnedCommit(t *testing.T) {
	test := testUtils.TestCase{
		// CIDs change under encryption/signing.
		MultiplierExcludes: []string{multiplier.SignedDocs, multiplier.EncryptedDocs},
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

			// doc 0 - private, owned by identity 1.
			&action.AddDoc{
				CollectionID: 0,
				Identity:     testUtils.ClientIdentity(1),
				Doc:          userDoc,
			},

			// doc 1 - private, owned by identity 2.
			&action.AddDoc{
				CollectionID: 0,
				Identity:     testUtils.ClientIdentity(2),
				Doc:          secondUserDoc,
			},

			// Query both CIDs. Only the one owned by identity 1 should be returned.
			&action.Request{
				Identity: testUtils.ClientIdentity(1),
				Request: `query {
					_commits(
						cid: ["` + userDocCompositeCid + `", "` + secondUserDocCompositeCid + `"]
					) {
						cid
					}
				}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{"cid": userDocCompositeCid},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Test that querying multiple CIDs without an identity returns nothing for private docs.
func TestACP_QueryCommitsWithMultipleCIDsAcrossPrivateDocs_WithoutIdentity_CanNotSeeCommits(t *testing.T) {
	test := testUtils.TestCase{
		MultiplierExcludes: []string{multiplier.SignedDocs, multiplier.EncryptedDocs},
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

			&action.AddDoc{
				CollectionID: 0,
				Identity:     testUtils.ClientIdentity(2),
				Doc:          secondUserDoc,
			},

			// Unsigned request returns nothing.
			&action.Request{
				Request: `query {
					_commits(
						cid: ["` + userDocCompositeCid + `", "` + secondUserDocCompositeCid + `"]
					) {
						cid
					}
				}`,
				Results: map[string]any{
					"_commits": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Test that docID filter silently excludes CIDs from other documents in a multi-CID request.
func TestACP_QueryCommitsWithDocIDAndMultipleCIDs_SecondCIDFromOtherDoc_IsFiltered(t *testing.T) {
	test := testUtils.TestCase{
		MultiplierExcludes: []string{multiplier.SignedDocs, multiplier.EncryptedDocs},
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

			// doc 0 - owned by identity 1.
			&action.AddDoc{
				CollectionID: 0,
				Identity:     testUtils.ClientIdentity(1),
				Doc:          userDoc,
			},

			// doc 1 - owned by identity 2.
			&action.AddDoc{
				CollectionID: 0,
				Identity:     testUtils.ClientIdentity(2),
				Doc:          secondUserDoc,
			},

			// Grant identity 1 read access to doc1.
			testUtils.AddDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(2),
				TargetIdentity:    testUtils.ClientIdentity(1),
				CollectionID:      0,
				DocID:             1,
				Relation:          "reader",
				ExpectedExistence: false,
			},

			// Only doc0 CID is returned, doc1 CID is filtered out by docID constraint.
			&action.Request{
				Identity: testUtils.ClientIdentity(1),
				Request: `query {
					_commits(
						docID: "` + userDocID + `",
						cid: ["` + userDocCompositeCid + `", "` + secondUserDocCompositeCid + `"]
					) {
						cid
					}
				}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{"cid": userDocCompositeCid},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Test that a single CID not matching the docID filter returns an error.
func TestACP_QueryCommitsWithDocIDAndCID_CIDFromOtherDoc_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		MultiplierExcludes: []string{multiplier.SignedDocs, multiplier.EncryptedDocs},
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

			// doc 0 - owned by identity 1.
			&action.AddDoc{
				CollectionID: 0,
				Identity:     testUtils.ClientIdentity(1),
				Doc:          userDoc,
			},

			// doc 1 - owned by identity 2.
			&action.AddDoc{
				CollectionID: 0,
				Identity:     testUtils.ClientIdentity(2),
				Doc:          secondUserDoc,
			},

			// Grant identity 1 read access to doc1.
			testUtils.AddDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(2),
				TargetIdentity:    testUtils.ClientIdentity(1),
				CollectionID:      0,
				DocID:             1,
				Relation:          "reader",
				ExpectedExistence: false,
			},

			// Fails because the CID belongs to a different document.
			&action.Request{
				Identity: testUtils.ClientIdentity(1),
				Request: `query {
					_commits(
						docID: "` + userDocID + `",
						cid: "` + secondUserDocCompositeCid + `"
					) {
						cid
					}
				}`,
				Results: map[string]any{
					"_commits": []map[string]any{},
				},
				ExpectedError: "cid either does not exist or belong to document",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
