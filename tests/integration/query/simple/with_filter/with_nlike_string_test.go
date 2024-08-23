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
	test := testUtils.TestCase{
		Description: "Simple query with basic not like-string filter contains string",
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
			testUtils.Request{
				Request: `query {
					Users(filter: {Name: {_nlike: "%Stormborn%"}}) {
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

func TestQuerySimple_WithNotCaseInsensitiveLikeString_ShouldMatchString(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic not case insensitive like-string filter contains string",
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
			testUtils.Request{
				Request: `query {
					Users(filter: {Name: {_nilike: "%stormborn%"}}) {
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

func TestQuerySimpleWithNotLikeStringContainsFilterBlockAsPrefixString(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic not like-string filter with string as prefix",
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
			testUtils.Request{
				Request: `query {
					Users(filter: {Name: {_nlike: "Viserys%"}}) {
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

func TestQuerySimple_WithNotCaseInsensitiveLikeString_ShouldMatchPrefixString(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic not case insensitive like-string filter with string as prefix",
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
			testUtils.Request{
				Request: `query {
					Users(filter: {Name: {_nilike: "viserys%"}}) {
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

func TestQuerySimpleWithNotLikeStringContainsFilterBlockAsSuffixString(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic not like-string filter with string as suffix",
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
			testUtils.Request{
				Request: `query {
					Users(filter: {Name: {_nlike: "%Andals"}}) {
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

func TestQuerySimple_WithNotCaseInsensitiveLikeString_ShouldMatchSuffixString(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic not like-string filter with string as suffix",
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
			testUtils.Request{
				Request: `query {
					Users(filter: {Name: {_nilike: "%andals"}}) {
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

func TestQuerySimpleWithNotLikeStringContainsFilterBlockExactString(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic not like-string filter with string as suffix",
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
			testUtils.Request{
				Request: `query {
					Users(filter: {Name: {_nlike: "Daenerys Stormborn of House Targaryen, the First of Her Name"}}) {
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

func TestQuerySimple_WithNotCaseInsensitiveLikeString_MatchExactString(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic not case insensitive like-string filter with string as suffix",
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
			testUtils.Request{
				Request: `query {
					Users(filter: {Name: {_nilike: "daenerys stormborn of house targaryen, the first of her name"}}) {
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

func TestQuerySimpleWithNotLikeStringContainsFilterBlockContainsStringMuplitpleResults(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic not like-string filter with contains string multiple results",
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
			testUtils.Request{
				Request: `query {
					Users(filter: {Name: {_nlike: "%Targaryen%"}}) {
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

func TestQuerySimpleWithNotLikeStringContainsFilterBlockHasStartAndEnd(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic not like-string filter with string as start and end",
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
			testUtils.Request{
				Request: `query {
					Users(filter: {Name: {_nlike: "Daenerys%Name"}}) {
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

func TestQuerySimpleWithNotLikeStringContainsFilterBlockHasBoth(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic not like-string filter with none of the strings",
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
			testUtils.Request{
				Request: `query {
					Users(filter: {_and: [{Name: {_nlike: "%Baratheon%"}}, {Name: {_nlike: "%Stormborn%"}}]}) {
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

func TestQuerySimpleWithNotLikeStringContainsFilterBlockHasEither(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic not like-string filter with either strings",
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
			testUtils.Request{
				Request: `query {
					Users(filter: {_or: [{Name: {_nlike: "%Baratheon%"}}, {Name: {_nlike: "%Stormborn%"}}]}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Viserys I Targaryen, King of the Andals",
						},
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

func TestQuerySimpleWithNotLikeStringContainsFilterBlockPropNotSet(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic not like-string filter with either strings",
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
			testUtils.Request{
				Request: `query {
					Users(filter: {Name: {_nlike: "%King%"}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Daenerys Stormborn of House Targaryen, the First of Her Name",
						},
						{
							"Name": nil,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
