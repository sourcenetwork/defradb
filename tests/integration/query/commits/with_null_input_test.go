// Copyright 2024 Democratized Data Foundation
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

func TestQueryCommitsWithNullDepth(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with null depth",
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
						commits(depth: null) {
							cid
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid": "bafyreifzyy7bmpx2eywj4lznxzrzrvh6vrz6l7bhthkpexdq3wtho3vz6i",
						},
						{
							"cid": "bafyreic2sba5sffkfnt32wfeoaw4qsqozjb5acwwtouxuzllb3aymjwute",
						},
						{
							"cid": "bafyreihv7jqe32wsuff5vwzlp7izoo6pqg6kgqf5edknp3mqm3344gu35q",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithNullCID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with null cid",
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
						commits(cid: null) {
							cid
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid": "bafyreifzyy7bmpx2eywj4lznxzrzrvh6vrz6l7bhthkpexdq3wtho3vz6i",
						},
						{
							"cid": "bafyreic2sba5sffkfnt32wfeoaw4qsqozjb5acwwtouxuzllb3aymjwute",
						},
						{
							"cid": "bafyreihv7jqe32wsuff5vwzlp7izoo6pqg6kgqf5edknp3mqm3344gu35q",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithNullFieldID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with null fieldId",
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
						commits(fieldId: null) {
							cid
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid": "bafyreifzyy7bmpx2eywj4lznxzrzrvh6vrz6l7bhthkpexdq3wtho3vz6i",
						},
						{
							"cid": "bafyreic2sba5sffkfnt32wfeoaw4qsqozjb5acwwtouxuzllb3aymjwute",
						},
						{
							"cid": "bafyreihv7jqe32wsuff5vwzlp7izoo6pqg6kgqf5edknp3mqm3344gu35q",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithNullOrder(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with null order",
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
						commits(order: null) {
							cid
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid": "bafyreifzyy7bmpx2eywj4lznxzrzrvh6vrz6l7bhthkpexdq3wtho3vz6i",
						},
						{
							"cid": "bafyreic2sba5sffkfnt32wfeoaw4qsqozjb5acwwtouxuzllb3aymjwute",
						},
						{
							"cid": "bafyreihv7jqe32wsuff5vwzlp7izoo6pqg6kgqf5edknp3mqm3344gu35q",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithNullOrderField(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with null order field",
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
						commits(docID: null) {
							cid
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid": "bafyreifzyy7bmpx2eywj4lznxzrzrvh6vrz6l7bhthkpexdq3wtho3vz6i",
						},
						{
							"cid": "bafyreic2sba5sffkfnt32wfeoaw4qsqozjb5acwwtouxuzllb3aymjwute",
						},
						{
							"cid": "bafyreihv7jqe32wsuff5vwzlp7izoo6pqg6kgqf5edknp3mqm3344gu35q",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithNullLimit(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with null limit",
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
						commits(limit: null) {
							cid
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid": "bafyreifzyy7bmpx2eywj4lznxzrzrvh6vrz6l7bhthkpexdq3wtho3vz6i",
						},
						{
							"cid": "bafyreic2sba5sffkfnt32wfeoaw4qsqozjb5acwwtouxuzllb3aymjwute",
						},
						{
							"cid": "bafyreihv7jqe32wsuff5vwzlp7izoo6pqg6kgqf5edknp3mqm3344gu35q",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithNullOffset(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with null offset",
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
						commits(offset: null) {
							cid
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid": "bafyreifzyy7bmpx2eywj4lznxzrzrvh6vrz6l7bhthkpexdq3wtho3vz6i",
						},
						{
							"cid": "bafyreic2sba5sffkfnt32wfeoaw4qsqozjb5acwwtouxuzllb3aymjwute",
						},
						{
							"cid": "bafyreihv7jqe32wsuff5vwzlp7izoo6pqg6kgqf5edknp3mqm3344gu35q",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithNullGroupBy(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with null groupBy",
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
						commits(groupBy: null) {
							cid
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid": "bafyreifzyy7bmpx2eywj4lznxzrzrvh6vrz6l7bhthkpexdq3wtho3vz6i",
						},
						{
							"cid": "bafyreic2sba5sffkfnt32wfeoaw4qsqozjb5acwwtouxuzllb3aymjwute",
						},
						{
							"cid": "bafyreihv7jqe32wsuff5vwzlp7izoo6pqg6kgqf5edknp3mqm3344gu35q",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
