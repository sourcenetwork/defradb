// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package aggregates

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaAggregateSimpleCreatesUsersCount(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {}
				`,
			},
			testUtils.IntrospectionRequest{
				Request: `
					query {
						__type (name: "Users") {
							name
							fields {
								name
								args {
									name
									type {
										name
										inputFields {
											name
											type {
												name
											}
										}
									}
								}
							}
						}
					}
				`,
				ContainsData: map[string]any{
					"__type": map[string]any{
						"name": "Users",
						"fields": []any{
							map[string]any{
								"name": "COUNT",
								"args": []any{
									map[string]any{
										"name": "GROUP",
										"type": map[string]any{
											"name": "Users__CountSelector",
											"inputFields": []any{
												map[string]any{
													"name": "filter",
													"type": map[string]any{
														"name": "UsersFilterArg",
													},
												},
												map[string]any{
													"name": "limit",
													"type": map[string]any{
														"name": "Int",
													},
												},
												map[string]any{
													"name": "offset",
													"type": map[string]any{
														"name": "Int",
													},
												},
											},
										},
									},
									map[string]any{
										"name": "_version",
										"type": map[string]any{
											"name": "Users___version__CountSelector",
											"inputFields": []any{
												map[string]any{
													"name": "limit",
													"type": map[string]any{
														"name": "Int",
													},
												},
												map[string]any{
													"name": "offset",
													"type": map[string]any{
														"name": "Int",
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

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaAggregateSimpleCreatesUsersSum(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {}
				`,
			},
			testUtils.IntrospectionRequest{
				Request: `
					query {
						__type (name: "Users") {
							name
							fields {
								name
								args {
									name
									type {
										name
										inputFields {
											name
											type {
												name
												kind
												ofType {
													name
												}
											}
										}
									}
								}
							}
						}
					}
				`,
				ContainsData: map[string]any{
					"__type": map[string]any{
						"fields": []any{
							map[string]any{
								"args": []any{
									map[string]any{
										"name": "docID",
										"type": map[string]any{
											"inputFields": any(nil),
											"name":        any(nil),
										},
									},
									map[string]any{
										"name": "filter",
										"type": map[string]any{
											"inputFields": []any{
												map[string]any{
													"name": "_alias",
													"type": map[string]any{
														"kind":   "SCALAR",
														"name":   "JSON",
														"ofType": any(nil),
													},
												},
												map[string]any{
													"name": "_and",
													"type": map[string]any{
														"kind": "LIST",
														"name": any(nil),
														"ofType": map[string]any{
															"name": any(nil),
														},
													},
												},
												map[string]any{
													"name": "_docID",
													"type": map[string]any{
														"kind":   "INPUT_OBJECT",
														"name":   "IDOperatorBlock",
														"ofType": any(nil),
													},
												},
												map[string]any{
													"name": "_not",
													"type": map[string]any{
														"kind":   "INPUT_OBJECT",
														"name":   "UsersFilterArg",
														"ofType": any(nil),
													},
												},
												map[string]any{
													"name": "_or",
													"type": map[string]any{
														"kind": "LIST",
														"name": any(nil),
														"ofType": map[string]any{
															"name": any(nil),
														},
													},
												},
											},
											"name": "UsersFilterArg",
										},
									},
									map[string]any{
										"name": "groupBy",
										"type": map[string]any{
											"inputFields": any(nil),
											"name":        any(nil),
										},
									},
									map[string]any{
										"name": "limit",
										"type": map[string]any{
											"inputFields": any(nil),
											"name":        "Int",
										},
									},
									map[string]any{
										"name": "offset",
										"type": map[string]any{
											"inputFields": any(nil),
											"name":        "Int",
										},
									},
									map[string]any{
										"name": "order",
										"type": map[string]any{
											"inputFields": any(nil),
											"name":        any(nil),
										},
									},
								},
								"name": "GROUP",
							},
							map[string]any{
								"args": []any{
									map[string]any{
										"name": "GROUP",
										"type": map[string]any{
											"inputFields": []any{
												map[string]any{
													"name": "field",
													"type": map[string]any{
														"kind": "NON_NULL",
														"name": any(nil),
														"ofType": map[string]any{
															"name": "UsersNumericFieldsArg",
														},
													},
												},
												map[string]any{
													"name": "filter",
													"type": map[string]any{
														"kind":   "INPUT_OBJECT",
														"name":   "UsersFilterArg",
														"ofType": any(nil),
													},
												},
												map[string]any{
													"name": "limit",
													"type": map[string]any{
														"kind":   "SCALAR",
														"name":   "Int",
														"ofType": any(nil),
													},
												},
												map[string]any{
													"name": "offset",
													"type": map[string]any{
														"kind":   "SCALAR",
														"name":   "Int",
														"ofType": any(nil),
													},
												},
												map[string]any{
													"name": "order",
													"type": map[string]any{
														"kind": "LIST",
														"name": any(nil),
														"ofType": map[string]any{
															"name": "UsersOrderArg",
														},
													},
												},
											},
											"name": "Users__NumericSelector",
										},
									},
								},
								"name": "SUM",
							},
						},
						"name": "Users",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaAggregateSimpleCreatesUsersAverage(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `type Users {}`,
			},
			testUtils.IntrospectionRequest{
				Request: `
					query {
						__type(name: "Users") {
							name
							fields {
								name
								args {
									name
									type {
										name
										inputFields {
											name
											type { name kind ofType { name } }
										}
									}
								}
							}
						}
					}
				`,
				ContainsData: map[string]any{
					"__type": map[string]any{
						"name": "Users",
						"fields": []any{
							map[string]any{
								"name": "AVG",
								"args": []any{
									map[string]any{
										"name": "GROUP",
										"type": map[string]any{
											"name": "Users__NumericSelector",
											"inputFields": []any{
												map[string]any{
													"name": "field",
													"type": map[string]any{
														"kind": "NON_NULL",
														"name": any(nil),
														"ofType": map[string]any{
															"name": "UsersNumericFieldsArg",
														},
													},
												},
												map[string]any{
													"name": "filter",
													"type": map[string]any{
														"kind":   "INPUT_OBJECT",
														"name":   "UsersFilterArg",
														"ofType": any(nil),
													},
												},
												map[string]any{
													"name": "limit",
													"type": map[string]any{
														"kind":   "SCALAR",
														"name":   "Int",
														"ofType": any(nil),
													},
												},
												map[string]any{
													"name": "offset",
													"type": map[string]any{
														"kind":   "SCALAR",
														"name":   "Int",
														"ofType": any(nil),
													},
												},
												map[string]any{
													"name": "order",
													"type": map[string]any{
														"kind": "LIST",
														"name": any(nil),
														"ofType": map[string]any{
															"name": "UsersOrderArg",
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

	testUtils.ExecuteTestCase(t, test)
}
