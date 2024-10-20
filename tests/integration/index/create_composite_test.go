// Copyright 2023 Democratized Data Foundation
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

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestCompositeIndexCreate_WhenCreated_CanRetrieve(t *testing.T) {
	test := testUtils.TestCase{
		Description: "create composite index and retrieve it",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						age: Int 
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Shahzad",
						"age":	22
					}`,
			},
			testUtils.CreateIndex{
				CollectionID: 0,
				IndexName:    "name_age_index",
				Fields:       []testUtils.IndexedField{{Name: "name"}, {Name: "age"}},
			},
			testUtils.GetIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name: "name_age_index",
						ID:   1,
						Fields: []client.IndexedFieldDescription{
							{
								Name: "name",
							},
							{
								Name: "age",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestCompositeIndexCreate_UsingObjectDirective_SetsDefaultDirection(t *testing.T) {
	test := testUtils.TestCase{
		Description: "create composite index using object directive sets default direction",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(direction: DESC, includes: [{field: "name"}, {field: "age"}]) {
						name: String
						age: Int 
					}
				`,
			},
			testUtils.GetIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						// this should be User_name_DESC
						Name: "User_name_ASC",
						ID:   1,
						Fields: []client.IndexedFieldDescription{
							{
								Name:       "name",
								Descending: true,
							},
							{
								Name:       "age",
								Descending: true,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestCompositeIndexCreate_UsingObjectDirective_OverridesDefaultDirection(t *testing.T) {
	test := testUtils.TestCase{
		Description: "create composite object using field directive overrides default direction",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(direction: DESC, includes: [{field: "name"}, {field: "age", direction: ASC}]) {
						name: String
						age: Int 
					}
				`,
			},
			testUtils.GetIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						// this should be User_name_DESC
						Name: "User_name_ASC",
						ID:   1,
						Fields: []client.IndexedFieldDescription{
							{
								Name:       "name",
								Descending: true,
							},
							{
								Name:       "age",
								Descending: false,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestCompositeIndexCreate_UsingFieldDirective_ImplicitlyAddsField(t *testing.T) {
	test := testUtils.TestCase{
		Description: "create composite index using field directive implicitly adds field",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String @index(includes: [{field: "age"}])
						age: Int 
					}
				`,
			},
			testUtils.GetIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name: "User_name_ASC",
						ID:   1,
						Fields: []client.IndexedFieldDescription{
							{
								Name: "name",
							},
							{
								Name: "age",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestCompositeIndexCreate_UsingFieldDirective_SetsDefaultDirection(t *testing.T) {
	test := testUtils.TestCase{
		Description: "create composite index using field directive sets default direction",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String @index(direction: DESC, includes: [{field: "age"}])
						age: Int 
					}
				`,
			},
			testUtils.GetIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						// this should be User_name_DESC
						Name: "User_name_ASC",
						ID:   1,
						Fields: []client.IndexedFieldDescription{
							{
								Name:       "name",
								Descending: true,
							},
							{
								Name:       "age",
								Descending: true,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestCompositeIndexCreate_UsingFieldDirective_OverridesDefaultDirection(t *testing.T) {
	test := testUtils.TestCase{
		Description: "create composite index using field directive overrides default direction",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String @index(direction: DESC, includes: [{field: "age", direction: ASC}])
						age: Int 
					}
				`,
			},
			testUtils.GetIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						// this should be User_name_DESC
						Name: "User_name_ASC",
						ID:   1,
						Fields: []client.IndexedFieldDescription{
							{
								Name:       "name",
								Descending: true,
							},
							{
								Name:       "age",
								Descending: false,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestCompositeIndexCreate_UsingFieldDirective_WithExplicitIncludes_RespectsOrder(t *testing.T) {
	test := testUtils.TestCase{
		Description: "create composite index using field directive with explicit includes respects order",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String @index(includes: [{field: "age"}, {field: "name"}])
						age: Int 
					}
				`,
			},
			testUtils.GetIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name: "User_age_ASC",
						ID:   1,
						Fields: []client.IndexedFieldDescription{
							{
								Name: "age",
							},
							{
								Name: "name",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
