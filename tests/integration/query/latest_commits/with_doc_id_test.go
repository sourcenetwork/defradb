// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package latest_commits

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryLatestCommitsWithDocID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			testUtils.Request{
				Request: `query {
					_latestCommits(docID: "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738") {
						cid
						links {
							cid
							name
						}
					}
				}`,
				Results: map[string]any{
					"_latestCommits": []map[string]any{
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

	executeTestCase(t, test)
}

func TestQueryLatestCommitsWithDocIDWithSchemaVersionIDField(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			testUtils.Request{
				Request: `query {
					_latestCommits(docID: "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738") {
						cid
						schemaVersionId
					}
				}`,
				Results: map[string]any{
					"_latestCommits": []map[string]any{
						{
							"cid":             "bafyreihpq4duzngkledmxkxx3jevlp2q4aimhmbjygpv5chmgbf6u2fsqm",
							"schemaVersionId": "bafyreicrgjxxcviov5jawe2haq5fbtd4jxt63vsdhqpcyaaahiothj72tu",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryLatestCommits_WithDocIDAndAliased_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			testUtils.Request{
				Request: `query {
					history: _latestCommits(docID: "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738") {
						cid
						links {
							cid
							name
						}
					}
				}`,
				Results: map[string]any{
					"history": []map[string]any{
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

	executeTestCase(t, test)
}
