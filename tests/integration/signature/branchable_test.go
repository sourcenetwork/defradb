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

	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestSignature_WithBranchableCollection_ShouldSignCollectionBlocks(t *testing.T) {
	test := testUtils.TestCase{
		EnableSigning: true,
		SupportedClientTypes: immutable.Some([]state.ClientType{
			// C bindings do not support calling functions with non-Secp256k key yet
			state.GoClientType,
			state.CLIClientType,
			state.HTTPClientType,
			state.JSClientType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users @branchable {
						name: String
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.Request{
				Request: `query {
						_commits {
							fieldName
							signature {
								type
								identity
								value
							}
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"fieldName": nil,
							"signature": map[string]any{
								"type":     coreblock.SignatureTypeECDSA256K,
								"identity": gomega.Not(gomega.BeEmpty()),
								"value":    gomega.Not(gomega.BeEmpty()),
							},
						},
						{
							"fieldName": "name",
							"signature": map[string]any{
								"type":     coreblock.SignatureTypeECDSA256K,
								"identity": gomega.Not(gomega.BeEmpty()),
								"value":    gomega.Not(gomega.BeEmpty()),
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
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
