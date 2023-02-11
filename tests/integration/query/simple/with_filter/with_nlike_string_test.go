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

func TestQuerySimpleWithNotLikeStringContainsFilterBlockContainsString(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with basic not like-string filter contains string",
		Request: `query {
					users(filter: {Name: {_nlike: "%Stormborn%"}}) {
						Name
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "Daenerys Stormborn of House Targaryen, the First of Her Name",
					"HeightM": 1.65
				}`,
				`{
					"Name": "Viserys I Targaryen, King of the Andals",
					"HeightM": 1.82
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "Viserys I Targaryen, King of the Andals",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithNotLikeStringContainsFilterBlockAsPrefixString(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with basic not like-string filter with string as prefix",
		Request: `query {
					users(filter: {Name: {_nlike: "Viserys%"}}) {
						Name
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "Daenerys Stormborn of House Targaryen, the First of Her Name",
					"HeightM": 1.65
				}`,
				`{
					"Name": "Viserys I Targaryen, King of the Andals",
					"HeightM": 1.82
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "Daenerys Stormborn of House Targaryen, the First of Her Name",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithNotLikeStringContainsFilterBlockAsSuffixString(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with basic not like-string filter with string as suffix",
		Request: `query {
					users(filter: {Name: {_nlike: "%Andals"}}) {
						Name
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "Daenerys Stormborn of House Targaryen, the First of Her Name",
					"HeightM": 1.65
				}`,
				`{
					"Name": "Viserys I Targaryen, King of the Andals",
					"HeightM": 1.82
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "Daenerys Stormborn of House Targaryen, the First of Her Name",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithNotLikeStringContainsFilterBlockExactString(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with basic not like-string filter with string as suffix",
		Request: `query {
					users(filter: {Name: {_nlike: "Daenerys Stormborn of House Targaryen, the First of Her Name"}}) {
						Name
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "Daenerys Stormborn of House Targaryen, the First of Her Name",
					"HeightM": 1.65
				}`,
				`{
					"Name": "Viserys I Targaryen, King of the Andals",
					"HeightM": 1.82
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "Viserys I Targaryen, King of the Andals",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithNotLikeStringContainsFilterBlockContainsStringMuplitpleResults(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with basic not like-string filter with contains string mulitple results",
		Request: `query {
					users(filter: {Name: {_nlike: "%Targaryen%"}}) {
						Name
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "Daenerys Stormborn of House Targaryen, the First of Her Name",
					"HeightM": 1.65
				}`,
				`{
					"Name": "Viserys I Targaryen, King of the Andals",
					"HeightM": 1.82
				}`,
			},
		},
		Results: []map[string]any{},
	}

	executeTestCase(t, test)
}
