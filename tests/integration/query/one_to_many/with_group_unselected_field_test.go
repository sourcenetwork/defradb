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

package one_to_many

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

var booksByThreeAuthors = []any{
	&action.AddDoc{
		CollectionID: 0,
		Doc: `{
			"name": "Painted House",
			"rating": 4.9,
			"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
		}`,
	},
	&action.AddDoc{
		CollectionID: 0,
		Doc: `{
			"name": "A Time for Mercy",
			"rating": 4.5,
			"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
		}`,
	},
	&action.AddDoc{
		CollectionID: 0,
		Doc: `{
			"name": "Candide",
			"rating": 4.95,
			"_authorID": "bae-b9c6cd5a-a931-5984-994d-7c435baa9f32"
		}`,
	},
	&action.AddDoc{
		CollectionID: 0,
		Doc: `{
			"name": "Zadig",
			"rating": 4.91,
			"_authorID": "bae-b9c6cd5a-a931-5984-994d-7c435baa9f32"
		}`,
	},
	&action.AddDoc{
		CollectionID: 1,
		Doc: `{
			"name": "John Grisham",
			"age": 65,
			"verified": true
		}`,
	},
	&action.AddDoc{
		CollectionID: 1,
		Doc: `{
			"name": "Voltaire",
			"age": 327,
			"verified": true
		}`,
	},
}

// Regression test for https://github.com/sourcenetwork/defradb/issues/4954 for the related-id
// field. Grouping by the foreign-key id must still occur even when the id is not rendered.
func TestQueryOneToManyWithGroupByRelatedIDWithoutRenderedGroupField(t *testing.T) {
	test := testUtils.TestCase{
		Actions: append(
			append([]any{}, booksByThreeAuthors...),
			&action.Request{
				Request: `query {
					Book(groupBy: [_authorID]) {
						GROUP {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"GROUP": []map[string]any{
								{"name": "Painted House"},
								{"name": "A Time for Mercy"},
							},
						},
						{
							"GROUP": []map[string]any{
								{"name": "Candide"},
								{"name": "Zadig"},
							},
						},
					},
				},
				NonOrderedResults: true,
			},
		),
	}

	executeTestCase(t, test)
}

// Grouping by the relation object field name while also rendering the foreign-key id. This
// exercises the path where the fetched group-by dependency and an explicitly selected field
// resolve to the same underlying field, ensuring the id is rendered exactly once per group.
func TestQueryOneToManyWithGroupByRelationObjectRenderingRelatedID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: append(
			append([]any{}, booksByThreeAuthors...),
			&action.Request{
				Request: `query {
					Book(groupBy: [author]) {
						_authorID
						GROUP {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab",
							"GROUP": []map[string]any{
								{"name": "Painted House"},
								{"name": "A Time for Mercy"},
							},
						},
						{
							"_authorID": "bae-b9c6cd5a-a931-5984-994d-7c435baa9f32",
							"GROUP": []map[string]any{
								{"name": "Candide"},
								{"name": "Zadig"},
							},
						},
					},
				},
				NonOrderedResults: true,
			},
		),
	}

	executeTestCase(t, test)
}

// As above but grouping by the relation object field name (which is internally remapped to its
// foreign-key id) rather than the id field directly.
func TestQueryOneToManyWithGroupByRelationObjectWithoutRenderedGroupField(t *testing.T) {
	test := testUtils.TestCase{
		Actions: append(
			append([]any{}, booksByThreeAuthors...),
			&action.Request{
				Request: `query {
					Book(groupBy: [author]) {
						GROUP {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"GROUP": []map[string]any{
								{"name": "Painted House"},
								{"name": "A Time for Mercy"},
							},
						},
						{
							"GROUP": []map[string]any{
								{"name": "Candide"},
								{"name": "Zadig"},
							},
						},
					},
				},
				NonOrderedResults: true,
			},
		),
	}

	executeTestCase(t, test)
}
