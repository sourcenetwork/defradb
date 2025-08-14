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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestDocEncryption_WithEncryptionSecondaryRelations_ShouldStoreEncryptedCommit(t *testing.T) {
	const userDocID = "bae-73ba4fa1-e4ff-5e03-850a-c6d3b1ccd84f"
	const deviceDocID = "bae-66b61a55-8fd4-5e9d-8d3b-d7838f356297"

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
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
							"fieldName": "_C",
						},
						{
							"delta":     encrypt(testUtils.CBORValue("Chris"), userDocID, ""),
							"docID":     userDocID,
							"fieldName": "name",
						},
						{
							"delta":     nil,
							"docID":     userDocID,
							"fieldName": "_C",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
