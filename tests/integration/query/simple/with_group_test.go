// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package simple

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimpleWithGroupByEmpty(t *testing.T) {
	test := testUtils.TestCase{
		Description: "groupBy with an empty field list returns an error.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: []) {
						GROUP {
							Name
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"GROUP": []map[string]any{
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
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByNumber(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Group documents by an integer field returns one entry per distinct value.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Age]) {
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": int64(32),
						},
						{
							"Age": int64(55),
						},
						{
							"Age": int64(19),
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByDateTime(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Group documents by a DateTime field returns one entry per distinct datetime.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"CreatedAt": "2011-07-23T03:46:56-05:00"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"CreatedAt": "2011-07-23T03:46:56-05:00"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Carlo",
					"CreatedAt": "2012-07-23T03:46:56-05:00"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"CreatedAt": "2013-07-23T03:46:56-05:00"
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [CreatedAt]) {
						CreatedAt
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"CreatedAt": testUtils.MustParseTime("2012-07-23T03:46:56-05:00"),
						},
						{
							"CreatedAt": testUtils.MustParseTime("2013-07-23T03:46:56-05:00"),
						},
						{
							"CreatedAt": testUtils.MustParseTime("2011-07-23T03:46:56-05:00"),
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByNumberWithGroupString(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Group by integer with a GROUP sub-selection returns nested documents per group.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Age]) {
						Age
						GROUP {
							Name
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": int64(32),
							"GROUP": []map[string]any{
								{
									"Name": "Bob",
								},
								{
									"Name": "John",
								},
							},
						},
						{
							"Age": int64(55),
							"GROUP": []map[string]any{
								{
									"Name": "Carlo",
								},
							},
						},
						{
							"Age": int64(19),
							"GROUP": []map[string]any{
								{
									"Name": "Alice",
								},
							},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByWithoutGroupedFieldSelectedWithInnerGroup(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Grouping without selecting the grouped field still returns correct GROUP children.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 25
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Name]) {
						Name
						GROUP {
							Age
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"GROUP": []map[string]any{
								{
									"Age": int64(25),
								},
								{
									"Age": int64(32),
								},
							},
						},
						{
							"Name": "Carlo",
							"GROUP": []map[string]any{
								{
									"Age": int64(55),
								},
							},
						},
						{
							"Name": "Alice",
							"GROUP": []map[string]any{
								{
									"Age": int64(19),
								},
							},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByString(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Group documents by a string field returns one entry per distinct string value.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 25
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Name]) {
						Name
						GROUP {
							Age
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"GROUP": []map[string]any{
								{
									"Age": int64(25),
								},
								{
									"Age": int64(32),
								},
							},
						},
						{
							"Name": "Carlo",
							"GROUP": []map[string]any{
								{
									"Age": int64(55),
								},
							},
						},
						{
							"Name": "Alice",
							"GROUP": []map[string]any{
								{
									"Age": int64(19),
								},
							},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringWithInnerGroupBoolean(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Group by string with a nested GROUP by boolean returns doubly-nested groups.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 25,
					"Verified": true
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 32,
					"Verified": true
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 34,
					"Verified": false
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55,
					"Verified": true
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19,
					"Verified": false
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Name]) {
						Name
						GROUP (groupBy: [Verified]){
							Verified
							GROUP {
								Age
							}
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"GROUP": []map[string]any{
								{
									"Verified": true,
									"GROUP": []map[string]any{
										{
											"Age": int64(32),
										},
										{
											"Age": int64(25),
										},
									},
								},
								{
									"Verified": false,
									"GROUP": []map[string]any{
										{
											"Age": int64(34),
										},
									},
								},
							},
						},
						{
							"Name": "Alice",
							"GROUP": []map[string]any{
								{
									"Verified": false,
									"GROUP": []map[string]any{
										{
											"Age": int64(19),
										},
									},
								},
							},
						},
						{
							"Name": "Carlo",
							"GROUP": []map[string]any{
								{
									"Verified": true,
									"GROUP": []map[string]any{
										{
											"Age": int64(55),
										},
									},
								},
							},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringThenBoolean(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Group by string then boolean returns correctly nested two-level groups.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 25,
					"Verified": true
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 32,
					"Verified": true
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 34,
					"Verified": false
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55,
					"Verified": true
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19,
					"Verified": false
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Name, Verified]) {
						Name
						Verified
						GROUP {
							Age
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":     "John",
							"Verified": true,
							"GROUP": []map[string]any{
								{
									"Age": int64(32),
								},
								{
									"Age": int64(25),
								},
							},
						},
						{
							"Name":     "John",
							"Verified": false,
							"GROUP": []map[string]any{
								{
									"Age": int64(34),
								},
							},
						},
						{
							"Name":     "Alice",
							"Verified": false,
							"GROUP": []map[string]any{
								{
									"Age": int64(19),
								},
							},
						},
						{
							"Name":     "Carlo",
							"Verified": true,
							"GROUP": []map[string]any{
								{
									"Age": int64(55),
								},
							},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByBooleanThenNumber(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Group by boolean then number returns correctly nested two-level groups.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 25,
					"Verified": true
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 32,
					"Verified": true
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 34,
					"Verified": false
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55,
					"Verified": true
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19,
					"Verified": false
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Verified, Name]) {
						Name
						Verified
						GROUP {
							Age
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":     "John",
							"Verified": true,
							"GROUP": []map[string]any{
								{
									"Age": int64(32),
								},
								{
									"Age": int64(25),
								},
							},
						},
						{
							"Name":     "John",
							"Verified": false,
							"GROUP": []map[string]any{
								{
									"Age": int64(34),
								},
							},
						},
						{
							"Name":     "Alice",
							"Verified": false,
							"GROUP": []map[string]any{
								{
									"Age": int64(19),
								},
							},
						},
						{
							"Name":     "Carlo",
							"Verified": true,
							"GROUP": []map[string]any{
								{
									"Age": int64(55),
								},
							},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByNumberOnUndefined(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Group by an undefined field with empty collection returns no groups.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice"
				}`,
			},
			&action.Request{
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
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByNumberOnUndefinedWithChildren(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Group by an undefined field with children returns no groups and no sub-documents.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice"
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Age]) {
						Age
						GROUP {
							Name
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": nil,
							"GROUP": []map[string]any{
								{
									"Name": "Bob",
								},
								{
									"Name": "Alice",
								},
							},
						},
						{
							"Age": int64(32),
							"GROUP": []map[string]any{
								{
									"Name": "John",
								},
							},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleErrorsWithNonGroupFieldsSelected(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Selecting a non-grouped field outside of GROUP returns an error.",
		Actions: []any{
			&action.Request{
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
