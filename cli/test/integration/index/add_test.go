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

func TestIndexAdd_WithSingleField_ShouldSucceed(t *testing.T) {
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
			&action.IndexAdd{
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

func TestIndexAdd_WithMultipleFieldsAndOrders_ShouldSucceed(t *testing.T) {
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
			&action.IndexAdd{
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

func TestIndexAdd_WithUniqueFlag_ShouldCreateUniqueIndex(t *testing.T) {
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
			&action.IndexAdd{
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

func TestIndexAdd_WithoutName_ShouldGenerateName(t *testing.T) {
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
			&action.IndexAdd{
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

func TestIndexAdd_WithUnknownCollection_ShouldReturnError(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.IndexAdd{
				Collection:  "NonExistentCollection",
				Name:        "TestIndex",
				Fields:      []string{"field1"},
				ExpectError: "collection not found",
			},
		},
	}

	test.Execute(t)
}

func TestIndexAdd_WithoutCollection_ShouldReturnError(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.IndexAdd{
				// Collection is empty
				Name:        "TestIndex",
				Fields:      []string{"field1"},
				ExpectError: "collection not found",
			},
		},
	}

	test.Execute(t)
}

func TestIndexAdd_WithoutFields_ShouldReturnError(t *testing.T) {
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
			&action.IndexAdd{
				Collection: "User",
				Name:       "EmptyIndex",
				// Fields is empty
				ExpectError: "index missing fields",
			},
		},
	}

	test.Execute(t)
}

func TestIndexAdd_WithInvalidFieldOrder_ShouldReturnError(t *testing.T) {
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
			&action.IndexAdd{
				Collection:  "User",
				Name:        "InvalidOrderIndex",
				Fields:      []string{"name:INVALID"},
				ExpectError: "invalid order: expected ASC or DESC",
			},
		},
	}

	test.Execute(t)
}

func TestIndexAdd_WithNonExistentField_ShouldReturnError(t *testing.T) {
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
			&action.IndexAdd{
				Collection:  "User",
				Name:        "InvalidFieldIndex",
				Fields:      []string{"nonexistent"},
				ExpectError: "adding an index on a non-existing property",
			},
		},
	}

	test.Execute(t)
}

func TestIndexAdd_WithDuplicateName_ShouldReturnError(t *testing.T) {
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
			&action.IndexAdd{
				Collection: "User",
				Name:       "DuplicateIndex",
				Fields:     []string{"name"},
			},
			&action.IndexAdd{
				Collection:  "User",
				Name:        "DuplicateIndex",
				Fields:      []string{"age"},
				ExpectError: "already exists",
			},
		},
	}

	test.Execute(t)
}
