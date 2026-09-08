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

package field_kinds

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationAddFieldKinds_WithString(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.Request{
				Request: `query {
					User {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "John",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNillableString_Nil(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": nil,
				},
			},
			&action.Request{
				Request: `query {
					User {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNonNillableString(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String!
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.Request{
				Request: `query {
					User {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "John",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNillableStringArray_FromTypedSlice(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						preferredStrings: [String]
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"preferredStrings": []string{"first", "second", "third"},
				},
			},
			&action.Request{
				Request: `query {
					User {
						preferredStrings
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"preferredStrings": []immutable.Option[string]{
								immutable.Some("first"),
								immutable.Some("second"),
								immutable.Some("third"),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNillableStringArray_FromAnySlice(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						preferredStrings: [String]
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"preferredStrings": []any{"first", nil, "third"},
				},
			},
			&action.Request{
				Request: `query {
					User {
						preferredStrings
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"preferredStrings": []immutable.Option[string]{
								immutable.Some("first"),
								immutable.None[string](),
								immutable.Some("third"),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNonNillableStringArray_FromTypedSlice(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						preferredStrings: [String!]
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"preferredStrings": []string{"first", "second", "third"},
				},
			},
			&action.Request{
				Request: `query {
					User {
						preferredStrings
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"preferredStrings": []string{"first", "second", "third"},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNonNillableStringArray_FromAnySlice(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						preferredStrings: [String!]
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"preferredStrings": []any{"first", "second", "third"},
				},
			},
			&action.Request{
				Request: `query {
					User {
						preferredStrings
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"preferredStrings": []string{"first", "second", "third"},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
