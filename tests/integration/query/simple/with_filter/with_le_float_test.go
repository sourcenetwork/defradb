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

func TestQuerySimpleWithFloatLEFilterBlockWithEqualValue(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic le float filter with equal value",
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
					Users(filter: {HeightM: {_le: 1.82}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Bob",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithFloatLEFilterBlockWithGreaterValue(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic le float filter with greater value",
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
					Users(filter: {HeightM: {_le: 1.820000000001}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Bob",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithFloatLEFilterBlockWithGreaterIntValue(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic le float filter with greater int value",
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
					Users(filter: {HeightM: {_le: 2}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Bob",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithFloatLEFilterBlockWithNullValue(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic le float filter with null value",
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
					Users(filter: {HeightM: {_le: null}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Bob",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
