// Copyright 2024 Democratized Data Foundation
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
	test := testUtils.TestCase{
		EnabledBlockSigning: true,
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						age: Int @crdt(type: lww)
						verified: Boolean
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
							"signature": nil,
						},
						{
							"fieldName": "name",
							"signature": nil,
						},
						{
							"fieldName": nil,
							"signature": map[string]any{
								"type":     "ECDSA",
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
