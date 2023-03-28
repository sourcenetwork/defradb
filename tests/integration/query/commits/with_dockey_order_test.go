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
						"Name":	"John",
						"Age":	21
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"Age":	22
				}`,
			},
			testUtils.Request{
				Request: `query {
						commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f", order: {height: DESC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeiepww5b67jrrliuiy27erfjuivwnjca5ptdpbxrrjrqkh3b2hckyy",
						"height": int64(2),
					},
					{
						"cid":    "bafybeiebail45ch3n5rh7myumqn2jfeefnynba2ldwiesge3ddq5hu6olq",
						"height": int64(2),
					},
					{
						"cid":    "bafybeihxvx3f7eejvco6zbxsidoeuph6ywpbo33lrqm3picna2aj7pdeiu",
						"height": int64(1),
					},
					{
						"cid":    "bafybeih25dvtgei2bryhlz24tbyfdcni5di7akgcx24pezxts27wz7v454",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiapquwo7dfow7b7ovwrn3nl4e2cv2g5eoufuzylq54b4o6tatfrny",
						"height": int64(1),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestQueryCommitsWithDockeyAndOrderHeightAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, order height asc",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"Name":	"John",
						"Age":	21
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"Age":	22
				}`,
			},
			testUtils.Request{
				Request: `query {
						commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f", order: {height: ASC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeihxvx3f7eejvco6zbxsidoeuph6ywpbo33lrqm3picna2aj7pdeiu",
						"height": int64(1),
					},
					{
						"cid":    "bafybeih25dvtgei2bryhlz24tbyfdcni5di7akgcx24pezxts27wz7v454",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiapquwo7dfow7b7ovwrn3nl4e2cv2g5eoufuzylq54b4o6tatfrny",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiepww5b67jrrliuiy27erfjuivwnjca5ptdpbxrrjrqkh3b2hckyy",
						"height": int64(2),
					},
					{
						"cid":    "bafybeiebail45ch3n5rh7myumqn2jfeefnynba2ldwiesge3ddq5hu6olq",
						"height": int64(2),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestQueryCommitsWithDockeyAndOrderCidDesc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, order cid desc",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"Name":	"John",
						"Age":	21
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"Age":	22
				}`,
			},
			testUtils.Request{
				Request: `query {
						commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f", order: {cid: DESC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeihxvx3f7eejvco6zbxsidoeuph6ywpbo33lrqm3picna2aj7pdeiu",
						"height": int64(1),
					},
					{
						"cid":    "bafybeih25dvtgei2bryhlz24tbyfdcni5di7akgcx24pezxts27wz7v454",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiepww5b67jrrliuiy27erfjuivwnjca5ptdpbxrrjrqkh3b2hckyy",
						"height": int64(2),
					},
					{
						"cid":    "bafybeiebail45ch3n5rh7myumqn2jfeefnynba2ldwiesge3ddq5hu6olq",
						"height": int64(2),
					},
					{
						"cid":    "bafybeiapquwo7dfow7b7ovwrn3nl4e2cv2g5eoufuzylq54b4o6tatfrny",
						"height": int64(1),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestQueryCommitsWithDockeyAndOrderCidAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, order cid asc",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"Name":	"John",
						"Age":	21
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"Age":	22
				}`,
			},
			testUtils.Request{
				Request: `query {
						commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f", order: {cid: ASC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeiapquwo7dfow7b7ovwrn3nl4e2cv2g5eoufuzylq54b4o6tatfrny",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiebail45ch3n5rh7myumqn2jfeefnynba2ldwiesge3ddq5hu6olq",
						"height": int64(2),
					},
					{
						"cid":    "bafybeiepww5b67jrrliuiy27erfjuivwnjca5ptdpbxrrjrqkh3b2hckyy",
						"height": int64(2),
					},
					{
						"cid":    "bafybeih25dvtgei2bryhlz24tbyfdcni5di7akgcx24pezxts27wz7v454",
						"height": int64(1),
					},
					{
						"cid":    "bafybeihxvx3f7eejvco6zbxsidoeuph6ywpbo33lrqm3picna2aj7pdeiu",
						"height": int64(1),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestQueryCommitsWithDockeyAndOrderAndMultiUpdatesCidAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, multiple updates with order cid asc",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"Name":	"John",
						"Age":	21
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"Age":	22
				}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"Age":	23
				}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"Age":	24
				}`,
			},
			testUtils.Request{
				Request: `query {
						 commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f", order: {height: ASC}) {
							 cid
							 height
						 }
					 }`,
				Results: []map[string]any{
					{
						"cid":    "bafybeihxvx3f7eejvco6zbxsidoeuph6ywpbo33lrqm3picna2aj7pdeiu",
						"height": int64(1),
					},
					{
						"cid":    "bafybeih25dvtgei2bryhlz24tbyfdcni5di7akgcx24pezxts27wz7v454",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiapquwo7dfow7b7ovwrn3nl4e2cv2g5eoufuzylq54b4o6tatfrny",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiepww5b67jrrliuiy27erfjuivwnjca5ptdpbxrrjrqkh3b2hckyy",
						"height": int64(2),
					},
					{
						"cid":    "bafybeiebail45ch3n5rh7myumqn2jfeefnynba2ldwiesge3ddq5hu6olq",
						"height": int64(2),
					},
					{
						"cid":    "bafybeidag5ogxiqsctdxktobw6gbq52pbfkmicp6np5ccuy6pjyvgobdxq",
						"height": int64(3),
					},
					{
						"cid":    "bafybeid2arccjpsmmocarl6s3fe7ygstbdlpujlksuflukmun4bvf5g4oe",
						"height": int64(3),
					},
					{
						"cid":    "bafybeiadnlnuu3gkfaf5fgib3iqfsntitb4pjrc5hffvmhpdjqgsxutc5q",
						"height": int64(4),
					},
					{
						"cid":    "bafybeifyiqii323sy5xvqkltdaj5wb4p2zsotepcmv6zqxfr73lky24qw4",
						"height": int64(4),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}
