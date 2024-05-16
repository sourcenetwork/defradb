// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package commits

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryCommitsWithUnknownDocID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with unknown document ID",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.Request{
				Request: `query {
						commits(docID: "unknown document ID") {
							cid
						}
					}`,
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.Request{
				Request: `query {
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7") {
							cid
						}
					}`,
				Results: []map[string]any{
					{
						"cid": "bafyreih5awhipv4pk7truqm3pyyhle7xersbiyzyyacud6c3f7urzutpui",
					},
					{
						"cid": "bafyreidcls23tu7qwp4siw3avyb42eukovpxg6dqifqruvy5wyc6b2ovvq",
					},
					{
						"cid": "bafyreid7n6a673spwjwl3ogtuqmrba4i4ntjqvsu4l3spqe6qutdtnqwlq",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocIDAndLinks(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID, with links",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.Request{
				Request: `query {
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7") {
							cid
							links {
								cid
								name
							}
						}
					}`,
				Results: []map[string]any{
					{
						"cid":   "bafyreih5awhipv4pk7truqm3pyyhle7xersbiyzyyacud6c3f7urzutpui",
						"links": []map[string]any{},
					},
					{
						"cid":   "bafyreidcls23tu7qwp4siw3avyb42eukovpxg6dqifqruvy5wyc6b2ovvq",
						"links": []map[string]any{},
					},
					{
						"cid": "bafyreid7n6a673spwjwl3ogtuqmrba4i4ntjqvsu4l3spqe6qutdtnqwlq",
						"links": []map[string]any{
							{
								"cid":  "bafyreidcls23tu7qwp4siw3avyb42eukovpxg6dqifqruvy5wyc6b2ovvq",
								"name": "name",
							},
							{
								"cid":  "bafyreih5awhipv4pk7truqm3pyyhle7xersbiyzyyacud6c3f7urzutpui",
								"name": "age",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocIDAndUpdate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID, multiple results",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	22
				}`,
			},
			testUtils.Request{
				Request: `query {
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7") {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafyreictnkwvit6jp4mwhai3xp75nvtacxpq6zgbbjm55ylae3t6qshrze",
						"height": int64(2),
					},
					{
						"cid":    "bafyreih5awhipv4pk7truqm3pyyhle7xersbiyzyyacud6c3f7urzutpui",
						"height": int64(1),
					},
					{
						"cid":    "bafyreidcls23tu7qwp4siw3avyb42eukovpxg6dqifqruvy5wyc6b2ovvq",
						"height": int64(1),
					},
					{
						"cid":    "bafyreif3pvxatyqbmwllb7mcxvs734fgfmdkavbu6ambhay37w6vxjzkx4",
						"height": int64(2),
					},
					{
						"cid":    "bafyreid7n6a673spwjwl3ogtuqmrba4i4ntjqvsu4l3spqe6qutdtnqwlq",
						"height": int64(1),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test is for documentation reasons only. This is not
// desired behaviour (first results includes link._head, second
// includes link._Name).
func TestQueryCommitsWithDocIDAndUpdateAndLinks(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID, multiple results and links",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	22
				}`,
			},
			testUtils.Request{
				Request: `query {
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7") {
							cid
							links {
								cid
								name
							}
						}
					}`,
				Results: []map[string]any{
					{
						"cid": "bafyreictnkwvit6jp4mwhai3xp75nvtacxpq6zgbbjm55ylae3t6qshrze",
						"links": []map[string]any{
							{
								"cid":  "bafyreih5awhipv4pk7truqm3pyyhle7xersbiyzyyacud6c3f7urzutpui",
								"name": "_head",
							},
						},
					},
					{
						"cid":   "bafyreih5awhipv4pk7truqm3pyyhle7xersbiyzyyacud6c3f7urzutpui",
						"links": []map[string]any{},
					},
					{
						"cid":   "bafyreidcls23tu7qwp4siw3avyb42eukovpxg6dqifqruvy5wyc6b2ovvq",
						"links": []map[string]any{},
					},
					{
						"cid": "bafyreif3pvxatyqbmwllb7mcxvs734fgfmdkavbu6ambhay37w6vxjzkx4",
						"links": []map[string]any{
							{
								"cid":  "bafyreid7n6a673spwjwl3ogtuqmrba4i4ntjqvsu4l3spqe6qutdtnqwlq",
								"name": "_head",
							},
							{
								"cid":  "bafyreictnkwvit6jp4mwhai3xp75nvtacxpq6zgbbjm55ylae3t6qshrze",
								"name": "age",
							},
						},
					},
					{
						"cid": "bafyreid7n6a673spwjwl3ogtuqmrba4i4ntjqvsu4l3spqe6qutdtnqwlq",
						"links": []map[string]any{
							{
								"cid":  "bafyreidcls23tu7qwp4siw3avyb42eukovpxg6dqifqruvy5wyc6b2ovvq",
								"name": "name",
							},
							{
								"cid":  "bafyreih5awhipv4pk7truqm3pyyhle7xersbiyzyyacud6c3f7urzutpui",
								"name": "age",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
