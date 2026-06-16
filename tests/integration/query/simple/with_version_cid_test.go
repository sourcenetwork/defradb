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

package simple

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/multiplier"
)

func TestQuerySimpleWithVersionAndCid(t *testing.T) {
	test := testUtils.TestCase{
		// hardcoded CIDs would change under encryption
		MultiplierExcludes: []string{multiplier.EncryptedDocs},
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						Name: String
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"Name": "John",
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"Name": "Chris",
				},
			},
			&action.Request{
				Request: `query {
					Users(cid: "{{.CID0_1_0}}") {
						Name
						_version {
							fieldName
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Chris",
							"_version": []map[string]any{
								{
									"fieldName": "_C",
								},
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
