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

func TestQuerySimpleWithFloatGreaterThanFilterBlock(t *testing.T) {
	tests := []testUtils.TestCase{
		{
			Actions: []any{
				testUtils.CreateDoc{
					Doc: `{
						"Name": "John",
						"HeightM": 2.1
					}`,
				},
				testUtils.CreateDoc{
					Doc: `{
						"Name": "Bob",
						"HeightM": 1.82
					}`,
				},
				testUtils.Request{
					Request: `query {
						Users(filter: {HeightM: {_gt: 2.0999999999999}}) {
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
		},
		{
			Actions: []any{
				testUtils.CreateDoc{
					Doc: `{
						"Name": "John",
						"HeightM": 2.1
					}`,
				},
				testUtils.CreateDoc{
					Doc: `{
						"Name": "Bob",
						"HeightM": 1.82
					}`,
				},
				testUtils.Request{
					Request: `query {
						Users(filter: {HeightM: {_gt: 40}}) {
							Name
						}
					}`,
					Results: map[string]any{
						"Users": []map[string]any{},
					},
				},
			},
		},
		{
			Actions: []any{
				testUtils.CreateDoc{
					Doc: `{
						"Name": "John",
						"HeightM": 2.1
					}`,
				},
				testUtils.CreateDoc{
					Doc: `{
						"Name": "Bob",
						"HeightM": 1.82
					}`,
				},
				testUtils.Request{
					Request: `query {
						Users(filter: {HeightM: {_gt: 1.8199999999999}}) {
							Name
						}
					}`,
					Results: map[string]any{
						"Users": []map[string]any{
							{
								"Name": "John",
							},
							{
								"Name": "Bob",
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

func TestQuerySimpleWithFloatGreaterThanFilterBlockWithIntFilterValue(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 2.1
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"HeightM": 1.82
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {HeightM: {_gt: 2}}) {
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

func TestQuerySimpleWithFloatGreaterThanFilterBlockWithNullFilterValue(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 2.1
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {HeightM: {_gt: null}}) {
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
