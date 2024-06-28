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

func TestQuerySimpleWithHeightMGEFilterBlockWithEqualValue(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with basic ge int filter with equal value",
		Request: `query {
					Users(filter: {HeightM: {_ge: 2.1}}) {
						Name
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"HeightM": 2.1
				}`,
				`{
					"Name": "Bob",
					"HeightM": 1.82
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "John",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithHeightMGEFilterBlockWithLesserValue(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with basic ge int filter with lesser value",
		Request: `query {
					Users(filter: {HeightM: {_ge: 2.0999999999999}}) {
						Name
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"HeightM": 2.1
				}`,
				`{
					"Name": "Bob",
					"HeightM": 1.82
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "John",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithHeightMGEFilterBlockWithLesserIntValue(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with basic ge int filter with lesser int value",
		Request: `query {
					Users(filter: {HeightM: {_ge: 2}}) {
						Name
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"HeightM": 2.1
				}`,
				`{
					"Name": "Bob",
					"HeightM": 1.82
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "John",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithHeightMGEFilterBlockWithNilValue(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with basic ge float nil filter",
		Request: `query {
					Users(filter: {HeightM: {_ge: null}}) {
						Name
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"HeightM": 2.1
				}`,
				`{
					"Name": "Bob"
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "John",
			},
			{
				"Name": "Bob",
			},
		},
	}

	executeTestCase(t, test)
}
