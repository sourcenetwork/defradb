// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package issues

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSimple_WithSevenDummyTypesBefore(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Type0 {
						f: String
					}
					type Type1 {
						f: String
					}
					type Type2 {
						f: String
					}
					type Type3 {
						f: String
					}
					type Type4 {
						f: String
					}
					type Type5 {
						f: String
					}
					type Type6 {
						f: String
					}

					type User {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 7,
				DocMap: map[string]any{
					"name": "John",
				},
			},
			testUtils.Request{
				Request: `
					query {
						User {
							name
						}
					}
				`,
				Results: []map[string]any{
					{
						"name": "John",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSimple_WithEightDummyTypesBefore(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Type0 {
						f: String
					}
					type Type1 {
						f: String
					}
					type Type2 {
						f: String
					}
					type Type3 {
						f: String
					}
					type Type4 {
						f: String
					}
					type Type5 {
						f: String
					}
					type Type6 {
						f: String
					}
					type Type7 {
						f: String
					}

					type User {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 8,
				DocMap: map[string]any{
					"name": "John",
				},
			},
			testUtils.Request{
				Request: `
					query {
						User {
							name
						}
					}
				`,
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSimple_WithEightDummyTypesBeforeInSplitDeclaration(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Type0 {
						f: String
					}
					type Type1 {
						f: String
					}
					type Type2 {
						f: String
					}
					type Type3 {
						f: String
					}
					type Type4 {
						f: String
					}
					type Type5 {
						f: String
					}
					type Type6 {
						f: String
					}
					type Type7 {
						f: String
					}
				`,
			},
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 8,
				DocMap: map[string]any{
					"name": "John",
				},
			},
			testUtils.Request{
				Request: `
					query {
						User {
							name
						}
					}
				`,
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSimple_WithEightDummyTypesAfter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
					}

					type Type0 {
						f: String
					}
					type Type1 {
						f: String
					}
					type Type2 {
						f: String
					}
					type Type3 {
						f: String
					}
					type Type4 {
						f: String
					}
					type Type5 {
						f: String
					}
					type Type6 {
						f: String
					}
					type Type7 {
						f: String
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "John",
				},
			},
			testUtils.Request{
				Request: `
					query {
						User {
							name
						}
					}
				`,
				Results: []map[string]any{
					{
						"name": "John",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSimple_WithSevenDummyTypesBeforeAndOneAfter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Type0 {
						f: String
					}
					type Type1 {
						f: String
					}
					type Type2 {
						f: String
					}
					type Type3 {
						f: String
					}
					type Type4 {
						f: String
					}
					type Type5 {
						f: String
					}
					type Type6 {
						f: String
					}

					type User {
						name: String
					}

					type Type7 {
						f: String
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 7,
				DocMap: map[string]any{
					"name": "John",
				},
			},
			testUtils.Request{
				Request: `
					query {
						User {
							name
						}
					}
				`,
				Results: []map[string]any{
					{
						"name": "John",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
