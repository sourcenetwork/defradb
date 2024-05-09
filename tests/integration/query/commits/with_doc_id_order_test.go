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

func TestQueryCommitsWithDocIDAndOrderHeightDesc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID, order height desc",
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
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {height: DESC}) {
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
						"cid":    "bafyreif3pvxatyqbmwllb7mcxvs734fgfmdkavbu6ambhay37w6vxjzkx4",
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
						"cid":    "bafyreid7n6a673spwjwl3ogtuqmrba4i4ntjqvsu4l3spqe6qutdtnqwlq",
						"height": int64(1),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocIDAndOrderHeightAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID, order height asc",
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
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {height: ASC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafyreih5awhipv4pk7truqm3pyyhle7xersbiyzyyacud6c3f7urzutpui",
						"height": int64(1),
					},
					{
						"cid":    "bafyreidcls23tu7qwp4siw3avyb42eukovpxg6dqifqruvy5wyc6b2ovvq",
						"height": int64(1),
					},
					{
						"cid":    "bafyreid7n6a673spwjwl3ogtuqmrba4i4ntjqvsu4l3spqe6qutdtnqwlq",
						"height": int64(1),
					},
					{
						"cid":    "bafyreictnkwvit6jp4mwhai3xp75nvtacxpq6zgbbjm55ylae3t6qshrze",
						"height": int64(2),
					},
					{
						"cid":    "bafyreif3pvxatyqbmwllb7mcxvs734fgfmdkavbu6ambhay37w6vxjzkx4",
						"height": int64(2),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocIDAndOrderCidDesc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID, order cid desc",
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
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {cid: DESC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafyreih5awhipv4pk7truqm3pyyhle7xersbiyzyyacud6c3f7urzutpui",
						"height": int64(1),
					},
					{
						"cid":    "bafyreif3pvxatyqbmwllb7mcxvs734fgfmdkavbu6ambhay37w6vxjzkx4",
						"height": int64(2),
					},
					{
						"cid":    "bafyreidcls23tu7qwp4siw3avyb42eukovpxg6dqifqruvy5wyc6b2ovvq",
						"height": int64(1),
					},
					{
						"cid":    "bafyreid7n6a673spwjwl3ogtuqmrba4i4ntjqvsu4l3spqe6qutdtnqwlq",
						"height": int64(1),
					},
					{
						"cid":    "bafyreictnkwvit6jp4mwhai3xp75nvtacxpq6zgbbjm55ylae3t6qshrze",
						"height": int64(2),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocIDAndOrderCidAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID, order cid asc",
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
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {cid: ASC}) {
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
						"cid":    "bafyreid7n6a673spwjwl3ogtuqmrba4i4ntjqvsu4l3spqe6qutdtnqwlq",
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
						"cid":    "bafyreih5awhipv4pk7truqm3pyyhle7xersbiyzyyacud6c3f7urzutpui",
						"height": int64(1),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocIDAndOrderAndMultiUpdatesCidAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID, multiple updates with order cid asc",
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
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	23
				}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	24
				}`,
			},
			testUtils.Request{
				Request: `query {
						 commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {height: ASC}) {
							 cid
							 height
						 }
					 }`,
				Results: []map[string]any{
					{
						"cid":    "bafyreih5awhipv4pk7truqm3pyyhle7xersbiyzyyacud6c3f7urzutpui",
						"height": int64(1),
					},
					{
						"cid":    "bafyreidcls23tu7qwp4siw3avyb42eukovpxg6dqifqruvy5wyc6b2ovvq",
						"height": int64(1),
					},
					{
						"cid":    "bafyreid7n6a673spwjwl3ogtuqmrba4i4ntjqvsu4l3spqe6qutdtnqwlq",
						"height": int64(1),
					},
					{
						"cid":    "bafyreictnkwvit6jp4mwhai3xp75nvtacxpq6zgbbjm55ylae3t6qshrze",
						"height": int64(2),
					},
					{
						"cid":    "bafyreif3pvxatyqbmwllb7mcxvs734fgfmdkavbu6ambhay37w6vxjzkx4",
						"height": int64(2),
					},
					{
						"cid":    "bafyreial53rqep7uoheucc3rzvhs6dbhydnkbqe4w2bhd3hsybub4u3h6m",
						"height": int64(3),
					},
					{
						"cid":    "bafyreifyknutg2lsajcsqrfegr65t5h5s743jkp3bfuzphx4nmqztfwmga",
						"height": int64(3),
					},
					{
						"cid":    "bafyreiarqxac3gnfrfj3j6fukof375kp56i6jcvgmpfytenczsilx4xkey",
						"height": int64(4),
					},
					{
						"cid":    "bafyreihazmfdrwrk3udgnnkuxipcg7cjmzndfazif3t2vxyrq36qwfkrt4",
						"height": int64(4),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
