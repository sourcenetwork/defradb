// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package index

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// TestQueryWithIndex_FilterOnIndexedFieldAndRelation_ShouldAND tests that when filtering
// by both an indexed field on the root type AND fields on a related type, the filters are
// combined with AND semantics, not OR.
func TestQueryWithIndex_FilterOnIndexedFieldAndRelation_ShouldAND(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						publicID: String @index(unique: true)
						isEnabled: Boolean
						passkeys: [PasskeyCredential]
					}
					
					type PasskeyCredential {
						credentialID: String @index(unique: true)
						user: User
					}
				`,
			},
			// Create a user
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"publicID": "user-abc-123",
					"isEnabled": true
				}`,
			},
			// Create 3 passkey credentials for the same user
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"credentialID": "credential-1",
					"user":         testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"credentialID": "credential-2",
					"user":         testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"credentialID": "credential-3",
					"user":         testUtils.NewDocIndex(0, 0),
				},
			},
			// Query for a specific credentialID AND user conditions
			// Should return only 1 result (credential-2), not all 3
			testUtils.Request{
				Request: `query {
					PasskeyCredential(filter: {
						credentialID: {_eq: "credential-2"},
						user: {
							isEnabled: {_eq: true},
							publicID: {_eq: "user-abc-123"}
						}
					}) {
						credentialID
						user {
							publicID
						}
					}
				}`,
				Results: map[string]any{
					"PasskeyCredential": []map[string]any{
						{
							"credentialID": "credential-2",
							"user": map[string]any{
								"publicID": "user-abc-123",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// TestQueryWithIndex_FilterOnIndexedFieldAndRelation_NoMatch tests that when conditions
// don't match on both sides, no documents are returned.
func TestQueryWithIndex_FilterOnIndexedFieldAndRelation_NoMatch(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						publicID: String @index(unique: true)
						isEnabled: Boolean
						passkeys: [PasskeyCredential]
					}
					
					type PasskeyCredential {
						credentialID: String @index(unique: true)
						user: User
					}
				`,
			},
			// Create a user
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"publicID": "user-abc-123",
					"isEnabled": true
				}`,
			},
			// Create passkey credential
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"credentialID": "credential-1",
					"user":         testUtils.NewDocIndex(0, 0),
				},
			},
			// Query for non-existent credentialID but matching user
			// Should return 0 results (AND semantics)
			// If OR semantics are used, it would incorrectly return 1 result
			testUtils.Request{
				Request: `query {
					PasskeyCredential(filter: {
						credentialID: {_eq: "non-existent"},
						user: {
							isEnabled: {_eq: true},
							publicID: {_eq: "user-abc-123"}
						}
					}) {
						credentialID
					}
				}`,
				Results: map[string]any{
					"PasskeyCredential": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// TestQueryWithIndex_FilterOnIndexedFieldAndRelation_WrongUser tests that when the
// user conditions don't match, no documents are returned even if credentialID matches.
func TestQueryWithIndex_FilterOnIndexedFieldAndRelation_WrongUser(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						publicID: String @index(unique: true)
						isEnabled: Boolean
						passkeys: [PasskeyCredential]
					}
					
					type PasskeyCredential {
						credentialID: String @index(unique: true)
						user: User
					}
				`,
			},
			// Create two users
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"publicID": "user-1",
					"isEnabled": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"publicID": "user-2",
					"isEnabled": true
				}`,
			},
			// Create credential for user-1
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"credentialID": "credential-1",
					"user":         testUtils.NewDocIndex(0, 0),
				},
			},
			// Query for credential-1 but with user-2's publicID
			// Should return 0 results (AND semantics)
			testUtils.Request{
				Request: `query {
					PasskeyCredential(filter: {
						credentialID: {_eq: "credential-1"},
						user: {
							publicID: {_eq: "user-2"}
						}
					}) {
						credentialID
					}
				}`,
				Results: map[string]any{
					"PasskeyCredential": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
