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

// A bare, unquoted identifier in a comparison position names another field of the same
// document, e.g. `{chunkCount: {_lt: expectedChunks}}`.

var fileGQLSchema = `
	type Files {
		name: String
		chunkCount: Int
		expectedChunks: Int
	}
`

var indexedFileGQLSchema = `
	type Files {
		name: String
		chunkCount: Int @index
		expectedChunks: Int
	}
`

func fileDocs() []any {
	return []any{
		&action.AddDoc{
			Doc: `{
				"name": "behind",
				"chunkCount": 3,
				"expectedChunks": 10
			}`,
		},
		&action.AddDoc{
			Doc: `{
				"name": "complete",
				"chunkCount": 10,
				"expectedChunks": 10
			}`,
		},
		&action.AddDoc{
			Doc: `{
				"name": "overshot",
				"chunkCount": 12,
				"expectedChunks": 10
			}`,
		},
	}
}

func executeFileTestCase(t *testing.T, sdl string, request *action.Request) {
	actions := []any{&action.AddCollection{SDL: sdl}}
	actions = append(actions, fileDocs()...)
	actions = append(actions, request)

	testUtils.ExecuteTestCase(t, testUtils.TestCase{Actions: actions})
}

func TestQuerySimpleWithFieldReferenceInLessThanFilter(t *testing.T) {
	executeFileTestCase(t, fileGQLSchema, &action.Request{
		Request: `query {
			Files(filter: {chunkCount: {_lt: expectedChunks}}) {
				name
				chunkCount
				expectedChunks
			}
		}`,
		Results: map[string]any{
			"Files": []map[string]any{
				{
					"name":           "behind",
					"chunkCount":     int64(3),
					"expectedChunks": int64(10),
				},
			},
		},
	})
}

// The referenced field must be fetched even though it is not part of the selection set.
func TestQuerySimpleWithFieldReferenceNotInSelectionSet(t *testing.T) {
	executeFileTestCase(t, fileGQLSchema, &action.Request{
		Request: `query {
			Files(filter: {chunkCount: {_lt: expectedChunks}}) {
				name
			}
		}`,
		Results: map[string]any{
			"Files": []map[string]any{
				{
					"name": "behind",
				},
			},
		},
	})
}

func TestQuerySimpleWithFieldReferenceInEqualFilter(t *testing.T) {
	executeFileTestCase(t, fileGQLSchema, &action.Request{
		Request: `query {
			Files(filter: {chunkCount: {_eq: expectedChunks}}) {
				name
			}
		}`,
		Results: map[string]any{
			"Files": []map[string]any{
				{
					"name": "complete",
				},
			},
		},
	})
}

func TestQuerySimpleWithFieldReferenceInsideCompoundFilter(t *testing.T) {
	executeFileTestCase(t, fileGQLSchema, &action.Request{
		Request: `query {
			Files(filter: {_or: [
				{chunkCount: {_gt: expectedChunks}},
				{name: {_eq: "complete"}}
			]}) {
				name
			}
		}`,
		Results: map[string]any{
			"Files": []map[string]any{
				{
					"name": "complete",
				},
				{
					"name": "overshot",
				},
			},
		},
	})
}

// An index cannot build a range key from a field-to-field comparison, so it falls back to a
// scan.  The results must still be correct.
func TestQuerySimpleWithFieldReferenceOnIndexedField(t *testing.T) {
	executeFileTestCase(t, indexedFileGQLSchema, &action.Request{
		Request: `query {
			Files(filter: {chunkCount: {_lt: expectedChunks}}) {
				name
			}
		}`,
		Results: map[string]any{
			"Files": []map[string]any{
				{
					"name": "behind",
				},
			},
		},
	})
}

func TestQuerySimpleWithFieldReferenceToUnknownField(t *testing.T) {
	executeFileTestCase(t, fileGQLSchema, &action.Request{
		Request: `query {
			Files(filter: {chunkCount: {_lt: notAField}}) {
				name
			}
		}`,
		ExpectedError: `field or alias not found. Name: notAField`,
	})
}

func TestQuerySimpleWithFieldReferenceToIncompatibleType(t *testing.T) {
	executeFileTestCase(t, fileGQLSchema, &action.Request{
		Request: `query {
			Files(filter: {chunkCount: {_lt: name}}) {
				name
			}
		}`,
		ExpectedError: `unexpected type. Property: condition, Actual: string`,
	})
}

// A bare identifier is only meaningful in a comparison position.  Everywhere else the shared
// JSON scalar is still in use, where a bare identifier stays a plain string.
func TestQuerySimpleWithFieldReferenceInOrderAliasIsUnchanged(t *testing.T) {
	executeFileTestCase(t, fileGQLSchema, &action.Request{
		Request: `query {
			Files(order: {_alias: {renamed: DESC}}) {
				renamed: name
			}
		}`,
		Results: map[string]any{
			"Files": []map[string]any{
				{"renamed": "overshot"},
				{"renamed": "complete"},
				{"renamed": "behind"},
			},
		},
	})
}
