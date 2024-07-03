// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package encryption

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestDocEncryptionField_WithEncryptionOnField_ShouldStoreOnlyFieldsDeltaEncrypted(t *testing.T) {
	const docID = "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3"

	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
				EncryptedFields: []string{"age"},
			},
			testUtils.Request{
				Request: `
					query {
						commits {
							cid
							delta
							docID
							fieldId
							fieldName
						}
					}
				`,
				Results: []map[string]any{
					{
						"cid":       "bafyreih7ry7ef26xn3lm2rhxusf2rbgyvl535tltrt6ehpwtvdnhlmptiu",
						"delta":     encrypt(testUtils.CBORValue(21)),
						"docID":     docID,
						"fieldId":   "1",
						"fieldName": "age",
					},
					{
						"cid":       "bafyreic2sba5sffkfnt32wfeoaw4qsqozjb5acwwtouxuzllb3aymjwute",
						"delta":     testUtils.CBORValue("John"),
						"docID":     docID,
						"fieldId":   "2",
						"fieldName": "name",
					},
					{
						"cid":       "bafyreifwckkbrr4vzgv5k6sc6jbp6xsns6w75lm2pemjcaenlkyz5qqzam",
						"delta":     nil,
						"docID":     docID,
						"fieldId":   "C",
						"fieldName": nil,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
