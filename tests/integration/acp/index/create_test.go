// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_index

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_IndexCreateWithSeparateRequest_OnCollectionWithPolicy_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test acp, with creating new index using separate request on permissioned collection, no error",
		Actions: []any{

			testUtils.AddPolicy{
				Identity:         testUtils.ClientIdentity(1),
				Policy:           userPolicy,
				ExpectedPolicyID: "abe378ae8dac56f43238b56126a5a5ff1d1021e6bf8027d477b5a366e6238fc2",
			},

			testUtils.SchemaUpdate{
				Schema: `
					type Users @policy(
						id: "abe378ae8dac56f43238b56126a5a5ff1d1021e6bf8027d477b5a366e6238fc2",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,
			},

			testUtils.CreateIndex{
				CollectionID: 0,
				IndexName:    "some_index",
				FieldName:    "name",
			},

			testUtils.Request{
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

func TestACP_IndexCreateWithDirective_OnCollectionWithPolicy_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test acp, with creating new index using directive on permissioned collection, no error",
		Actions: []any{

			testUtils.AddPolicy{
				Identity:         testUtils.ClientIdentity(1),
				Policy:           userPolicy,
				ExpectedPolicyID: "abe378ae8dac56f43238b56126a5a5ff1d1021e6bf8027d477b5a366e6238fc2",
			},

			testUtils.SchemaUpdate{
				Schema: `
					type Users @policy(
						id: "abe378ae8dac56f43238b56126a5a5ff1d1021e6bf8027d477b5a366e6238fc2",
						resource: "users"
					) {
						name: String @index
						age: Int
					}
				`,
			},

			testUtils.Request{
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
