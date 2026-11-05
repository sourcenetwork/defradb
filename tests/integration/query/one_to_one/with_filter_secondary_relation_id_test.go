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

package one_to_one

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

const secondaryRelationIDSchema = `
	type Book {
		title: String
		publisher: Publisher
	}
	type Publisher {
		name: String
		book: Book @primary
	}
`

func secondaryRelationIDDocs() []any {
	return []any{
		&action.AddCollection{
			SDL: secondaryRelationIDSchema,
		},
		&action.AddDoc{
			CollectionID: 0,
			Doc:          `{"title": "Book With Publisher"}`,
		},
		&action.AddDoc{
			CollectionID: 0,
			Doc:          `{"title": "Orphan Book 1"}`,
		},
		&action.AddDoc{
			CollectionID: 0,
			Doc:          `{"title": "Orphan Book 2"}`,
		},
		&action.AddDoc{
			CollectionID: 1,
			DocMap: map[string]any{
				"name": "Publisher2020",
				"book": testUtils.NewDocIndex(0, 0),
			},
		},
	}
}

func TestQueryOneToOne_WithEqNullFilterOnSecondaryRelationID_ShouldReturnOrphans(t *testing.T) {
	test := testUtils.TestCase{
		Actions: append(
			secondaryRelationIDDocs(),
			&action.Request{
				Request: `query {
					Book(filter: {_publisherID: {_eq: null}}, order: {title: ASC}) {
						title
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{"title": "Orphan Book 1"},
						{"title": "Orphan Book 2"},
					},
				},
			},
		),
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOne_WithNeqNullFilterOnSecondaryRelationID_ShouldReturnRelated(t *testing.T) {
	test := testUtils.TestCase{
		Actions: append(
			secondaryRelationIDDocs(),
			&action.Request{
				Request: `query {
					Book(filter: {_publisherID: {_neq: null}}) {
						title
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{"title": "Book With Publisher"},
					},
				},
			},
		),
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOne_WithEqFilterOnSecondaryRelationID_ShouldReturnRelated(t *testing.T) {
	test := testUtils.TestCase{
		Actions: append(
			secondaryRelationIDDocs(),
			&action.Request{
				Request: `query {
					Book(filter: {_publisherID: {_eq: "{{.DocID1_0}}"}}) {
						title
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{"title": "Book With Publisher"},
					},
				},
			},
		),
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOne_WithEqNullFilterOnSecondaryRelationIDAndOtherField_ShouldReturnMatchingOrphan(t *testing.T) {
	test := testUtils.TestCase{
		Actions: append(
			secondaryRelationIDDocs(),
			&action.Request{
				Request: `query {
					Book(filter: {_publisherID: {_eq: null}, title: {_eq: "Orphan Book 2"}}) {
						title
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{"title": "Orphan Book 2"},
					},
				},
			},
		),
	}

	testUtils.ExecuteTestCase(t, test)
}
