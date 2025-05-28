// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package searchable_encryption

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestEncryptedIndex_IfDocCreated_ShouldNotHinderQuerying(t *testing.T) {
	test := testUtils.TestCase{
		EnableSearchableEncryption: true,
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String 
						age: Int @encryptedIndex
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
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
					"Users": []map[string]any{
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
