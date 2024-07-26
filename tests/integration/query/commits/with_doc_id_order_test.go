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
						commits(docID: "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3", order: {height: DESC}) {
							cid
							height
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid":    "bafyreiay56ley5dvsptso37fsonfcrtbuphwlfhi67d2y52vzzexba6vua",
							"height": int64(2),
						},
						{
							"cid":    "bafyreicsavx5oblk6asfoqyssz4ge2gf5ekfouvi7o6l7adly275op5oje",
							"height": int64(2),
						},
						{
							"cid":    "bafyreifzyy7bmpx2eywj4lznxzrzrvh6vrz6l7bhthkpexdq3wtho3vz6i",
							"height": int64(1),
						},
						{
							"cid":    "bafyreic2sba5sffkfnt32wfeoaw4qsqozjb5acwwtouxuzllb3aymjwute",
							"height": int64(1),
						},
						{
							"cid":    "bafyreihv7jqe32wsuff5vwzlp7izoo6pqg6kgqf5edknp3mqm3344gu35q",
							"height": int64(1),
						},
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
						commits(docID: "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3", order: {height: ASC}) {
							cid
							height
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid":    "bafyreifzyy7bmpx2eywj4lznxzrzrvh6vrz6l7bhthkpexdq3wtho3vz6i",
							"height": int64(1),
						},
						{
							"cid":    "bafyreic2sba5sffkfnt32wfeoaw4qsqozjb5acwwtouxuzllb3aymjwute",
							"height": int64(1),
						},
						{
							"cid":    "bafyreihv7jqe32wsuff5vwzlp7izoo6pqg6kgqf5edknp3mqm3344gu35q",
							"height": int64(1),
						},
						{
							"cid":    "bafyreiay56ley5dvsptso37fsonfcrtbuphwlfhi67d2y52vzzexba6vua",
							"height": int64(2),
						},
						{
							"cid":    "bafyreicsavx5oblk6asfoqyssz4ge2gf5ekfouvi7o6l7adly275op5oje",
							"height": int64(2),
						},
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
						commits(docID: "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3", order: {cid: DESC}) {
							cid
							height
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid":    "bafyreihv7jqe32wsuff5vwzlp7izoo6pqg6kgqf5edknp3mqm3344gu35q",
							"height": int64(1),
						},
						{
							"cid":    "bafyreifzyy7bmpx2eywj4lznxzrzrvh6vrz6l7bhthkpexdq3wtho3vz6i",
							"height": int64(1),
						},
						{
							"cid":    "bafyreicsavx5oblk6asfoqyssz4ge2gf5ekfouvi7o6l7adly275op5oje",
							"height": int64(2),
						},
						{
							"cid":    "bafyreic2sba5sffkfnt32wfeoaw4qsqozjb5acwwtouxuzllb3aymjwute",
							"height": int64(1),
						},
						{
							"cid":    "bafyreiay56ley5dvsptso37fsonfcrtbuphwlfhi67d2y52vzzexba6vua",
							"height": int64(2),
						},
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
						commits(docID: "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3", order: {cid: ASC}) {
							cid
							height
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid":    "bafyreiay56ley5dvsptso37fsonfcrtbuphwlfhi67d2y52vzzexba6vua",
							"height": int64(2),
						},
						{
							"cid":    "bafyreic2sba5sffkfnt32wfeoaw4qsqozjb5acwwtouxuzllb3aymjwute",
							"height": int64(1),
						},
						{
							"cid":    "bafyreicsavx5oblk6asfoqyssz4ge2gf5ekfouvi7o6l7adly275op5oje",
							"height": int64(2),
						},
						{
							"cid":    "bafyreifzyy7bmpx2eywj4lznxzrzrvh6vrz6l7bhthkpexdq3wtho3vz6i",
							"height": int64(1),
						},
						{
							"cid":    "bafyreihv7jqe32wsuff5vwzlp7izoo6pqg6kgqf5edknp3mqm3344gu35q",
							"height": int64(1),
						},
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
						 commits(docID: "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3", order: {height: ASC}) {
							 cid
							 height
						 }
					 }`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid":    "bafyreifzyy7bmpx2eywj4lznxzrzrvh6vrz6l7bhthkpexdq3wtho3vz6i",
							"height": int64(1),
						},
						{
							"cid":    "bafyreic2sba5sffkfnt32wfeoaw4qsqozjb5acwwtouxuzllb3aymjwute",
							"height": int64(1),
						},
						{
							"cid":    "bafyreihv7jqe32wsuff5vwzlp7izoo6pqg6kgqf5edknp3mqm3344gu35q",
							"height": int64(1),
						},
						{
							"cid":    "bafyreiay56ley5dvsptso37fsonfcrtbuphwlfhi67d2y52vzzexba6vua",
							"height": int64(2),
						},
						{
							"cid":    "bafyreicsavx5oblk6asfoqyssz4ge2gf5ekfouvi7o6l7adly275op5oje",
							"height": int64(2),
						},
						{
							"cid":    "bafyreibpiyrugj4gku336wp5lvcw3fgyxqpjvugm3t4z7v5h3ulwxs3x2y",
							"height": int64(3),
						},
						{
							"cid":    "bafyreieydjk3sqrxs5aqhsiy7ct25vu5qtbtpmzbytzee4apeidx6dq7je",
							"height": int64(3),
						},
						{
							"cid":    "bafyreic6rjkn7qsoxpboviode2l64ahg4yajsrb3p25zeooisnaxcweccu",
							"height": int64(4),
						},
						{
							"cid":    "bafyreieifkfzufdvlvni4o5pbdtuvm3w6x4fnqyelyq2owvsliiwjvddpi",
							"height": int64(4),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
