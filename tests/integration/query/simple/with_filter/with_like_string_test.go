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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/multiplier"
)

func TestQuerySimpleWithLikeStringContainsFilterBlockContainsString(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Daenerys Stormborn of House Targaryen, the First of Her Name",
					"HeightM": 1.65
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Viserys I Targaryen, King of the Andals",
					"HeightM": 1.82
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {Name: {_like: "%Stormborn%"}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Daenerys Stormborn of House Targaryen, the First of Her Name",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithCaseInsensitiveLike_ShouldMatchString(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Daenerys Stormborn of House Targaryen, the First of Her Name",
					"HeightM": 1.65
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Viserys I Targaryen, King of the Andals",
					"HeightM": 1.82
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {Name: {_ilike: "%stormborn%"}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Daenerys Stormborn of House Targaryen, the First of Her Name",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithLikeStringContainsFilterBlockAsPrefixString(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Daenerys Stormborn of House Targaryen, the First of Her Name",
					"HeightM": 1.65
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Viserys I Targaryen, King of the Andals",
					"HeightM": 1.82
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {Name: {_like: "Viserys%"}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Viserys I Targaryen, King of the Andals",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithCaseInsensitiveLikeString_ShouldMatchPrefixString(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Daenerys Stormborn of House Targaryen, the First of Her Name",
					"HeightM": 1.65
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Viserys I Targaryen, King of the Andals",
					"HeightM": 1.82
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {Name: {_ilike: "viserys%"}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Viserys I Targaryen, King of the Andals",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithLikeStringContainsFilterBlockAsSuffixString(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Daenerys Stormborn of House Targaryen, the First of Her Name",
					"HeightM": 1.65
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Viserys I Targaryen, King of the Andals",
					"HeightM": 1.82
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {Name: {_like: "%Andals"}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Viserys I Targaryen, King of the Andals",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithCaseInsensitiveLikeString_ShouldMatchSuffixString(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Daenerys Stormborn of House Targaryen, the First of Her Name",
					"HeightM": 1.65
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Viserys I Targaryen, King of the Andals",
					"HeightM": 1.82
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {Name: {_ilike: "%andals"}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Viserys I Targaryen, King of the Andals",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithLikeStringContainsFilterBlockExactString(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Daenerys Stormborn of House Targaryen, the First of Her Name",
					"HeightM": 1.65
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Viserys I Targaryen, King of the Andals",
					"HeightM": 1.82
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {Name: {_like: "Daenerys Stormborn of House Targaryen, the First of Her Name"}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Daenerys Stormborn of House Targaryen, the First of Her Name",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithCaseInsensitiveLikeString_ShouldMatchExactString(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Daenerys Stormborn of House Targaryen, the First of Her Name",
					"HeightM": 1.65
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Viserys I Targaryen, King of the Andals",
					"HeightM": 1.82
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {Name: {_ilike: "daenerys stormborn of house targaryen, the first of her name"}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Daenerys Stormborn of House Targaryen, the First of Her Name",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithLikeStringContainsFilterBlockContainsStringMuplitpleResults(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Daenerys Stormborn of House Targaryen, the First of Her Name",
					"HeightM": 1.65
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Viserys I Targaryen, King of the Andals",
					"HeightM": 1.82
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {Name: {_like: "%Targaryen%"}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Daenerys Stormborn of House Targaryen, the First of Her Name",
						},
						{
							"Name": "Viserys I Targaryen, King of the Andals",
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithLikeStringContainsFilterBlockHasStartAndEnd(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Daenerys Stormborn of House Targaryen, the First of Her Name",
					"HeightM": 1.65
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Viserys I Targaryen, King of the Andals",
					"HeightM": 1.82
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {Name: {_like: "Daenerys%Name"}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Daenerys Stormborn of House Targaryen, the First of Her Name",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithLikeStringContainsFilterBlockHasBoth(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Daenerys Stormborn of House Targaryen, the First of Her Name",
					"HeightM": 1.65
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Viserys I Targaryen, King of the Andals",
					"HeightM": 1.82
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {_and: [{Name: {_like: "%Baratheon%"}}, {Name: {_like: "%Stormborn%"}}]}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithLikeStringContainsFilterBlockHasEither(t *testing.T) {
	test := testUtils.TestCase{
		// TODO: https://github.com/sourcenetwork/defradb/issues/4353
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Daenerys Stormborn of House Targaryen, the First of Her Name",
					"HeightM": 1.65
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Viserys I Targaryen, King of the Andals",
					"HeightM": 1.82
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {_or: [{Name: {_like: "%Baratheon%"}}, {Name: {_like: "%Stormborn%"}}]}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Daenerys Stormborn of House Targaryen, the First of Her Name",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithLikeStringContainsFilterBlockPropNotSet(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Daenerys Stormborn of House Targaryen, the First of Her Name",
					"HeightM": 1.65
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Viserys I Targaryen, King of the Andals",
					"HeightM": 1.82
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"HeightM": 1.92
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {Name: {_like: "%King%"}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Viserys I Targaryen, King of the Andals",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
