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

package encryption

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestDocEncryption_WithEncryptionSecondaryRelations_ShouldStoreEncryptedCommit(t *testing.T) {
	const userDocID = "bae-32a035a1-1d5c-5a38-9637-04abfe64dd16"
	const deviceDocID = "bae-3d4ad011-fdf2-502a-a672-9df76b4bbc51"

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Chris"
				}`,
				IsDocEncrypted: true,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
				IsDocEncrypted: true,
			},
			&action.Request{
				Request: `
					query {
						_commits {
							delta
							docID
							fieldName
						}
					}
				`,
				Results: map[string]any{
					"_commits": []map[string]any{
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
						{
							"delta":     encrypt(testUtils.CBORValue(userDocID), deviceDocID, ""),
							"docID":     deviceDocID,
							"fieldName": "_ownerID",
						},
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
							"delta":     nil,
							"docID":     deviceDocID,
							"fieldName": "_C",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
