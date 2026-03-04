// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package test_explain_execute

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
)

func TestExecuteExplainRequestWithAOneToOneJoin(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			add2AddressDocuments(),
			add2AuthorContactDocuments(),
			add2AuthorDocuments(),

			&action.ExplainRequest{
				Request: `query @explain(type: execute) {
					Author {
						OnlyEmail: contact {
							email
						}
					}
				}`,

				ExpectedFullGraph: dataMap{
					"explain": dataMap{
						"executionSuccess": true,
						"sizeOfResult":     1,
						"planExecutions":   uint64(2),
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"selectNode": dataMap{
										"iterations":    uint64(3),
										"filterMatches": uint64(2),
										"typeIndexJoin": dataMap{
											"iterations": uint64(3),
											"typeJoinOne": dataMap{
												"root": dataMap{
													"scanNode": dataMap{
														"iterations":   uint64(3),
														"docFetches":   uint64(2),
														"fieldFetches": uint64(8),
														"indexFetches": uint64(0),
													},
												},
												"subType": dataMap{
													"selectTopNode": dataMap{
														"selectNode": dataMap{
															"iterations":    uint64(2),
															"filterMatches": uint64(2),
															"scanNode": dataMap{
																"iterations":   uint64(2),
																"docFetches":   uint64(2),
																"fieldFetches": uint64(6),
																"indexFetches": uint64(0),
															},
														},
													},
												},
											},
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

	explainUtils.ExecuteTestCase(t, test)
}

func TestExecuteExplainWithMultipleOneToOneJoins(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			add2AddressDocuments(),
			add2AuthorContactDocuments(),
			add2AuthorDocuments(),

			&action.ExplainRequest{
				Request: `query @explain(type: execute) {
					Author {
						OnlyEmail: contact {
							email
						}
						contact {
							cell
							email
						}
					}
				}`,

				ExpectedFullGraph: dataMap{
					"explain": dataMap{
						"executionSuccess": true,
						"sizeOfResult":     1,
						"planExecutions":   uint64(2),
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"selectNode": dataMap{
										"iterations":    uint64(3),
										"filterMatches": uint64(2),
										"parallelNode": []dataMap{
											{
												"typeIndexJoin": dataMap{
													"iterations": uint64(3),
													"typeJoinOne": dataMap{
														"root": dataMap{
															"scanNode": dataMap{
																"iterations":   uint64(3),
																"docFetches":   uint64(2),
																"fieldFetches": uint64(8),
																"indexFetches": uint64(0),
															},
														},
														"subType": dataMap{
															"selectTopNode": dataMap{
																"selectNode": dataMap{
																	"iterations":    uint64(2),
																	"filterMatches": uint64(2),
																	"scanNode": dataMap{
																		"iterations":   uint64(2),
																		"docFetches":   uint64(2),
																		"fieldFetches": uint64(6),
																		"indexFetches": uint64(0),
																	},
																},
															},
														},
													},
												},
											},
											{
												"typeIndexJoin": dataMap{
													"iterations": uint64(3),
													"typeJoinOne": dataMap{
														"root": dataMap{
															"scanNode": dataMap{
																"iterations":   uint64(3),
																"docFetches":   uint64(2),
																"fieldFetches": uint64(8),
																"indexFetches": uint64(0),
															},
														},
														"subType": dataMap{
															"selectTopNode": dataMap{
																"selectNode": dataMap{
																	"iterations":    uint64(2),
																	"filterMatches": uint64(2),
																	"scanNode": dataMap{
																		"iterations":   uint64(2),
																		"docFetches":   uint64(2),
																		"fieldFetches": uint64(6),
																		"indexFetches": uint64(0),
																	},
																},
															},
														},
													},
												},
											},
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

	explainUtils.ExecuteTestCase(t, test)
}

func TestExecuteExplainWithTwoLevelDeepNestedJoins(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			add2AddressDocuments(),
			add2AuthorContactDocuments(),
			add2AuthorDocuments(),

			&action.ExplainRequest{
				Request: `query @explain(type: execute) {
					Author {
						name
						contact {
							email
							address {
								city
							}
						}
					}
				}`,

				ExpectedFullGraph: dataMap{
					"explain": dataMap{
						"executionSuccess": true,
						"sizeOfResult":     1,
						"planExecutions":   uint64(2),
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"selectNode": dataMap{
										"iterations":    uint64(3),
										"filterMatches": uint64(2),
										"typeIndexJoin": dataMap{
											"iterations": uint64(3),
											"typeJoinOne": dataMap{
												"root": dataMap{
													"scanNode": dataMap{
														"iterations":   uint64(3),
														"docFetches":   uint64(2),
														"fieldFetches": uint64(8),
														"indexFetches": uint64(0),
													},
												},
												"subType": dataMap{
													"selectTopNode": dataMap{
														"selectNode": dataMap{
															"iterations":    uint64(2),
															"filterMatches": uint64(2),
															"typeIndexJoin": dataMap{
																"iterations": uint64(2),
																"typeJoinOne": dataMap{
																	"root": dataMap{
																		"scanNode": dataMap{
																			"iterations":   uint64(2),
																			"docFetches":   uint64(2),
																			"fieldFetches": uint64(6),
																			"indexFetches": uint64(0),
																		},
																	},
																	"subType": dataMap{
																		"selectTopNode": dataMap{
																			"selectNode": dataMap{
																				"iterations":    uint64(2),
																				"filterMatches": uint64(2),
																				"scanNode": dataMap{
																					"iterations":   uint64(2),
																					"docFetches":   uint64(2),
																					"fieldFetches": uint64(4),
																					"indexFetches": uint64(0),
																				},
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
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

	explainUtils.ExecuteTestCase(t, test)
}

func TestExecuteExplain_WithOneToOneJoinFromSecondarySide_ShouldIncludeIndex(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			add2AddressDocuments(),
			add2AuthorContactDocuments(),
			add2AuthorDocuments(),

			&action.ExplainRequest{
				Request: `query @explain(type: execute) {
					ContactAddress {
						city
						contact {
							email
						}
					}
				}`,

				ExpectedFullGraph: dataMap{
					"explain": dataMap{
						"executionSuccess": true,
						"sizeOfResult":     1,
						"planExecutions":   uint64(2),
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"selectNode": dataMap{
										"iterations":    uint64(3),
										"filterMatches": uint64(2),
										"typeIndexJoin": dataMap{
											"iterations": uint64(3),
											"typeJoinOne": dataMap{
												"root": dataMap{
													"scanNode": dataMap{
														"iterations":   uint64(3),
														"docFetches":   uint64(2),
														"fieldFetches": uint64(4),
														"indexFetches": uint64(0),
													},
												},
												"subType": dataMap{
													"selectTopNode": dataMap{
														"selectNode": dataMap{
															"iterations":    uint64(4),
															"filterMatches": uint64(2),
															"scanNode": dataMap{
																"iterations":   uint64(4),
																"docFetches":   uint64(2),
																"fieldFetches": uint64(6),
																"indexFetches": uint64(2),
															},
														},
													},
												},
											},
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

	explainUtils.ExecuteTestCase(t, test)
}
