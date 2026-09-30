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
							"delta":     encryptedCBORValueWithKey(testUtils.CBORValue("Chris"), docRefKey(1, 1), ""),
							"docID":     testUtils.NewDocIndex(0, 0),
							"fieldName": "name",
						},
						{
							"delta":     nil,
							"docID":     testUtils.NewDocIndex(0, 0),
							"fieldName": "_C",
						},
						{
							"delta":     notPlainCBORDocID(testUtils.NewDocIndex(0, 0)),
							"docID":     testUtils.NewDocIndex(1, 0),
							"fieldName": "_ownerID",
						},
						{
							"delta":     notPlainCBORValue(testUtils.CBORValue("Sony")),
							"docID":     testUtils.NewDocIndex(1, 0),
							"fieldName": "manufacturer",
						},
						{
							"delta":     notPlainCBORValue(testUtils.CBORValue("Walkman")),
							"docID":     testUtils.NewDocIndex(1, 0),
							"fieldName": "model",
						},
						{
							"delta":     nil,
							"docID":     testUtils.NewDocIndex(1, 0),
							"fieldName": "_C",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
