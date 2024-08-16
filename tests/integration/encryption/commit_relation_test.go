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

func TestDocEncryption_WithEncryptionSecondaryRelations_ShouldStoreEncryptedCommit(t *testing.T) {
	const userDocID = "bae-4d563681-e131-5e01-8ab4-6c65ac0d0478"
	const deviceDocID = "bae-29ab9ee8-80cb-53eb-a467-f96a170f4cb7"

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						devices: [Device]
					}

					type Device {
						model: String 
						manufacturer: String
						owner: User
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Chris"
				}`,
				IsDocEncrypted: true,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
				IsDocEncrypted: true,
			},
			testUtils.Request{
				Request: `
					query {
						commits {
							delta
							docID
							fieldName
						}
					}
				`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"delta":     encrypt(testUtils.CBORValue("Sony"), deviceDocID, ""),
							"docID":     deviceDocID,
							"fieldName": "manufacturer",
						},
						{
							"delta":     encrypt(testUtils.CBORValue("Walkman"), deviceDocID, ""),
							"docID":     deviceDocID,
							"fieldName": "model",
						},
						{
							"delta":     encrypt(testUtils.CBORValue(userDocID), deviceDocID, ""),
							"docID":     deviceDocID,
							"fieldName": "owner_id",
						},
						{
							"delta":     nil,
							"docID":     deviceDocID,
							"fieldName": nil,
						},
						{
							"delta":     encrypt(testUtils.CBORValue("Chris"), userDocID, ""),
							"docID":     userDocID,
							"fieldName": "name",
						},
						{
							"delta":     nil,
							"docID":     userDocID,
							"fieldName": nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
