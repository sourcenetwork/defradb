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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryOneToOne_WithVersionOnOuterBeforeJoin(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
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
					"published": "bae-7183862b-1638-5fc1-a3dd-b567fc1346e3"
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						Book {
							name
							author {
								name
							}
							_version {
								docID
							}
						}
					}
				`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name": "فارسی دوم دبستان",
							"_version": []map[string]any{
								{
									"docID": "bae-7183862b-1638-5fc1-a3dd-b567fc1346e3",
								},
							},
							"author": map[string]any{
								"name": "نمی دانم",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOne_WithVersionOnOuterAfterJoin(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
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
					"published": "bae-7183862b-1638-5fc1-a3dd-b567fc1346e3"
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
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name": "فارسی دوم دبستان",
							"_version": []map[string]any{
								{
									"docID": "bae-7183862b-1638-5fc1-a3dd-b567fc1346e3",
								},
							},
							"author": map[string]any{
								"name": "نمی دانم",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
