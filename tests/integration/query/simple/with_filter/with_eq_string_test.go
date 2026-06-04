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

func TestQuerySimpleWithStringFilterBlock(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
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
					Users(filter: {Name: {_eq: "John"}}) {
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
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithStringEqualsNilFilterBlock(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
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
					"Age": 60
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {Name: {_eq: null}}) {
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": nil,
							"Age":  int64(60),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithStringFilterBlockAndSelect_SelectSameFieldAsFilterWithMatch(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
						"Name": "John",
						"Age": 21
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
						Users(filter: {Name: {_eq: "John"}}) {
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
func TestQuerySimpleWithStringFilterBlockAndSelect_SelectDifferentFieldThanFilterWithMatch(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
						"Name": "John",
						"Age": 21
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
						Users(filter: {Name: {_eq: "John"}}) {
							Age
						}
					}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": int64(21),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
func TestQuerySimpleWithStringFilterBlockAndSelect_SelectMultipleFieldsButNoMatch(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
						"Name": "John",
						"Age": 21
					}`,
			},
			&action.Request{
				Request: `query {
						Users(filter: {Name: {_eq: "Bob"}}) {
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
