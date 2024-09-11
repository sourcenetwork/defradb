// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package collection_description

import (
	"testing"

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"

	"github.com/sourcenetwork/immutable"
)

func TestCollectionDescription_WithDefaultFieldValues(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						active: Boolean @default(bool: true)
						created: DateTime @default(dateTime: "2000-07-23T03:00:00-00:00")
						name: String @default(string: "Bob")
						age: Int @default(int: 10)
						points: Float @default(float: 30)
						metadata: JSON @default(json: "{\"value\":1}")
						image: Blob @default(blob: "ff0099")
					}
				`,
			},
			testUtils.GetCollections{
				ExpectedResults: []client.CollectionDescription{
					{
						Name: immutable.Some("Users"),
						Fields: []client.CollectionFieldDescription{
							{
								ID:   0,
								Name: "_docID",
							},
							{
								ID:           1,
								Name:         "active",
								DefaultValue: true,
							},
							{
								ID:           2,
								Name:         "age",
								DefaultValue: float64(10),
							},
							{
								ID:           3,
								Name:         "created",
								DefaultValue: "2000-07-23T03:00:00-00:00",
							},
							{
								ID:           4,
								Name:         "image",
								DefaultValue: "ff0099",
							},
							{
								ID:           5,
								Name:         "metadata",
								DefaultValue: "{\"value\":1}",
							},
							{
								ID:           6,
								Name:         "name",
								DefaultValue: "Bob",
							},
							{
								ID:           7,
								Name:         "points",
								DefaultValue: float64(30),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
