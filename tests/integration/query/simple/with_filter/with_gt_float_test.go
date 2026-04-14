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

func TestQuerySimpleWithFloatGreaterThanFilterBlock_OneMatchingResult(t *testing.T) {
	test := testUtils.TestCase{
		Description: "_gt filter on a Float field returns one matching document that strictly exceeds the threshold.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
						"Name": "John",
						"HeightM": 2.1
					}`,
			},
			&action.AddDoc{
				Doc: `{
						"Name": "Bob",
						"HeightM": 1.82
					}`,
			},
			&action.Request{
				Request: `query {
						Users(filter: {HeightM: {_gt: 2.0999999999999}}) {
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

func TestQuerySimpleWithFloatGreaterThanFilterBlock_NoMatchingResult(t *testing.T) {
	test := testUtils.TestCase{
		Description: "_gt filter on a Float field with a threshold greater than all documents returns an empty result.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
						"Name": "John",
						"HeightM": 2.1
					}`,
			},
			&action.AddDoc{
				Doc: `{
						"Name": "Bob",
						"HeightM": 1.82
					}`,
			},
			&action.Request{
				Request: `query {
						Users(filter: {HeightM: {_gt: 40}}) {
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

func TestQuerySimpleWithFloatGreaterThanFilterBlock_AllMatchingResult(t *testing.T) {
	test := testUtils.TestCase{
		Description: "_gt filter on a Float field with a threshold below all documents returns all documents.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
						"Name": "John",
						"HeightM": 2.1
					}`,
			},
			&action.AddDoc{
				Doc: `{
						"Name": "Bob",
						"HeightM": 1.82
					}`,
			},
			&action.Request{
				Request: `query {
						Users(filter: {HeightM: {_gt: 1.8199999999999}}) {
							Name
						}
					}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Bob",
						},
						{
							"Name": "John",
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithFloatGreaterThanFilterBlockWithIntFilterValue(t *testing.T) {
	test := testUtils.TestCase{
		Description: "_gt filter on a Float field using an integer threshold correctly returns documents above it.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 2.1
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"HeightM": 1.82
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {HeightM: {_gt: 2}}) {
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

func TestQuerySimpleWithFloatGreaterThanFilterBlockWithNullFilterValue(t *testing.T) {
	test := testUtils.TestCase{
		Description: "_gt null filter on a Float field returns only documents that have a non-null float value.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 2.1
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob"
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {HeightM: {_gt: null}}) {
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
