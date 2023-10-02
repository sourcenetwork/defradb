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

func TestQueryCommitsWithDockeyAndOrderHeightDesc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, order height desc",
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
						commits(dockey: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {height: DESC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeicddzzjp4k6itagzpnsputz5pgq57bu4qpwvrzxq7qi2bwguvsine",
						"height": int64(2),
					},
					{
						"cid":    "bafybeic2z67t72ty7op6aoqzpz7larpubb473naqipho7rftoivkmubh7a",
						"height": int64(2),
					},
					{
						"cid":    "bafybeic4x7hxoh7yhqmvo7c3mqoyv6j7lnnajkt2hzf2j3mjaf6wmwwl6u",
						"height": int64(1),
					},
					{
						"cid":    "bafybeidd6rsya2q5gxaarx52da22ih5jdn5wgxsfehcuwquffgjvmdrh34",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiax37emgcmyjjsiae7kwqis675whyc73wth44amhcmsndfygfhl7m",
						"height": int64(1),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDockeyAndOrderHeightAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, order height asc",
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
						commits(dockey: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {height: ASC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeic4x7hxoh7yhqmvo7c3mqoyv6j7lnnajkt2hzf2j3mjaf6wmwwl6u",
						"height": int64(1),
					},
					{
						"cid":    "bafybeidd6rsya2q5gxaarx52da22ih5jdn5wgxsfehcuwquffgjvmdrh34",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiax37emgcmyjjsiae7kwqis675whyc73wth44amhcmsndfygfhl7m",
						"height": int64(1),
					},
					{
						"cid":    "bafybeicddzzjp4k6itagzpnsputz5pgq57bu4qpwvrzxq7qi2bwguvsine",
						"height": int64(2),
					},
					{
						"cid":    "bafybeic2z67t72ty7op6aoqzpz7larpubb473naqipho7rftoivkmubh7a",
						"height": int64(2),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDockeyAndOrderCidDesc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, order cid desc",
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
						commits(dockey: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {cid: DESC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeidd6rsya2q5gxaarx52da22ih5jdn5wgxsfehcuwquffgjvmdrh34",
						"height": int64(1),
					},
					{
						"cid":    "bafybeicddzzjp4k6itagzpnsputz5pgq57bu4qpwvrzxq7qi2bwguvsine",
						"height": int64(2),
					},
					{
						"cid":    "bafybeic4x7hxoh7yhqmvo7c3mqoyv6j7lnnajkt2hzf2j3mjaf6wmwwl6u",
						"height": int64(1),
					},
					{
						"cid":    "bafybeic2z67t72ty7op6aoqzpz7larpubb473naqipho7rftoivkmubh7a",
						"height": int64(2),
					},
					{
						"cid":    "bafybeiax37emgcmyjjsiae7kwqis675whyc73wth44amhcmsndfygfhl7m",
						"height": int64(1),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDockeyAndOrderCidAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, order cid asc",
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
						commits(dockey: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {cid: ASC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeiax37emgcmyjjsiae7kwqis675whyc73wth44amhcmsndfygfhl7m",
						"height": int64(1),
					},
					{
						"cid":    "bafybeic2z67t72ty7op6aoqzpz7larpubb473naqipho7rftoivkmubh7a",
						"height": int64(2),
					},
					{
						"cid":    "bafybeic4x7hxoh7yhqmvo7c3mqoyv6j7lnnajkt2hzf2j3mjaf6wmwwl6u",
						"height": int64(1),
					},
					{
						"cid":    "bafybeicddzzjp4k6itagzpnsputz5pgq57bu4qpwvrzxq7qi2bwguvsine",
						"height": int64(2),
					},
					{
						"cid":    "bafybeidd6rsya2q5gxaarx52da22ih5jdn5wgxsfehcuwquffgjvmdrh34",
						"height": int64(1),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDockeyAndOrderAndMultiUpdatesCidAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, multiple updates with order cid asc",
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
						 commits(dockey: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {height: ASC}) {
							 cid
							 height
						 }
					 }`,
				Results: []map[string]any{
					{
						"cid":    "bafybeic4x7hxoh7yhqmvo7c3mqoyv6j7lnnajkt2hzf2j3mjaf6wmwwl6u",
						"height": int64(1),
					},
					{
						"cid":    "bafybeidd6rsya2q5gxaarx52da22ih5jdn5wgxsfehcuwquffgjvmdrh34",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiax37emgcmyjjsiae7kwqis675whyc73wth44amhcmsndfygfhl7m",
						"height": int64(1),
					},
					{
						"cid":    "bafybeicddzzjp4k6itagzpnsputz5pgq57bu4qpwvrzxq7qi2bwguvsine",
						"height": int64(2),
					},
					{
						"cid":    "bafybeic2z67t72ty7op6aoqzpz7larpubb473naqipho7rftoivkmubh7a",
						"height": int64(2),
					},
					{
						"cid":    "bafybeids3sq532txo55elkc5eba7ns65s5lfbkx5x7cg2sxtpm253cyw5i",
						"height": int64(3),
					},
					{
						"cid":    "bafybeidb3snylmwsqdrroaunenha7w2lvsxksxrpc7tljp6hmhulxghncy",
						"height": int64(3),
					},
					{
						"cid":    "bafybeieu4fr3imdr7ituubknfjw7kl2mxjhnitvmvb2u5moolrjeetyjq4",
						"height": int64(4),
					},
					{
						"cid":    "bafybeiakwirtapzmvz3qkussybgas3v3v7sot7mjmxakklxpqz4ixrkqgy",
						"height": int64(4),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
