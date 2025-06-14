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

	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSignature_WithBranchableCollection_ShouldSignCollectionBlocks(t *testing.T) {
	test := testUtils.TestCase{
		EnableSigning: true,
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users @branchable {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
				},
			},
			testUtils.Request{
				Request: `query {
						commits {
							fieldName
							signature {
								type
								identity
								value
							}
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
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
