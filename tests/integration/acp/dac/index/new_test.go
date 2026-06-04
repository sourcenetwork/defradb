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

package test_acp_dac_index

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_IndexNewWithSeparateRequest_OnCollectionWithPolicy_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{

			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   userPolicy,
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

			&action.NewIndex{
				CollectionID: 0,
				IndexName:    "some_index",
				FieldName:    "name",
			},

			&action.Request{
				Request: `
					query  {
						Users {
							name
							age
						}
					}`,

				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_IndexNewWithDirective_OnCollectionWithPolicy_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{

			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   userPolicy,
			},

			&action.AddCollection{
				SDL: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String @index
						age: Int
					}
				`,
			},

			&action.Request{
				Request: `
					query  {
						Users {
							name
							age
						}
					}`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
