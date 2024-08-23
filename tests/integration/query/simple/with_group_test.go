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

func TestQuerySimpleWithGroupByEmpty(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group by empty set, children",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
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
					Users(groupBy: []) {
						_group {
							Name
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"_group": []map[string]any{
								{
									"Name": "Bob",
								},
								{
									"Name": "John",
								},
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByNumber(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group by number, no children",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(groupBy: [Age]) {
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": int64(55),
						},
						{
							"Age": int64(32),
						},
						{
							"Age": int64(19),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByDateTime(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group by number, no children",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"CreatedAt": "2011-07-23T03:46:56-05:00"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"CreatedAt": "2011-07-23T03:46:56-05:00"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Carlo",
					"CreatedAt": "2012-07-23T03:46:56-05:00"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"CreatedAt": "2013-07-23T03:46:56-05:00"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(groupBy: [CreatedAt]) {
						CreatedAt
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"CreatedAt": testUtils.MustParseTime("2011-07-23T03:46:56-05:00"),
						},
						{
							"CreatedAt": testUtils.MustParseTime("2012-07-23T03:46:56-05:00"),
						},
						{
							"CreatedAt": testUtils.MustParseTime("2013-07-23T03:46:56-05:00"),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByNumberWithGroupString(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group by string, child string",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(groupBy: [Age]) {
						Age
						_group {
							Name
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": int64(55),
							"_group": []map[string]any{
								{
									"Name": "Carlo",
								},
							},
						},
						{
							"Age": int64(32),
							"_group": []map[string]any{
								{
									"Name": "Bob",
								},
								{
									"Name": "John",
								},
							},
						},
						{
							"Age": int64(19),
							"_group": []map[string]any{
								{
									"Name": "Alice",
								},
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByWithoutGroupedFieldSelectedWithInnerGroup(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with groupBy without selecting field grouped by, with inner _group.",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 25
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(groupBy: [Name]) {
						Name
						_group {
							Age
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Carlo",
							"_group": []map[string]any{
								{
									"Age": int64(55),
								},
							},
						},
						{
							"Name": "John",
							"_group": []map[string]any{
								{
									"Age": int64(25),
								},
								{
									"Age": int64(32),
								},
							},
						},
						{
							"Name": "Alice",
							"_group": []map[string]any{
								{
									"Age": int64(19),
								},
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByString(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group by string",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 25
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(groupBy: [Name]) {
						Name
						_group {
							Age
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Carlo",
							"_group": []map[string]any{
								{
									"Age": int64(55),
								},
							},
						},
						{
							"Name": "John",
							"_group": []map[string]any{
								{
									"Age": int64(25),
								},
								{
									"Age": int64(32),
								},
							},
						},
						{
							"Name": "Alice",
							"_group": []map[string]any{
								{
									"Age": int64(19),
								},
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringWithInnerGroupBoolean(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group by string, with child group by boolean",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 25,
					"Verified": true
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32,
					"Verified": true
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 34,
					"Verified": false
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55,
					"Verified": true
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19,
					"Verified": false
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(groupBy: [Name]) {
						Name
						_group (groupBy: [Verified]){
							Verified
							_group {
								Age
							}
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"_group": []map[string]any{
								{
									"Verified": true,
									"_group": []map[string]any{
										{
											"Age": int64(25),
										},
										{
											"Age": int64(32),
										},
									},
								},
								{
									"Verified": false,
									"_group": []map[string]any{
										{
											"Age": int64(34),
										},
									},
								},
							},
						},
						{
							"Name": "Carlo",
							"_group": []map[string]any{
								{
									"Verified": true,
									"_group": []map[string]any{
										{
											"Age": int64(55),
										},
									},
								},
							},
						},
						{
							"Name": "Alice",
							"_group": []map[string]any{
								{
									"Verified": false,
									"_group": []map[string]any{
										{
											"Age": int64(19),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringThenBoolean(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group by string then by boolean",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 25,
					"Verified": true
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32,
					"Verified": true
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 34,
					"Verified": false
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55,
					"Verified": true
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19,
					"Verified": false
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(groupBy: [Name, Verified]) {
						Name
						Verified
						_group {
							Age
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":     "John",
							"Verified": true,
							"_group": []map[string]any{
								{
									"Age": int64(25),
								},
								{
									"Age": int64(32),
								},
							},
						},
						{
							"Name":     "John",
							"Verified": false,
							"_group": []map[string]any{
								{
									"Age": int64(34),
								},
							},
						},
						{
							"Name":     "Carlo",
							"Verified": true,
							"_group": []map[string]any{
								{
									"Age": int64(55),
								},
							},
						},
						{
							"Name":     "Alice",
							"Verified": false,
							"_group": []map[string]any{
								{
									"Age": int64(19),
								},
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByBooleanThenNumber(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group by boolean then by string",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 25,
					"Verified": true
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32,
					"Verified": true
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 34,
					"Verified": false
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55,
					"Verified": true
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19,
					"Verified": false
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(groupBy: [Verified, Name]) {
						Name
						Verified
						_group {
							Age
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":     "John",
							"Verified": true,
							"_group": []map[string]any{
								{
									"Age": int64(25),
								},
								{
									"Age": int64(32),
								},
							},
						},
						{
							"Name":     "John",
							"Verified": false,
							"_group": []map[string]any{
								{
									"Age": int64(34),
								},
							},
						},
						{
							"Name":     "Carlo",
							"Verified": true,
							"_group": []map[string]any{
								{
									"Age": int64(55),
								},
							},
						},
						{
							"Name":     "Alice",
							"Verified": false,
							"_group": []map[string]any{
								{
									"Age": int64(19),
								},
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByNumberOnUndefined(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group by number, no children, undefined group value",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(groupBy: [Age]) {
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": nil,
						},
						{
							"Age": int64(32),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByNumberOnUndefinedWithChildren(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group by number, with children, undefined group value",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(groupBy: [Age]) {
						Age
						_group {
							Name
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": nil,
							"_group": []map[string]any{
								{
									"Name": "Alice",
								},
								{
									"Name": "Bob",
								},
							},
						},
						{
							"Age": int64(32),
							"_group": []map[string]any{
								{
									"Name": "John",
								},
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleErrorsWithNonGroupFieldsSelected(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group by number, no children",
		Actions: []any{
			testUtils.Request{
				Request: `query {
					Users(groupBy: [Age]) {
						Age
						Name
					}
				}`,
				ExpectedError: "cannot select a non-group-by field at group-level",
			},
		},
	}

	executeTestCase(t, test)
}
