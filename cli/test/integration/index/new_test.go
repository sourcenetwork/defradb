// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package index

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/cli/test/action"
	"github.com/sourcenetwork/defradb/cli/test/integration"
	"github.com/sourcenetwork/defradb/client"
)

func withIndexArgs(index *action.NewIndex, args ...string) *action.NewIndex {
	index.AddArgs(args...)
	return index
}

func TestIndexNew_WithSingleField_ShouldSucceed(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.AddCollection{
				InlineSDL: `
					type User {
						name: String
						age: Int
						email: String
					}
				`,
			},
			&action.NewIndex{
				Collection: "User",
				Name:       "UsersByName",
				Fields:     []string{"name"},
				Expected: immutable.Some(client.IndexDescription{
					Name: "UsersByName",
					Fields: []client.IndexedFieldDescription{
						{Name: "name", Descending: false},
					},
					Unique: false,
				}),
			},
		},
	}

	test.Execute(t)
}

func TestIndexNew_WithMultipleFieldsAndOrders_ShouldSucceed(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.AddCollection{
				InlineSDL: `
					type User {
						name: String
						age: Int
						email: String
					}
				`,
			},
			&action.NewIndex{
				Collection: "User",
				Name:       "UsersByNameAndAge",
				Fields:     []string{"name:ASC", "age:DESC"},
				Expected: immutable.Some(client.IndexDescription{
					Name: "UsersByNameAndAge",
					Fields: []client.IndexedFieldDescription{
						{Name: "name", Descending: false},
						{Name: "age", Descending: true},
					},
					Unique: false,
				}),
			},
		},
	}

	test.Execute(t)
}

func TestIndexNew_WithUniqueFlag_ShouldMakeNewUniqueIndex(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.AddCollection{
				InlineSDL: `
					type User {
						name: String
						age: Int
						email: String
					}
				`,
			},
			&action.NewIndex{
				Collection: "User",
				Name:       "UniqueEmail",
				Fields:     []string{"email"},
				Unique:     true,
				Expected: immutable.Some(client.IndexDescription{
					Name: "UniqueEmail",
					Fields: []client.IndexedFieldDescription{
						{Name: "email", Descending: false},
					},
					Unique: true,
				}),
			},
		},
	}

	test.Execute(t)
}

func TestIndexNew_WithoutName_ShouldGenerateName(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.AddCollection{
				InlineSDL: `
					type User {
						name: String
						age: Int
						email: String
					}
				`,
			},
			&action.NewIndex{
				Collection: "User",
				Fields:     []string{"age"},
				Expected: immutable.Some(client.IndexDescription{
					// Name will be auto-generated, so we don't check it
					Fields: []client.IndexedFieldDescription{
						{Name: "age", Descending: false},
					},
					Unique: false,
				}),
			},
		},
	}

	test.Execute(t)
}

func TestIndexNew_WithUnknownCollection_ShouldReturnError(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.NewIndex{
				Collection:  "NonExistentCollection",
				Name:        "TestIndex",
				Fields:      []string{"field1"},
				ExpectError: "collection not found",
			},
		},
	}

	test.Execute(t)
}

func TestIndexNew_WithoutCollection_ShouldReturnError(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.NewIndex{
				// Collection is empty
				Name:        "TestIndex",
				Fields:      []string{"field1"},
				ExpectError: "collection not found",
			},
		},
	}

	test.Execute(t)
}

func TestIndexNew_WithoutFields_ShouldReturnError(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.AddCollection{
				InlineSDL: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			&action.NewIndex{
				Collection: "User",
				Name:       "EmptyIndex",
				// Fields is empty
				ExpectError: "index missing fields",
			},
		},
	}

	test.Execute(t)
}

func TestIndexNew_WithInvalidFieldOrder_ShouldReturnError(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.AddCollection{
				InlineSDL: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			&action.NewIndex{
				Collection:  "User",
				Name:        "InvalidOrderIndex",
				Fields:      []string{"name:INVALID"},
				ExpectError: "invalid order: expected ASC or DESC",
			},
		},
	}

	test.Execute(t)
}

func TestIndexNew_WithNonExistentField_ShouldReturnError(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.AddCollection{
				InlineSDL: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			&action.NewIndex{
				Collection:  "User",
				Name:        "InvalidFieldIndex",
				Fields:      []string{"nonexistent"},
				ExpectError: "making a new index on a non-existing property",
			},
		},
	}

	test.Execute(t)
}

func TestIndexNew_WithDuplicateName_ShouldReturnError(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.AddCollection{
				InlineSDL: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			&action.NewIndex{
				Collection: "User",
				Name:       "DuplicateIndex",
				Fields:     []string{"name"},
			},
			&action.NewIndex{
				Collection:  "User",
				Name:        "DuplicateIndex",
				Fields:      []string{"age"},
				ExpectError: "already exists",
			},
		},
	}

	test.Execute(t)
}

func TestIndexNew_WithFullTextFlag_ShouldCreateTypedIndex(t *testing.T) {
	test := &integration.Test{Actions: []action.Action{
		&action.AddCollection{InlineSDL: `type Article { body: String }`},
		withIndexArgs(&action.NewIndex{
			Collection: "Article",
			Fields:     []string{"body"},
			Expected: immutable.Some(client.IndexDescription{
				Kind: client.IndexKindFullText,
				KindDescription: &client.FullTextIndexDescription{
					Algorithm: client.FullTextAlgorithmBM25,
					BM25:      &client.BM25Params{K1: 2, B: 0.5},
				},
				Fields: []client.IndexedFieldDescription{{Name: "body"}},
			}),
		}, "--full-text", `{"Algorithm":"BM25","BM25":{"K1":2,"B":0.5}}`),
	}}

	test.Execute(t)
}

func TestIndexNew_WithTrigramFlag_ShouldCreateTypedIndex(t *testing.T) {
	test := &integration.Test{Actions: []action.Action{
		&action.AddCollection{InlineSDL: `type User { name: String }`},
		withIndexArgs(&action.NewIndex{
			Collection: "User",
			Fields:     []string{"name"},
			Expected: immutable.Some(client.IndexDescription{
				Kind:            client.IndexKindTrigram,
				KindDescription: &client.TrigramIndexDescription{},
				Fields:          []client.IndexedFieldDescription{{Name: "name"}},
			}),
		}, "--trigram"),
	}}

	test.Execute(t)
}

func TestIndexNew_WithInvalidFullTextJSON_ShouldReturnError(t *testing.T) {
	test := &integration.Test{Actions: []action.Action{
		&action.AddCollection{InlineSDL: `type Article { body: String }`},
		withIndexArgs(&action.NewIndex{
			Collection:  "Article",
			Fields:      []string{"body"},
			ExpectError: "invalid full-text index config",
		}, "--full-text", `{not-json}`),
	}}

	test.Execute(t)
}
