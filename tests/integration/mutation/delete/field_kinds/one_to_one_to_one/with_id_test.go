// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_one_to_one

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestRelationalDeletionOfADocumentUsingSingleKey_Success(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				// Books
				CollectionID: 0,
				// bae-320cb3e1-4dff-51e8-bccd-b1852b616031
				Doc: `{
						"name": "100 Go Mistakes to Avoid.",
						"rating": 4.8,
						"_publisherID": "bae-180f2922-98e3-53cf-8012-a2b28192b8bb"
					}`,
			},
			testUtils.CreateDoc{
				// Authors
				CollectionID: 1,
				// bae-2a512f5c-a48d-55b1-8a72-b5d01b9bd897
				Doc: `{
						"name": "Teiva Harsanyi",
						"age": 48,
						"verified": true,
						"_wroteID": "bae-320cb3e1-4dff-51e8-bccd-b1852b616031"
					}`,
			},
			testUtils.CreateDoc{
				// Publishers
				CollectionID: 2,
				// bae-180f2922-98e3-53cf-8012-a2b28192b8bb
				Doc: `{
						"name": "Manning Early Access Program (MEAP)",
						"address": "Online"
					}`,
			},
			testUtils.Request{
				Request: `mutation {
						delete_Author(docID: "bae-2a512f5c-a48d-55b1-8a72-b5d01b9bd897") {
							_docID
						}
					}`,
				Results: map[string]any{
					"delete_Author": []map[string]any{
						{
							"_docID": "bae-2a512f5c-a48d-55b1-8a72-b5d01b9bd897",
						},
					},
				},
			},
		},
	}

	execute(t, test)
}

func TestRelationalDeletionOfADocumentUsingSingleKeyWithAlias_Success(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				// Books
				CollectionID: 0,
				// bae-320cb3e1-4dff-51e8-bccd-b1852b616031
				Doc: `{
						"name": "100 Go Mistakes to Avoid.",
						"rating": 4.8,
						"_publisherID": "bae-180f2922-98e3-53cf-8012-a2b28192b8bb"
					}`,
			},
			testUtils.CreateDoc{
				// Authors
				CollectionID: 1,
				// bae-2a512f5c-a48d-55b1-8a72-b5d01b9bd897
				Doc: `{
						"name": "Teiva Harsanyi",
						"age": 48,
						"verified": true,
						"_wroteID": "bae-320cb3e1-4dff-51e8-bccd-b1852b616031"
					}`,
			},
			testUtils.CreateDoc{
				// Publishers
				CollectionID: 2,
				// bae-180f2922-98e3-53cf-8012-a2b28192b8bb
				Doc: `{
						"name": "Manning Early Access Program (MEAP)",
						"address": "Online"
					}`,
			},
			testUtils.Request{
				Request: `mutation {
						delete_Author(docID: "bae-2a512f5c-a48d-55b1-8a72-b5d01b9bd897") {
							AliasOfKey: _docID
						}
					}`,
				Results: map[string]any{
					"delete_Author": []map[string]any{
						{
							"AliasOfKey": "bae-2a512f5c-a48d-55b1-8a72-b5d01b9bd897",
						},
					},
				},
			},
		},
	}

	execute(t, test)
}

func TestRelationalDeletionOfADocumentUsingSingleKeyWithMultiDocumentsWithAlias_Success(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				// Books
				CollectionID: 0,
				// bae-320cb3e1-4dff-51e8-bccd-b1852b616031
				Doc: `{
						"name": "100 Go Mistakes to Avoid.",
						"rating": 4.8,
						"_publisherID": "bae-180f2922-98e3-53cf-8012-a2b28192b8bb"
					}`,
			},
			testUtils.CreateDoc{
				// Authors
				CollectionID: 1,
				// bae-2a512f5c-a48d-55b1-8a72-b5d01b9bd897
				Doc: `{
						"name": "Teiva Harsanyi",
						"age": 48,
						"verified": true,
						"_wroteID": "bae-320cb3e1-4dff-51e8-bccd-b1852b616031"
					}`,
			},
			testUtils.CreateDoc{
				// Publishers
				CollectionID: 2,
				// bae-180f2922-98e3-53cf-8012-a2b28192b8bb
				Doc: `{
						"name": "Manning Early Access Program (MEAP)",
						"address": "Online"
					}`,
			},
			testUtils.CreateDoc{
				// Publishers
				CollectionID: 2,
				// bae-df73d5f3-1d99-5269-ac5a-ea75c4b18815
				Doc: `{
						"name": "Manning Publications",
						"address": "Website"
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 1,
				DocID:        0,
				Doc: `{
						"name": "Teiva Harsanyiiiiiiiiii",
						"age": 49
					}`,
			},
			testUtils.Request{
				Request: `mutation {
						delete_Author(docID: "bae-2a512f5c-a48d-55b1-8a72-b5d01b9bd897") {
							Key: _docID
						}
					}`,
				Results: map[string]any{
					"delete_Author": []map[string]any{
						{
							"Key": "bae-2a512f5c-a48d-55b1-8a72-b5d01b9bd897",
						},
					},
				},
			},
		},
	}

	execute(t, test)
}
