// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_one

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/stretchr/testify/require"
)

// This test documents unwanted behaviour, see the linked ticket for more info:
// https://github.com/sourcenetwork/defradb/issues/1709
func TestQueryOneToOne_WithVersionOnOuter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Embedded commits query within one-one query",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Book {
						name: String
						author: Author
					}
				
					type Author {
						name: String
						published: Book @primary
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "فارسی دوم دبستان"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "نمی دانم",
					"published": "bae-c052eade-23f6-5ee3-8067-20004e746be3"
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						Book {
							name
							_version {
								docID
							}
							author {
								name
							}
						}
					}
				`,
				Results: []map[string]any{
					{
						"name": "نمی دانم",
						"_version": []map[string]any{
							{
								"docID": "bae-c052eade-23f6-5ee3-8067-20004e746be3",
							},
						},
						"author": map[string]any{
							"name": "فارسی دوم دبستان",
						},
					},
				},
			},
		},
	}

	require.Panics(t,
		func() { testUtils.ExecuteTestCase(t, test) },
	)
}
