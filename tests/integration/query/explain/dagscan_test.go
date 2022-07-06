// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_explain

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// ===============================[ allCommits tests ]===============================
func TestExplainAllCommitsDagScan(t *testing.T) {
	test := testUtils.QueryTestCase{

		Description: "Explain allCommits query.",

		Query: `query @explain {
			allCommits (dockey: "bae-41598f0c-19bc-5da6-813b-e80f14a10df3", field: "1") {
				links {
					cid
				}
			}
		}`,

		Docs: map[int][]string{
			//authors
			2: {
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
				// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"selectNode": dataMap{
							"filter": nil,
							"commitSelectNode": dataMap{
								"dagScanNode": dataMap{
									"spans": []dataMap{
										{
											"start": "/bae-41598f0c-19bc-5da6-813b-e80f14a10df3/1",
											"end":   "/bae-41598f0c-19bc-5da6-813b-e80f14a10df3/2",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestExplainAllCommitsDagScanWithoutField(t *testing.T) {
	test := testUtils.QueryTestCase{

		Description: "Explain allCommits query with only dockey (no field).",

		Query: `query @explain {
			allCommits (dockey: "bae-41598f0c-19bc-5da6-813b-e80f14a10df3") {
				links {
					cid
				}
			}
		}`,

		Docs: map[int][]string{
			//authors
			2: {
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
				// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"selectNode": dataMap{
							"filter": nil,
							"commitSelectNode": dataMap{
								"dagScanNode": dataMap{
									"spans": []dataMap{
										{
											"start": "/bae-41598f0c-19bc-5da6-813b-e80f14a10df3/C",
											"end":   "/bae-41598f0c-19bc-5da6-813b-e80f14a10df3/D",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestExplainAllCommitsDagScanWithoutDocKey_Failure(t *testing.T) {
	test := testUtils.QueryTestCase{

		Description: "Explain allCommits query without the dockey.",

		Query: `query @explain {
			allCommits (field: "1") {
				links {
					cid
				}
			}
		}`,

		ExpectedError: "[Field \"allCommits\" argument \"dockey\" of type \"ID!\" is required but not provided.]",
	}

	executeTestCase(t, test)
}

func TestExplainAllCommitsDagScanWithoutAnyArguments_Failure(t *testing.T) {
	test := testUtils.QueryTestCase{

		Description: "Explain allCommits query without any arguments.",

		Query: `query @explain {
			allCommits {
				links {
					cid
				}
			}
		}`,

		ExpectedError: "[Field \"allCommits\" argument \"dockey\" of type \"ID!\" is required but not provided.]",
	}

	executeTestCase(t, test)
}

// ===============================[ allCommits tests ]===============================
func TestExplainLatestCommitsDagScan(t *testing.T) {
	test := testUtils.QueryTestCase{

		Description: "Explain latestCommits query.",

		Query: `query @explain {
			latestCommits(dockey: "bae-41598f0c-19bc-5da6-813b-e80f14a10df3", field: "1") {
				cid
				links {
					cid
				}
			}
		}`,

		Docs: map[int][]string{
			//authors
			2: {
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
				// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"selectNode": dataMap{
							"filter": nil,
							"commitSelectNode": dataMap{
								"dagScanNode": dataMap{
									"spans": []dataMap{
										{
											"start": "/bae-41598f0c-19bc-5da6-813b-e80f14a10df3/1",
											"end":   "/bae-41598f0c-19bc-5da6-813b-e80f14a10df3/2",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestExplainLatestCommitsDagScanWithoutField(t *testing.T) {
	test := testUtils.QueryTestCase{

		Description: "Explain latestCommits query with only dockey (no field).",

		Query: `query @explain {
			latestCommits(dockey: "bae-41598f0c-19bc-5da6-813b-e80f14a10df3") {
				cid
				links {
					cid
				}
			}
		}`,

		Docs: map[int][]string{
			//authors
			2: {
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
				// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"selectNode": dataMap{
							"filter": nil,
							"commitSelectNode": dataMap{
								"dagScanNode": dataMap{
									"spans": []dataMap{
										{
											"start": "/bae-41598f0c-19bc-5da6-813b-e80f14a10df3/C",
											"end":   "/bae-41598f0c-19bc-5da6-813b-e80f14a10df3/D",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestExplainLatestCommitsDagScanWithoutDocKey_Failure(t *testing.T) {
	test := testUtils.QueryTestCase{

		Description: "Explain latestCommits query without DocKey.",

		Query: `query @explain {
			latestCommits(field: "1") {
				cid
				links {
					cid
				}
			}
		}`,

		ExpectedError: "Field \"latestCommits\" argument \"dockey\" of type \"ID!\" is required but not provided.",
	}

	executeTestCase(t, test)
}

func TestExplainLatestCommitsDagScanWithoutAnyArguments_Failure(t *testing.T) {
	test := testUtils.QueryTestCase{

		Description: "Explain latestCommits query without any arguments.",

		Query: `query @explain {
			latestCommits {
				cid
				links {
					cid
				}
			}
		}`,

		ExpectedError: "Field \"latestCommits\" argument \"dockey\" of type \"ID!\" is required but not provided.",
	}

	executeTestCase(t, test)
}

// ===============================[ commit test ]===============================
func TestExplainOneCommitDagScan(t *testing.T) {
	test := testUtils.QueryTestCase{

		Description: "Explain commit query.",

		Query: `query @explain {
		  commit(cid: "bafybeieys3v7timp2c2k4krqeojaa7ecqaairvtf7lo7jah24gtuicpk3u") {
				cid
				height
				delta
			}
		}`,

		Docs: map[int][]string{
			//authors
			2: {
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
				// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"selectNode": dataMap{
							"filter": nil,
							"commitSelectNode": dataMap{
								"dagScanNode": dataMap{
									"spans": []dataMap{},
								},
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestExplainOneCommitDagScanWithNoArguments(t *testing.T) {
	test := testUtils.QueryTestCase{

		Description: "Explain commit query with no arguments.",

		Query: `query @explain {
		  commit {
				cid
				height
				delta
			}
		}`,

		ExpectedError: "Field \"commit\" argument \"cid\" of type \"ID!\" is required but not provided.",
	}

	executeTestCase(t, test)
}
