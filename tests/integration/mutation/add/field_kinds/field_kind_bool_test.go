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

func TestMutationAddFieldKinds_WithBool(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						active: Boolean
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"active": true,
				},
			},
			&action.Request{
				Request: `query {
					User {
						active
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"active": true,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNillableBool_Nil(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						active: Boolean
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"active": nil,
				},
			},
			&action.Request{
				Request: `query {
					User {
						active
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"active": nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNonNillableBool(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						active: Boolean!
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"active": true,
				},
			},
			&action.Request{
				Request: `query {
					User {
						active
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"active": true,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNillableBoolArray_FromTypedSlice(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						flags: [Boolean]
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"flags": []bool{true, false, true},
				},
			},
			&action.Request{
				Request: `query {
					User {
						flags
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"flags": []immutable.Option[bool]{
								immutable.Some(true),
								immutable.Some(false),
								immutable.Some(true),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNillableBoolArray_WithNilElement(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						flags: [Boolean]
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"flags": []any{true, nil, false},
				},
			},
			&action.Request{
				Request: `query {
					User {
						flags
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"flags": []immutable.Option[bool]{
								immutable.Some(true),
								immutable.None[bool](),
								immutable.Some(false),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNonNillableBoolArray_FromTypedSlice(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						flags: [Boolean!]
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"flags": []bool{true, false, true},
				},
			},
			&action.Request{
				Request: `query {
					User {
						flags
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"flags": []bool{true, false, true},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNonNillableBoolArray_FromAnySlice(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						flags: [Boolean!]
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"flags": []any{true, false, true},
				},
			},
			&action.Request{
				Request: `query {
					User {
						flags
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"flags": []bool{true, false, true},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
