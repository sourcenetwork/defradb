// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package simple

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimpleWithIntGreaterThanFilterBlock(t *testing.T) {
	tests := []testUtils.TestCase{
		{
			Description: "Simple query with basic filter(age), greater than",
			Actions: []any{
				testUtils.CreateDoc{
					Doc: `{
						"Name": "John",
						"Age": 21
					}`,
				},
				testUtils.CreateDoc{
					Doc: `{
						"Name": "Bob",
						"Age": 19
					}`,
				},
				testUtils.Request{
					Request: `query {
						Users(filter: {Age: {_gt: 20}}) {
							Name
							Age
						}
					}`,
					Results: map[string]any{
						"Users": []map[string]any{
							{
								"Name": "John",
								"Age":  int64(21),
							},
						},
					},
				},
			},
		},
		{
			Description: "Simple query with basic filter(age), no results",
			Actions: []any{
				testUtils.CreateDoc{
					Doc: `{
						"Name": "John",
						"Age": 21
					}`,
				},
				testUtils.CreateDoc{
					Doc: `{
						"Name": "Bob",
						"Age": 32
					}`,
				},
				testUtils.Request{
					Request: `query {
						Users(filter: {Age: {_gt: 40}}) {
							Name
							Age
						}
					}`,
					Results: map[string]any{
						"Users": []map[string]any{},
					},
				},
			},
		},
		{
			Description: "Simple query with basic filter(age), multiple results",
			Actions: []any{
				testUtils.CreateDoc{
					Doc: `{
						"Name": "John",
						"Age": 21
					}`,
				},
				testUtils.CreateDoc{
					Doc: `{
						"Name": "Bob",
						"Age": 32
					}`,
				},
				testUtils.Request{
					Request: `query {
						Users(filter: {Age: {_gt: 20}}) {
							Name
							Age
						}
					}`,
					Results: map[string]any{
						"Users": []map[string]any{
							{
								"Name": "John",
								"Age":  int64(21),
							},
							{
								"Name": "Bob",
								"Age":  int64(32),
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		executeTestCase(t, test)
	}
}

func TestQuerySimpleWithIntGreaterThanFilterBlockWithNullFilterValue(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic int greater than filter, with null filter value",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {Age: {_gt: null}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
