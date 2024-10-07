// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package simple

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimple_WithGEOpOnJSONField_WithEqualValue_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with _ge JSON filter with equal value",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
						Custom: JSON
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Custom": 21
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Custom": 32
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {Custom: {_ge: 32}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Bob",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimple_WithGEOpOnJSONField_WithGreaterValue_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with _ge JSON filter with greater value",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
						Custom: JSON
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Custom": 21
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Custom": 32
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {Custom: {_ge: 31}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Bob",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimple_WithGEOpOnJSONField_WithNilValue_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic ge nil filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
						Custom: JSON
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Custom": 21
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {Custom: {_ge: null}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
						},
						{
							"Name": "Bob",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimple_WithGEOpOnJSONFieldNested_WithEqualValue_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with _ge JSON nested filter with equal value",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
						Custom: JSON
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Custom": {"age": 21}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Custom": {"age": 32}
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {Custom: {age: {_ge: 32}}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Bob",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// func TestQuerySimple_WithGEOpOnJSONField_WithGreaterValue_Succeeds(t *testing.T) {
// 	test := testUtils.TestCase{
// 		Description: "Simple query with _ge JSON filter with greater value",
// 		Actions: []any{
// 			testUtils.SchemaUpdate{
// 				Schema: `
// 					type Users {
// 						Name: String
// 						Custom: JSON
// 					}
// 				`,
// 			},
// 			testUtils.CreateDoc{
// 				Doc: `{
// 					"Name": "John",
// 					"Custom": 21
// 				}`,
// 			},
// 			testUtils.CreateDoc{
// 				Doc: `{
// 					"Name": "Bob",
// 					"Custom": 32
// 				}`,
// 			},
// 			testUtils.Request{
// 				Request: `query {
// 					Users(filter: {Custom: {_ge: 31}}) {
// 						Name
// 					}
// 				}`,
// 				Results: map[string]any{
// 					"Users": []map[string]any{
// 						{
// 							"Name": "Bob",
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}

// 	testUtils.ExecuteTestCase(t, test)
// }

// func TestQuerySimple_WithGEOpOnJSONField_WithNilValue_Succeeds(t *testing.T) {
// 	test := testUtils.TestCase{
// 		Description: "Simple query with basic ge nil filter",
// 		Actions: []any{
// 			testUtils.SchemaUpdate{
// 				Schema: `
// 					type Users {
// 						Name: String
// 						Custom: JSON
// 					}
// 				`,
// 			},
// 			testUtils.CreateDoc{
// 				Doc: `{
// 					"Name": "John",
// 					"Custom": 21
// 				}`,
// 			},
// 			testUtils.CreateDoc{
// 				Doc: `{
// 					"Name": "Bob"
// 				}`,
// 			},
// 			testUtils.Request{
// 				Request: `query {
// 					Users(filter: {Custom: {_ge: null}}) {
// 						Name
// 					}
// 				}`,
// 				Results: map[string]any{
// 					"Users": []map[string]any{
// 						{
// 							"Name": "John",
// 						},
// 						{
// 							"Name": "Bob",
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}

// 	testUtils.ExecuteTestCase(t, test)
// }
