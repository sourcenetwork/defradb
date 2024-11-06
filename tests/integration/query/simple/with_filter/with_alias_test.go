// Copyright 2024 Democratized Data Foundation
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

func TestQuerySimple_WithAliasEqualsFilterBlock_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with alias filter(age)",
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
					Users(filter: {_alias: {UserAge: {_eq: 21}}}) {
						Name
						UserAge: Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":    "John",
							"UserAge": int64(21),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithEmptyAlias_ShouldNotFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with empty alias filter",
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
					Users(filter: {_alias: {}}) {
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
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithNullAlias_ShouldFilterAll(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with null alias filter",
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
					Users(filter: {_alias: null}) {
						Name
						Age
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

func TestQuerySimple_WithNonObjectAlias_ShouldFilterAll(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with non object alias filter",
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
					Users(filter: {_alias: 1}) {
						Name
						Age
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

func TestQuerySimple_WithNonExistantAlias_ShouldReturnError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with non existant alias filter",
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
					Users(filter: {_alias: {UserAge: {_eq: 21}}}) {
						Name
						Age
					}
				}`,
				ExpectedError: `field or alias not found. Name: UserAge`,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithNonAliasedField_ShouldMatchFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with non aliased filter",
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
					Users(filter: {_alias: {Age: {_eq: 32}}}) {
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Bob",
							"Age":  int64(32),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithCompoundAlias_ShouldMatchFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with compound alias filter",
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
					Users(filter: {
						_and: [
							{_alias: {userAge: {_gt: 30}}},
							{_alias: {userAge: {_lt: 40}}}
						]
					}) {
						Name
						userAge: Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":    "Bob",
							"userAge": int64(32),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithAliasWithCompound_ShouldMatchFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with alias with compound filter",
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
					Users(filter: {
						_alias: {
							_and: [
								{userAge: {_gt: 30}},
								{userAge: {_lt: 40}}
							]
						}
					}) {
						Name
						userAge: Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":    "Bob",
							"userAge": int64(32),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
