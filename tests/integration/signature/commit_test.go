// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package signature

import (
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/onsi/gomega"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/crypto"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	corecrdt "github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func makeFieldBlock(fieldName string, value any) coreblock.Block {
	const docID = "bae-0b2f15e5-bfe7-5cb7-8045-471318d7dbc3"
	const schemaVersionID = "bafkreihhd6bqrjhl5zidwztgxzeseveplv3cj3fwtn3unjkdx7j2vr2vrq"

	fieldVal, err := cbor.Marshal(value)
	if err != nil {
		panic("failed to marshal field value")
	}

	delta := &corecrdt.LWWDelta{
		Data:            fieldVal,
		DocID:           []byte(docID),
		FieldName:       fieldName,
		SchemaVersionID: schemaVersionID,
		Priority:        1,
	}

	block := coreblock.New(delta, nil)
	return *block
}

func TestSignature_WithCommitQuery_ShouldIncludeSignatureData(t *testing.T) {
	sameIdentity := testUtils.NewSameValue()

	test := testUtils.TestCase{
		EnableSigning: true,
		SupportedClientTypes: immutable.Some([]state.ClientType{
			// C bindings do not support calling functions with non-Secp256k key yet
			testUtils.GoClientType,
			testUtils.CLIClientType,
			testUtils.HTTPClientType,
			testUtils.JSClientType,
		}),
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						age: Int 
					}`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
					"age":  21,
				},
			},
			testUtils.Request{
				Request: `
					query {
						commits {
							fieldName
							signature {
								type
								identity
								value
							}
						}
					}
				`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"fieldName": "age",
							"signature": map[string]any{
								"type": coreblock.SignatureTypeECDSA256K,
								"identity": gomega.And(
									gomega.Not(gomega.BeEmpty()),
									sameIdentity,
								),
								"value": newSignatureMatcher(makeFieldBlock("age", 21), crypto.KeyTypeSecp256k1),
							},
						},
						{
							"fieldName": "name",
							"signature": map[string]any{
								"type":     coreblock.SignatureTypeECDSA256K,
								"identity": sameIdentity,
								"value": newSignatureMatcher(
									makeFieldBlock("name", "John"), crypto.KeyTypeSecp256k1),
							},
						},
						{
							"fieldName": "_C",
							"signature": map[string]any{
								"type":     coreblock.SignatureTypeECDSA256K,
								"identity": sameIdentity,
								"value":    gomega.Not(gomega.BeEmpty()),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSignature_WithUpdatedDocsAndCommitQuery_ShouldSignOnlyFirstFieldBlocks(t *testing.T) {
	uniqueSignature := testUtils.NewUniqueValue()
	sameIdentity := testUtils.NewSameValue()

	test := testUtils.TestCase{
		EnableSigning: true,
		SupportedClientTypes: immutable.Some([]state.ClientType{
			// C bindings do not support calling functions with non-Secp256k key yet
			testUtils.GoClientType,
			testUtils.CLIClientType,
			testUtils.HTTPClientType,
			testUtils.JSClientType,
		}),
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
				},
			},
			testUtils.UpdateDoc{
				Doc: `{
					"name": "John Doe"
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"name": "John Doe Junior"
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						commits(order: {height: DESC}) {
							fieldName
							height
							signature {
								type
								identity
								value
							}
						}
					}
				`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"fieldName": "name",
							"height":    3,
							"signature": nil,
						},
						{
							"fieldName": "_C",
							"height":    3,
							"signature": map[string]any{
								"type":     coreblock.SignatureTypeECDSA256K,
								"identity": sameIdentity,
								"value":    uniqueSignature,
							},
						},
						{
							"fieldName": "name",
							"height":    2,
							"signature": nil,
						},
						{
							"fieldName": "_C",
							"height":    2,
							"signature": map[string]any{
								"type":     coreblock.SignatureTypeECDSA256K,
								"identity": sameIdentity,
								"value":    uniqueSignature,
							},
						},
						{
							"fieldName": "name",
							"height":    1,
							"signature": map[string]any{
								"type":     coreblock.SignatureTypeECDSA256K,
								"identity": sameIdentity,
								"value":    uniqueSignature,
							},
						},
						{
							"fieldName": "_C",
							"height":    1,
							"signature": map[string]any{
								"type":     coreblock.SignatureTypeECDSA256K,
								"identity": sameIdentity,
								"value":    uniqueSignature,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSignature_WithDeletedDocAndCommitQuery_ShouldIncludeSignatureData(t *testing.T) {
	uniqueSignature := testUtils.NewUniqueValue()

	test := testUtils.TestCase{
		EnableSigning: true,
		SupportedClientTypes: immutable.Some([]state.ClientType{
			// C bindings do not support calling functions with non-Secp256k key yet
			testUtils.GoClientType,
			testUtils.CLIClientType,
			testUtils.HTTPClientType,
			testUtils.JSClientType,
		}),
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
				},
			},
			testUtils.DeleteDoc{},
			testUtils.Request{
				Request: `
					query {
						commits(order: {height: DESC}, fieldName: "_C") {
							fieldName
							height
							signature {
								type
								identity
								value
							}
						}
					}
				`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"fieldName": "_C",
							"height":    2,
							"signature": map[string]any{
								"type":     coreblock.SignatureTypeECDSA256K,
								"identity": newIdentityMatcher(testUtils.NodeIdentity(0).Value()),
								"value":    uniqueSignature,
							},
						},
						{
							"fieldName": "_C",
							"height":    1,
							"signature": map[string]any{
								"type":     coreblock.SignatureTypeECDSA256K,
								"identity": newIdentityMatcher(testUtils.NodeIdentity(0).Value()),
								"value":    uniqueSignature,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSignature_WithEd25519KeyType_ShouldIncludeSignatureData(t *testing.T) {
	test := testUtils.TestCase{
		EnableSigning: true,
		IdentityTypes: map[state.Identity]crypto.KeyType{
			testUtils.NodeIdentity(0).Value(): crypto.KeyTypeEd25519,
		},
		SupportedClientTypes: immutable.Some([]state.ClientType{
			// C bindings do not support calling functions with non-Secp256k key yet
			testUtils.GoClientType,
			testUtils.CLIClientType,
			testUtils.HTTPClientType,
			testUtils.JSClientType,
		}),
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						age: Int 
					}`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
					"age":  21,
				},
			},
			testUtils.Request{
				Request: `
					query {
						commits {
							fieldName
							signature {
								type
								identity
								value
							}
						}
					}
				`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"fieldName": "age",
							"signature": map[string]any{
								"type":     coreblock.SignatureTypeEd25519,
								"identity": newIdentityMatcher(testUtils.NodeIdentity(0).Value()),
								"value":    newSignatureMatcher(makeFieldBlock("age", 21), crypto.KeyTypeEd25519),
							},
						},
						{
							"fieldName": "name",
							"signature": map[string]any{
								"type":     coreblock.SignatureTypeEd25519,
								"identity": newIdentityMatcher(testUtils.NodeIdentity(0).Value()),
								"value":    newSignatureMatcher(makeFieldBlock("name", "John"), crypto.KeyTypeEd25519),
							},
						},
						{
							"fieldName": "_C",
							"signature": map[string]any{
								"type":     coreblock.SignatureTypeEd25519,
								"identity": newIdentityMatcher(testUtils.NodeIdentity(0).Value()),
								"value":    gomega.Not(gomega.BeEmpty()),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// TODO: This test can be enabled as part of https://github.com/sourcenetwork/defradb/issues/3573
// It doesn't pass at the moment because the client identity is being passed to the node
func TestSignature_WithClientIdentity_ShouldUseItForSigning(t *testing.T) {
	t.Skip("Skipping test because signing with client identity is not supported yet")
	test := testUtils.TestCase{
		EnableSigning: true,
		IdentityTypes: map[state.Identity]crypto.KeyType{
			testUtils.ClientIdentity(0).Value(): crypto.KeyTypeEd25519,
			testUtils.NodeIdentity(0).Value():   crypto.KeyTypeSecp256k1,
		},
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						age: Int 
					}`,
			},
			testUtils.CreateDoc{
				Identity: testUtils.ClientIdentity(0),
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"age": 23
				}`,
			},
			testUtils.UpdateDoc{
				Identity: testUtils.ClientIdentity(0),
				Doc: `{
					"name": "John Doe"
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						commits(fieldName: "_C", order: {height: DESC}) {
							height
							signature {
								type
								identity
							}
						}
					}
				`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"height": 3,
							"signature": map[string]any{
								"type":     coreblock.SignatureTypeEd25519,
								"identity": newIdentityMatcher(testUtils.ClientIdentity(0).Value()),
							},
						},
						{
							"height": 2,
							"signature": map[string]any{
								"type":     coreblock.SignatureTypeECDSA256K,
								"identity": newIdentityMatcher(testUtils.NodeIdentity(0).Value()),
							},
						},
						{
							"height": 1,
							"signature": map[string]any{
								"type":     coreblock.SignatureTypeEd25519,
								"identity": newIdentityMatcher(testUtils.ClientIdentity(0).Value()),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
