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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/crypto"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestDocSignature_WithEnabledSigning_ShouldQuery(t *testing.T) {
	test := testUtils.TestCase{
		SigningAlg: immutable.Some(crypto.KeyTypeSecp256k1),
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
                    type Users {
                        name: String
                        age: Int
                    }
                `},
			testUtils.CreateDoc{
				Doc: `{
					"name":	"John",
					"age":	21
				}`,
			},
			testUtils.Request{
				Request: `
                    query {
                        Users {
                            _docID
                            name
                            age
                        }
                    }`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"_docID": testUtils.NewDocIndex(0, 0),
							"name":   "John",
							"age":    int64(21),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
