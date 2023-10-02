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

func TestQueryCommitsWithUnknownDockey(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with unknown dockey",
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
						commits(dockey: "unknown dockey") {
							cid
						}
					}`,
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDockey(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey",
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
						commits(dockey: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7") {
							cid
						}
					}`,
				Results: []map[string]any{
					{
						"cid": "bafybeic4x7hxoh7yhqmvo7c3mqoyv6j7lnnajkt2hzf2j3mjaf6wmwwl6u",
					},
					{
						"cid": "bafybeidd6rsya2q5gxaarx52da22ih5jdn5wgxsfehcuwquffgjvmdrh34",
					},
					{
						"cid": "bafybeiax37emgcmyjjsiae7kwqis675whyc73wth44amhcmsndfygfhl7m",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDockeyAndLinks(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, with links",
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
						commits(dockey: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7") {
							cid
							links {
								cid
								name
							}
						}
					}`,
				Results: []map[string]any{
					{
						"cid":   "bafybeic4x7hxoh7yhqmvo7c3mqoyv6j7lnnajkt2hzf2j3mjaf6wmwwl6u",
						"links": []map[string]any{},
					},
					{
						"cid":   "bafybeidd6rsya2q5gxaarx52da22ih5jdn5wgxsfehcuwquffgjvmdrh34",
						"links": []map[string]any{},
					},
					{
						"cid": "bafybeiax37emgcmyjjsiae7kwqis675whyc73wth44amhcmsndfygfhl7m",
						"links": []map[string]any{
							{
								"cid":  "bafybeic4x7hxoh7yhqmvo7c3mqoyv6j7lnnajkt2hzf2j3mjaf6wmwwl6u",
								"name": "age",
							},
							{
								"cid":  "bafybeidd6rsya2q5gxaarx52da22ih5jdn5wgxsfehcuwquffgjvmdrh34",
								"name": "name",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDockeyAndUpdate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, multiple results",
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
						commits(dockey: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7") {
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
						"cid":    "bafybeic4x7hxoh7yhqmvo7c3mqoyv6j7lnnajkt2hzf2j3mjaf6wmwwl6u",
						"height": int64(1),
					},
					{
						"cid":    "bafybeidd6rsya2q5gxaarx52da22ih5jdn5wgxsfehcuwquffgjvmdrh34",
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

// This test is for documentation reasons only. This is not
// desired behaviour (first results includes link._head, second
// includes link._Name).
func TestQueryCommitsWithDockeyAndUpdateAndLinks(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, multiple results and links",
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
						commits(dockey: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7") {
							cid
							links {
								cid
								name
							}
						}
					}`,
				Results: []map[string]any{
					{
						"cid": "bafybeicddzzjp4k6itagzpnsputz5pgq57bu4qpwvrzxq7qi2bwguvsine",
						"links": []map[string]any{
							{
								"cid":  "bafybeic4x7hxoh7yhqmvo7c3mqoyv6j7lnnajkt2hzf2j3mjaf6wmwwl6u",
								"name": "_head",
							},
						},
					},
					{
						"cid":   "bafybeic4x7hxoh7yhqmvo7c3mqoyv6j7lnnajkt2hzf2j3mjaf6wmwwl6u",
						"links": []map[string]any{},
					},
					{
						"cid":   "bafybeidd6rsya2q5gxaarx52da22ih5jdn5wgxsfehcuwquffgjvmdrh34",
						"links": []map[string]any{},
					},
					{
						"cid": "bafybeic2z67t72ty7op6aoqzpz7larpubb473naqipho7rftoivkmubh7a",
						"links": []map[string]any{
							{
								"cid":  "bafybeiax37emgcmyjjsiae7kwqis675whyc73wth44amhcmsndfygfhl7m",
								"name": "_head",
							},
							{
								"cid":  "bafybeicddzzjp4k6itagzpnsputz5pgq57bu4qpwvrzxq7qi2bwguvsine",
								"name": "age",
							},
						},
					},
					{
						"cid": "bafybeiax37emgcmyjjsiae7kwqis675whyc73wth44amhcmsndfygfhl7m",
						"links": []map[string]any{
							{
								"cid":  "bafybeic4x7hxoh7yhqmvo7c3mqoyv6j7lnnajkt2hzf2j3mjaf6wmwwl6u",
								"name": "age",
							},
							{
								"cid":  "bafybeidd6rsya2q5gxaarx52da22ih5jdn5wgxsfehcuwquffgjvmdrh34",
								"name": "name",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
