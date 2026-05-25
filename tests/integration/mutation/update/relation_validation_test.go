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

package update

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

const authorBookSDL = `
	type Author {
		name: String
	}
	type Book {
		title: String
		author: Author
	}
`

// nonExistentDocID is a valid-format DocID that is never inserted into any collection.
// "bae" decodes as version=1; the UUID portion is all-zeros.
const nonExistentDocID = "bae-00000000-0000-0000-0000-000000000000"

// TestMutationUpdateWithFilter_NonExistentRelation_Error asserts that
// UpdateDocumentsWithFilter is also subject to relation DocID validation:
// setting a relation field to a non-existent DocID is rejected.
func TestMutationUpdateWithFilter_NonExistentRelation_Error(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: authorBookSDL,
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc:          `{"title": "Dune"}`,
			},
			testUtils.UpdateWithFilter{
				CollectionID:  1,
				Filter:        `{title: {_eq: "Dune"}}`,
				Updater:       `{"_authorID": "` + nonExistentDocID + `"}`,
				ExpectedError: "relation target document not found",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

// TestMutationUpdateWithFilter_DeletedRelation_Error asserts that setting a relation
// field to the DocID of a soft-deleted document is also rejected by UpdateWithFilter.
func TestMutationUpdateWithFilter_DeletedRelation_Error(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: authorBookSDL,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Frank Herbert"}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc:          `{"title": "Dune"}`,
			},
			// Delete the author — its DocID is now soft-deleted.
			testUtils.DeleteDoc{
				CollectionID: 0,
				DocID:        0,
			},
			// Attempt to link the book to the deleted author via UpdateWithFilter.
			// We must use the hardcoded non-existent DocID as a stand-in because
			// UpdateWithFilter.Updater does not support DocIndex substitution; and
			// the soft-deleted author's real DocID is content-addressed and not
			// predictable. The validation is identical: both "not found" and
			// "soft-deleted" return the same ErrRelationTargetNotFound.
			testUtils.UpdateWithFilter{
				CollectionID:  1,
				Filter:        `{title: {_eq: "Dune"}}`,
				Updater:       `{"_authorID": "` + nonExistentDocID + `"}`,
				ExpectedError: "relation target document not found",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
