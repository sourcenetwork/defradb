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

package signature

import (
	"testing"

	"github.com/onsi/gomega"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/crypto"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestSignature_WithSigningEnabledAndPerRequestDisabled_ShouldNotSign(t *testing.T) {
	test := testUtils.TestCase{
		EnableSigning: true,
		SupportedMutationTypes: immutable.Some([]state.MutationType{
			state.CollectionNamedMutationType,
			state.CollectionSaveMutationType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						age: Int
					}`,
			},
			&action.AddDoc{
				EnableSigning: immutable.Some(false),
				DocMap: map[string]any{
					"name": "John",
					"age":  21,
				},
			},
			&action.Request{
				Request: `
					query {
						_commits {
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
					"_commits": []map[string]any{
						{
							"fieldName": "age",
							"signature": nil,
						},
						{
							"fieldName": "name",
							"signature": nil,
						},
						{
							"fieldName": "_C",
							"signature": nil,
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSignature_WithSigningDisabledAndPerRequestEnabled_ShouldSign(t *testing.T) {
	test := testUtils.TestCase{
		EnableSigning: false,
		SupportedMutationTypes: immutable.Some([]state.MutationType{
			state.CollectionNamedMutationType,
			state.CollectionSaveMutationType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						age: Int
					}`,
			},
			&action.AddDoc{
				EnableSigning: immutable.Some(true),
				DocMap: map[string]any{
					"name": "John",
					"age":  21,
				},
			},
			&action.Request{
				Request: `
					query {
						_commits {
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
					"_commits": []map[string]any{
						{
							"fieldName": "age",
							"signature": map[string]any{
								"type":     coreblock.SignatureTypeECDSA256K,
								"identity": gomega.Not(gomega.BeEmpty()),
								"value": newSignatureMatcher(
									makeFieldBlock("age", 21), crypto.KeyTypeSecp256k1),
							},
						},
						{
							"fieldName": "name",
							"signature": map[string]any{
								"type":     coreblock.SignatureTypeECDSA256K,
								"identity": gomega.Not(gomega.BeEmpty()),
								"value": newSignatureMatcher(
									makeFieldBlock("name", "John"), crypto.KeyTypeSecp256k1),
							},
						},
						{
							"fieldName": "_C",
							"signature": map[string]any{
								"type":     coreblock.SignatureTypeECDSA256K,
								"identity": gomega.Not(gomega.BeEmpty()),
								"value":    gomega.Not(gomega.BeEmpty()),
							},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSignature_WithSigningEnabledAndNoPerRequestOverride_ShouldSign(t *testing.T) {
	test := testUtils.TestCase{
		EnableSigning: true,
		SupportedMutationTypes: immutable.Some([]state.MutationType{
			state.CollectionNamedMutationType,
			state.CollectionSaveMutationType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						age: Int
					}`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
					"age":  21,
				},
			},
			&action.Request{
				Request: `
					query {
						_commits {
							fieldName
							signature {
								type
							}
						}
					}
				`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"fieldName": "age",
							"signature": map[string]any{
								"type": coreblock.SignatureTypeECDSA256K,
							},
						},
						{
							"fieldName": "name",
							"signature": map[string]any{
								"type": coreblock.SignatureTypeECDSA256K,
							},
						},
						{
							"fieldName": "_C",
							"signature": map[string]any{
								"type": coreblock.SignatureTypeECDSA256K,
							},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
