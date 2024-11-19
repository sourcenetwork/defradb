// Copyright 2022 Democratized Data Foundation
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

func TestQuerySimple(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with no filter",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						_docID
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"_docID": "bae-d4303725-7db9-53d2-b324-f3ee44020e52",
							"Name":   "John",
							"Age":    int64(21),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithAlias(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with alias, no filter",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						username: Name
						age: Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"username": "John",
							"age":      int64(21),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithMultipleRows(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with no filter, multiple rows",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 27
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Bob",
							"Age":  int64(27),
						},
						{
							"Name": "John",
							"Age":  int64(21),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithUndefinedField(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query for undefined field",
		Actions: []any{
			testUtils.Request{
				Request: `query {
					Users {
						Name
						ThisFieldDoesNotExists
					}
				}`,
				ExpectedError: "Cannot query field \"ThisFieldDoesNotExists\" on type \"Users\".",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithSomeDefaultValues(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with some default-value fields",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						Name
						Email
						Age
						HeightM
						Verified
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":     "John",
							"Email":    nil,
							"Age":      nil,
							"HeightM":  nil,
							"Verified": nil,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithDefaultValue(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with default-value fields",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{ }`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						Name
						Email
						Age
						HeightM
						Verified
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":     nil,
							"Email":    nil,
							"Age":      nil,
							"HeightM":  nil,
							"Verified": nil,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

// This test is to ensure that deleted docs from the next collection ID are not returned in the query results.
// It documents the fixing of the bug described in #3242.
func TestQuerySimple_WithDeletedDocsInCollection2_ShouldNotYieldDeletedDocsOnCollection1Query(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Deleted docs in collection 2 should not yield deleted docs on collection 1 query",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
                    type User {
                        name: String
                    }
                    type Friend {
                        name: String
                    }
                `,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "Shahzad",
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "John",
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name": "Andy",
				},
			},
			testUtils.Request{
				Request: `query {
                    User {
                        _docID
                    }
                }`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"_docID": "bae-1ef746f8-821e-586f-99b2-4cb1fb9b782f",
						},
						{
							"_docID": "bae-22dacd35-4560-583a-9a80-8edbf28aa85c",
						},
					},
				},
			},
			testUtils.DeleteDoc{
				CollectionID: 1,
				DocID:        0,
			},
			testUtils.Request{
				Request: `query {
                    User {
                        _docID
                    }
                }`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"_docID": "bae-1ef746f8-821e-586f-99b2-4cb1fb9b782f",
						},
						{
							"_docID": "bae-22dacd35-4560-583a-9a80-8edbf28aa85c",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
