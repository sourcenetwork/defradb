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
						_commits(docID: "unknown document ID") {
							cid
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocID(t *testing.T) {
	test := testUtils.TestCase{
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
						_commits(docID: "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738") {
							cid
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid": "bafyreihakk5jjukb4fw7klfejdmniwhuscnckcjo677p3mtcxrdpiahuea",
						},
						{
							"cid": "bafyreihx4lnknvruc6vonsg3dvb3nnlsycwzbbkeulcutnzgidkzfvea64",
						},
						{
							"cid": "bafyreihpq4duzngkledmxkxx3jevlp2q4aimhmbjygpv5chmgbf6u2fsqm",
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocIDAndLinks(t *testing.T) {
	test := testUtils.TestCase{
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
						_commits(docID: "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738") {
							cid
							links {
								cid
								name
							}
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid":   "bafyreihakk5jjukb4fw7klfejdmniwhuscnckcjo677p3mtcxrdpiahuea",
							"links": []map[string]any{},
						},
						{
							"cid":   "bafyreihx4lnknvruc6vonsg3dvb3nnlsycwzbbkeulcutnzgidkzfvea64",
							"links": []map[string]any{},
						},
						{
							"cid": "bafyreihpq4duzngkledmxkxx3jevlp2q4aimhmbjygpv5chmgbf6u2fsqm",
							"links": []map[string]any{
								{
									"cid":  "bafyreihakk5jjukb4fw7klfejdmniwhuscnckcjo677p3mtcxrdpiahuea",
									"name": "age",
								},
								{
									"cid":  "bafyreihx4lnknvruc6vonsg3dvb3nnlsycwzbbkeulcutnzgidkzfvea64",
									"name": "name",
								},
							},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocIDAndUpdate(t *testing.T) {
	test := testUtils.TestCase{
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
						_commits(docID: "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738") {
							cid
							height
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid":    "bafyreia5jhb6ughpzd2rjszl4qbdd4w5zrdjfoseyrvnmhm2xiyrudvja4",
							"height": int64(2),
						},
						{
							"cid":    "bafyreihakk5jjukb4fw7klfejdmniwhuscnckcjo677p3mtcxrdpiahuea",
							"height": int64(1),
						},
						{
							"cid":    "bafyreihx4lnknvruc6vonsg3dvb3nnlsycwzbbkeulcutnzgidkzfvea64",
							"height": int64(1),
						},
						{
							"cid":    "bafyreieira5p74wdicqhelwbjsin7jtnnvvlplngrrcqfapleq2phexqga",
							"height": int64(2),
						},
						{
							"cid":    "bafyreihpq4duzngkledmxkxx3jevlp2q4aimhmbjygpv5chmgbf6u2fsqm",
							"height": int64(1),
						},
					},
				},
				NonOrderedResults: true,
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
						_commits(docID: "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738") {
							cid
							links {
								cid
								name
							}
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid": "bafyreia5jhb6ughpzd2rjszl4qbdd4w5zrdjfoseyrvnmhm2xiyrudvja4",
							"links": []map[string]any{
								{
									"cid":  "bafyreihakk5jjukb4fw7klfejdmniwhuscnckcjo677p3mtcxrdpiahuea",
									"name": "_head",
								},
							},
						},
						{
							"cid":   "bafyreihakk5jjukb4fw7klfejdmniwhuscnckcjo677p3mtcxrdpiahuea",
							"links": []map[string]any{},
						},
						{
							"cid":   "bafyreihx4lnknvruc6vonsg3dvb3nnlsycwzbbkeulcutnzgidkzfvea64",
							"links": []map[string]any{},
						},
						{
							"cid": "bafyreieira5p74wdicqhelwbjsin7jtnnvvlplngrrcqfapleq2phexqga",
							"links": []map[string]any{
								{
									"cid":  "bafyreihpq4duzngkledmxkxx3jevlp2q4aimhmbjygpv5chmgbf6u2fsqm",
									"name": "_head",
								},
								{
									"cid":  "bafyreia5jhb6ughpzd2rjszl4qbdd4w5zrdjfoseyrvnmhm2xiyrudvja4",
									"name": "age",
								},
							},
						},
						{
							"cid": "bafyreihpq4duzngkledmxkxx3jevlp2q4aimhmbjygpv5chmgbf6u2fsqm",
							"links": []map[string]any{
								{
									"cid":  "bafyreihakk5jjukb4fw7klfejdmniwhuscnckcjo677p3mtcxrdpiahuea",
									"name": "age",
								},
								{
									"cid":  "bafyreihx4lnknvruc6vonsg3dvb3nnlsycwzbbkeulcutnzgidkzfvea64",
									"name": "name",
								},
							},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
