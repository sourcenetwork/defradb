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

	"github.com/onsi/gomega"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSignature_WithCommitQuery_ShouldIncludeSignatureData(t *testing.T) {
	uniqueSignature := testUtils.NewUniqueValue()
	sameIdentity := testUtils.NewSameValue()

	test := testUtils.TestCase{
		EnabledBlockSigning: true,
		Actions: []any{
			testUtils.SchemaUpdate{
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
								"type": "ECDSA",
								"identity": gomega.And(
									gomega.Not(gomega.BeEmpty()),
									sameIdentity,
								),
								"value": uniqueSignature,
							},
						},
						{
							"fieldName": "name",
							"signature": map[string]any{
								"type":     "ECDSA",
								"identity": sameIdentity,
								"value":    uniqueSignature,
							},
						},
						{
							"fieldName": nil,
							"signature": map[string]any{
								"type":     "ECDSA",
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

func TestSignature_WithUpdatedDocsAndCommitQuery_ShouldSignOnlyFirstFieldBlocks(t *testing.T) {
	uniqueSignature := testUtils.NewUniqueValue()
	sameIdentity := testUtils.NewSameValue()

	test := testUtils.TestCase{
		EnabledBlockSigning: true,
		Actions: []any{
			testUtils.SchemaUpdate{
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
							"fieldName": nil,
							"height":    3,
							"signature": map[string]any{
								"type":     "ECDSA",
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
							"fieldName": nil,
							"height":    2,
							"signature": map[string]any{
								"type":     "ECDSA",
								"identity": sameIdentity,
								"value":    uniqueSignature,
							},
						},
						{
							"fieldName": "name",
							"height":    1,
							"signature": map[string]any{
								"type":     "ECDSA",
								"identity": sameIdentity,
								"value":    uniqueSignature,
							},
						},
						{
							"fieldName": nil,
							"height":    1,
							"signature": map[string]any{
								"type":     "ECDSA",
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
