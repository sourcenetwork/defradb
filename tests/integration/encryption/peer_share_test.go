// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package encryption

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestDocEncryptionPeer_IfDocIsPublic_ShouldFetchKeyAndDecrypt(t *testing.T) {
	test := testUtils.TestCase{
		KMS: testUtils.KMS{Activated: true},
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.CreateDoc{
				NodeID:         immutable.Some(0),
				Doc:            john21Doc,
				IsDocEncrypted: true,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					User {
						age
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"age": int64(21),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryptionPeer_IfPublicDocHasEncryptedField_ShouldFetchKeyAndDecrypt(t *testing.T) {
	test := testUtils.TestCase{
		KMS: testUtils.KMS{Activated: true},
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.CreateDoc{
				NodeID:          immutable.Some(0),
				Doc:             john21Doc,
				EncryptedFields: []string{"age"},
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					User {
						name
						age
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "John",
							"age":  int64(21),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryptionPeer_IfEncryptedPublicDocHasEncryptedField_ShouldFetchKeysAndDecrypt(t *testing.T) {
	test := testUtils.TestCase{
		KMS: testUtils.KMS{Activated: true},
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.CreateDoc{
				NodeID:          immutable.Some(0),
				Doc:             john21Doc,
				IsDocEncrypted:  true,
				EncryptedFields: []string{"age"},
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					User {
						name
						age
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "John",
							"age":  int64(21),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryptionPeer_IfAllFieldsOfEncryptedPublicDocAreIndividuallyEncrypted_ShouldFetchKeysAndDecrypt(t *testing.T) {
	test := testUtils.TestCase{
		KMS: testUtils.KMS{Activated: true},
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.CreateDoc{
				NodeID:          immutable.Some(0),
				Doc:             john21Doc,
				IsDocEncrypted:  true,
				EncryptedFields: []string{"name", "age"},
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					User {
						name
						age
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "John",
							"age":  int64(21),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryptionPeer_IfAllFieldsOfPublicDocAreIndividuallyEncrypted_ShouldFetchKeysAndDecrypt(t *testing.T) {
	test := testUtils.TestCase{
		KMS: testUtils.KMS{Activated: true},
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.CreateDoc{
				NodeID:          immutable.Some(0),
				Doc:             john21Doc,
				EncryptedFields: []string{"name", "age"},
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					User {
						name
						age
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "John",
							"age":  int64(21),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryptionPeer_WithUpdatesOnEncryptedDeltaBasedCRDTField_ShouldDecryptAndCorrectlyMerge(t *testing.T) {
	test := testUtils.TestCase{
		KMS: testUtils.KMS{Activated: true},
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						age: Int @crdt(type: pcounter)
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.CreateDoc{
				NodeID:          immutable.Some(0),
				Doc:             john21Doc,
				EncryptedFields: []string{"age"},
			},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(0),
				Doc:    `{"age": 3}`,
			},
			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(0),
				Doc:    `{"age": 2}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					User {
						name
						age
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "John",
							"age":  int64(26),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryptionPeer_WithUpdatesOnDeltaBasedCRDTFieldOfEncryptedDoc_ShouldDecryptAndCorrectlyMerge(t *testing.T) {
	test := testUtils.TestCase{
		KMS: testUtils.KMS{Activated: true},
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						age: Int @crdt(type: pcounter)
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.CreateDoc{
				NodeID:         immutable.Some(0),
				Doc:            john21Doc,
				IsDocEncrypted: true,
			},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(0),
				Doc:    `{"age": 3}`,
			},
			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(0),
				Doc:    `{"age": 2}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					User {
						name
						age
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "John",
							"age":  int64(26),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryptionPeer_WithUpdatesThatSetsEmptyString_ShouldDecryptAndCorrectlyMerge(t *testing.T) {
	test := testUtils.TestCase{
		KMS: testUtils.KMS{Activated: true},
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						age: Int 
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.CreateDoc{
				NodeID:         immutable.Some(0),
				Doc:            john21Doc,
				IsDocEncrypted: true,
			},
			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(0),
				Doc:    `{"name": ""}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					User {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": ""},
					},
				},
			},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(0),
				Doc:    `{"name": "John"}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					User {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryptionPeer_WithUpdatesThatSetsStringToNull_ShouldDecryptAndCorrectlyMerge(t *testing.T) {
	test := testUtils.TestCase{
		KMS: testUtils.KMS{Activated: true},
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						age: Int 
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.CreateDoc{
				NodeID:         immutable.Some(0),
				Doc:            john21Doc,
				IsDocEncrypted: true,
			},
			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(0),
				Doc:    `{"name": null}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					User {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": nil},
					},
				},
			},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(0),
				Doc:    `{"name": "John"}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					User {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
