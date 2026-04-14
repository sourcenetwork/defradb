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

func TestQuerySimple(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Query a single document returns all requested scalar fields.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.Request{
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
							"_docID": "bae-619ea0d2-35ba-5e8c-ac4d-2b769937213b",
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
		Description: "Query with field aliases returns results under the aliased names.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.Request{
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
		Description: "Query with multiple documents returns all rows.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 27
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"Age":  int64(21),
						},
						{
							"Name": "Bob",
							"Age":  int64(27),
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
		Description: "Querying a field that does not exist on the type returns an error.",
		Actions: []any{
			&action.Request{
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
		Description: "Document created with only Name set returns nil for all other fields.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			&action.Request{
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
		Description: "Document created with empty body returns nil for every field.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{ }`,
			},
			&action.Request{
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
		Description: "Deleting docs from collection 2 does not affect results for collection 1.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
                    type User {
                        name: String
                    }
                    type Friend {
                        name: String
                    }
                `,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "Shahzad",
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name": "Andy",
				},
			},
			&action.Request{
				Request: `query {
                    User {
                        _docID
                    }
                }`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"_docID": testUtils.NewDocIndex(0, 1),
						},
						{
							"_docID": testUtils.NewDocIndex(0, 0),
						},
					},
				},
				NonOrderedResults: true,
			},
			testUtils.DeleteDoc{
				CollectionID: 1,
				DocID:        0,
			},
			&action.Request{
				Request: `query {
                    User {
                        _docID
                    }
                }`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"_docID": testUtils.NewDocIndex(0, 1),
						},
						{
							"_docID": testUtils.NewDocIndex(0, 0),
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
