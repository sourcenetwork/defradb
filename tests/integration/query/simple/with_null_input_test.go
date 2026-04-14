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

func TestQuerySimple_WithNullFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Passing null as the filter argument returns all documents.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			&action.Request{
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
		Description: "Passing null for individual filter fields behaves as if no filter is applied.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": null
				}`,
			},
			&action.Request{
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
		Description: "Passing null as the order argument returns all documents in default order.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			&action.Request{
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
		Description: "Passing null for individual order fields returns documents in default order.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			&action.Request{
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
		Description: "Passing null as the limit argument returns all documents without a limit.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			&action.Request{
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
		Description: "Passing null as the offset argument returns all documents from the beginning.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			&action.Request{
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
		Description: "Passing null as the docID argument returns all documents.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			&action.Request{
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
		Description: "Passing null for the docIDs argument returns all documents.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			&action.Request{
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
		Description: "Passing null as the cid argument returns all documents at the current state.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			&action.Request{
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
		Description: "Passing null as the groupBy argument returns all documents ungrouped.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			&action.Request{
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
		Description: "Passing null as showDeleted returns all non-deleted documents.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			&action.Request{
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
		Description: "Filter with a null _or condition behaves as if no _or filter is applied.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			&action.Request{
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
		Description: "Filter with a null element inside _or returns a validation error.",
		Actions: []any{
			&action.Request{
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
		Description: "Filter with a null field value inside _or returns a validation error.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": null
				}`,
			},
			&action.Request{
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
		Description: "Filter with a null _and condition behaves as if no _and filter is applied.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			&action.Request{
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
		Description: "Filter with a null element inside _and returns a validation error.",
		Actions: []any{
			&action.Request{
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
		Description: "Filter with a null field value inside _and returns a validation error.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": null
				}`,
			},
			&action.Request{
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
		Description: "Filter with a null _not condition behaves as if no _not filter is applied.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			&action.Request{
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
		Description: "Filter with a null field value inside _not behaves as if no _not filter is applied.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			&action.Request{
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
