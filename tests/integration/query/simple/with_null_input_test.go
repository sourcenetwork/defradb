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

func TestQuerySimple_WithNullFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query, with null filter",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: null) {
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

func TestQuerySimple_WithNullFilterFields_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query, with null filter fields",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": null
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {Name: null}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
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

func TestQuerySimple_WithNullOrder_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query, with null order",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(order: null) {
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

func TestQuerySimple_WithNullOrderFields_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query, with null order fields",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(order: {Name: null}) {
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

func TestQuerySimple_WithNullLimit_Succeed(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query, with null limit",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(limit: null) {
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

func TestQuerySimple_WithNullOffset_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query, with null offset",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(offset: null) {
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

func TestQuerySimple_WithNullDocID_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query, with null docID",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(docID: null) {
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

func TestQuerySimple_WithNullDocIDs_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query, with null docIDs",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(docID: null) {
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

func TestQuerySimple_WithNullCID_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query, with null cid",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(cid: null) {
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

func TestQuerySimple_WithNullGroupBy_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query, with null groupBy",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(groupBy: null) {
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

func TestQuerySimple_WithNullShowDeleted_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query, with null showDeleted",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(showDeleted: null) {
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

func TestQuerySimple_WithFilterWithNullOr_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query, with filter with null or",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {_or: null}) {
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

func TestQuerySimple_WithFilterWithNullOrElement_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query, with filter with null or element",
		Actions: []any{
			testUtils.Request{
				Request: `query {
					Users(filter: {_or: [null]}) {
						Name
					}
				}`,
				ExpectedError: `Expected "UsersFilterArg!", found null`,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithFilterWithNullOrField_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query, with filter with or with null field",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": null
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {_or: [{Name: null}]}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
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

func TestQuerySimple_WithFilterWithNullAnd_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query, with filter with null and",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {_and: null}) {
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

func TestQuerySimple_WithFilterWithNullAndElement_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query, with filter with null and element",
		Actions: []any{
			testUtils.Request{
				Request: `query {
					Users(filter: {_and: [null]}) {
						Name
					}
				}`,
				ExpectedError: `Expected "UsersFilterArg!", found null`,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithFilterWithNullAndField_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query, with filter with and with null field",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": null
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {_and: [{Name: null}]}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
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

func TestQuerySimple_WithFilterWithNullNot_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query, with filter with null not",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {_not: null}) {
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

func TestQuerySimple_WithFilterWithNullNotField_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query, with filter with null not field",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {_not: {Name: null}}) {
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
