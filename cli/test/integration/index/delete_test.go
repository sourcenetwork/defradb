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

	"github.com/sourcenetwork/defradb/cli/test/action"
	"github.com/sourcenetwork/defradb/cli/test/integration"
	"github.com/sourcenetwork/defradb/client"
)

func TestIndexDelete_WithExistingIndex_ShouldSucceed(t *testing.T) {
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
			},
			&action.ListIndexes{
				Collection: "User",
				ExpectedIndexes: []client.IndexDescription{
					{
						Name: "UsersByName",
						Fields: []client.IndexedFieldDescription{
							{Name: "name", Descending: false},
						},
						Unique: false,
					},
				},
			},
			&action.DeleteIndex{
				Collection: "User",
				Name:       "UsersByName",
			},
			&action.ListIndexes{
				Collection:      "User",
				ExpectedIndexes: []client.IndexDescription{},
			},
		},
	}

	test.Execute(t)
}

func TestIndexDelete_WithUnknownCollection_ShouldReturnError(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.DeleteIndex{
				Collection:  "NonExistentCollection",
				Name:        "SomeIndex",
				ExpectError: "collection not found",
			},
		},
	}

	test.Execute(t)
}

func TestIndexDelete_WithoutCollection_ShouldReturnError(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.DeleteIndex{
				// Collection is empty
				Name:        "SomeIndex",
				ExpectError: "collection not found",
			},
		},
	}

	test.Execute(t)
}

func TestIndexDelete_WithoutName_ShouldReturnError(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.AddCollection{
				InlineSDL: `
					type User {
						name: String
					}
				`,
			},
			&action.DeleteIndex{
				Collection: "User",
				// Name is empty
				ExpectError: "index name is required",
			},
		},
	}

	test.Execute(t)
}

func TestIndexDelete_WithNonExistentIndex_ShouldReturnError(t *testing.T) {
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
			&action.DeleteIndex{
				Collection:  "User",
				Name:        "NonExistentIndex",
				ExpectError: "index with name doesn't exists",
			},
		},
	}

	test.Execute(t)
}

func TestIndexDelete_WithMultipleIndexes_ShouldDropOnlySpecified(t *testing.T) {
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
			// Create multiple indexes
			&action.NewIndex{
				Collection: "User",
				Name:       "UsersByName",
				Fields:     []string{"name"},
			},
			&action.NewIndex{
				Collection: "User",
				Name:       "UsersByAge",
				Fields:     []string{"age"},
			},
			&action.NewIndex{
				Collection: "User",
				Name:       "UsersByEmail",
				Fields:     []string{"email"},
				Unique:     true,
			},
			// Verify all indexes exist
			&action.ListIndexes{
				Collection: "User",
				ExpectedIndexes: []client.IndexDescription{
					{
						Name: "UsersByName",
						Fields: []client.IndexedFieldDescription{
							{Name: "name", Descending: false},
						},
						Unique: false,
					},
					{
						Name: "UsersByAge",
						Fields: []client.IndexedFieldDescription{
							{Name: "age", Descending: false},
						},
						Unique: false,
					},
					{
						Name: "UsersByEmail",
						Fields: []client.IndexedFieldDescription{
							{Name: "email", Descending: false},
						},
						Unique: true,
					},
				},
			},
			// Delete one index
			&action.DeleteIndex{
				Collection: "User",
				Name:       "UsersByAge",
			},
			// Verify only two indexes remain
			&action.ListIndexes{
				Collection: "User",
				ExpectedIndexes: []client.IndexDescription{
					{
						Name: "UsersByName",
						Fields: []client.IndexedFieldDescription{
							{Name: "name", Descending: false},
						},
						Unique: false,
					},
					{
						Name: "UsersByEmail",
						Fields: []client.IndexedFieldDescription{
							{Name: "email", Descending: false},
						},
						Unique: true,
					},
				},
			},
		},
	}

	test.Execute(t)
}
