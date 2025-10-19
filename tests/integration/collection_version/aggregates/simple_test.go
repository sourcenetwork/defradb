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
								"name": "_count",
								"args": []any{
									map[string]any{
										"name": "_group",
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
						"name": "Users",
						"fields": []any{
							map[string]any{
								"name": "_sum",
								"args": []any{
									map[string]any{
										"name": "FavouriteFloats",
										"type": map[string]any{
											"name": "Users__FavouriteFloats__NumericSelector",
											"inputFields": []any{
												map[string]any{
													"name": "filter",
													"type": map[string]any{
														"name": "NotNullFloat64FilterArg",
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
												map[string]any{
													"name": "order",
													"type": map[string]any{
														"name": "Ordering",
													},
												},
											},
										},
									},
									map[string]any{
										"name": "_count",
										"type": map[string]any{
											"inputFields": []any{},
											"name":        "",
										},
									},
									map[string]any{
										"name": "_deleted",
										"type": map[string]any{
											"inputFields": []any{},
											"name":        "",
										},
									},
									map[string]any{
										"name": "_docID",
										"type": map[string]any{
											"inputFields": []any{},
											"name":        "",
										},
									},
									map[string]any{
										"name": "_group",
										"type": map[string]any{
											"name": "Users__NumericSelector",
											"inputFields": []any{
												map[string]any{
													"name": "field",
													"type": map[string]any{
														"name": nil,
													},
												},
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
												map[string]any{
													"name": "order",
													"type": map[string]any{
														"name": nil,
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

				//
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaAggregateSimpleCreatesUsersAverage(t *testing.T) {
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
				ContainsData: map[string]interface{}{
					"__type": map[string]interface{}{
						"fields": []interface{}{
							map[string]interface{}{
								"args": []interface{}{
									map[string]interface{}{
										"name": "_count",
										"type": map[string]interface{}{
											"inputFields": []interface{}{},
											"name":        "",
										},
									},
									map[string]interface{}{
										"name": "_deleted",
										"type": map[string]interface{}{
											"inputFields": []interface{}{},
											"name":        "",
										},
									},
									map[string]interface{}{
										"name": "_docID",
										"type": map[string]interface{}{
											"inputFields": []interface{}{},
											"name":        "",
										},
									},
									map[string]interface{}{
										"name": "_group",
										"type": map[string]interface{}{
											"inputFields": []interface{}{
												map[string]interface{}{
													"name": "field",
													"type": map[string]interface{}{
														"kind": "NON_NULL",
														"name": interface{}(nil),
														"ofType": map[string]interface{}{
															"name": "UsersNumericFieldsArg",
														},
													},
												},
												map[string]interface{}{
													"name": "filter",
													"type": map[string]interface{}{
														"kind":   "INPUT_OBJECT",
														"name":   "UsersFilterArg",
														"ofType": interface{}(nil),
													},
												},
												map[string]interface{}{
													"name": "limit",
													"type": map[string]interface{}{
														"kind":   "SCALAR",
														"name":   "Int",
														"ofType": interface{}(nil),
													},
												},
												map[string]interface{}{
													"name": "offset",
													"type": map[string]interface{}{
														"kind":   "SCALAR",
														"name":   "Int",
														"ofType": interface{}(nil),
													},
												},
												map[string]interface{}{
													"name": "order",
													"type": map[string]interface{}{
														"kind": "LIST",
														"name": interface{}(nil),
														"ofType": map[string]interface{}{
															"name": "UsersOrderArg",
														},
													},
												},
											},
											"name": "Users__NumericSelector",
										},
									},
									map[string]interface{}{
										"name": "_sum",
										"type": map[string]interface{}{
											"inputFields": []interface{}{},
											"name":        "",
										},
									},
								},
								"name": "_avg",
							},
							map[string]interface{}{
								"args": []interface{}{
									map[string]interface{}{
										"name": "_group",
										"type": map[string]interface{}{
											"inputFields": []interface{}{
												map[string]interface{}{
													"name": "filter",
													"type": map[string]interface{}{
														"kind":   "INPUT_OBJECT",
														"name":   "UsersFilterArg",
														"ofType": interface{}(nil),
													},
												},
												map[string]interface{}{
													"name": "limit",
													"type": map[string]interface{}{
														"kind":   "SCALAR",
														"name":   "Int",
														"ofType": interface{}(nil),
													},
												},
												map[string]interface{}{
													"name": "offset",
													"type": map[string]interface{}{
														"kind":   "SCALAR",
														"name":   "Int",
														"ofType": interface{}(nil),
													},
												},
											},
											"name": "Users__CountSelector",
										},
									},
									map[string]interface{}{
										"name": "_version",
										"type": map[string]interface{}{
											"inputFields": []interface{}{
												map[string]interface{}{
													"name": "limit",
													"type": map[string]interface{}{
														"kind":   "SCALAR",
														"name":   "Int",
														"ofType": interface{}(nil),
													},
												},
												map[string]interface{}{
													"name": "offset",
													"type": map[string]interface{}{
														"kind":   "SCALAR",
														"name":   "Int",
														"ofType": interface{}(nil),
													},
												},
											},
											"name": "Users___version__CountSelector",
										},
									},
								},
								"name": "_count",
							},
							map[string]interface{}{
								"args": []interface{}{},
								"name": "_deleted",
							},
							map[string]interface{}{
								"args": []interface{}{},
								"name": "_docID",
							},
							map[string]interface{}{
								"args": []interface{}{
									map[string]interface{}{
										"name": "docID",
										"type": map[string]interface{}{
											"inputFields": interface{}(nil),
											"name":        interface{}(nil),
										},
									},
									map[string]interface{}{
										"name": "filter",
										"type": map[string]interface{}{
											"inputFields": []interface{}{
												map[string]interface{}{
													"name": "_alias",
													"type": map[string]interface{}{
														"kind":   "SCALAR",
														"name":   "JSON",
														"ofType": interface{}(nil),
													},
												},
												map[string]interface{}{
													"name": "_and",
													"type": map[string]interface{}{
														"kind": "LIST",
														"name": interface{}(nil),
														"ofType": map[string]interface{}{
															"name": interface{}(nil),
														},
													},
												},
												map[string]interface{}{
													"name": "_docID",
													"type": map[string]interface{}{
														"kind":   "INPUT_OBJECT",
														"name":   "IDOperatorBlock",
														"ofType": interface{}(nil),
													},
												},
												map[string]interface{}{
													"name": "_not",
													"type": map[string]interface{}{
														"kind":   "INPUT_OBJECT",
														"name":   "UsersFilterArg",
														"ofType": interface{}(nil),
													},
												},
												map[string]interface{}{
													"name": "_or",
													"type": map[string]interface{}{
														"kind": "LIST",
														"name": interface{}(nil),
														"ofType": map[string]interface{}{
															"name": interface{}(nil),
														},
													},
												},
											},
											"name": "UsersFilterArg",
										},
									},
									map[string]interface{}{
										"name": "groupBy",
										"type": map[string]interface{}{
											"inputFields": interface{}(nil),
											"name":        interface{}(nil),
										},
									},
									map[string]interface{}{
										"name": "limit",
										"type": map[string]interface{}{
											"inputFields": interface{}(nil),
											"name":        "Int",
										},
									},
									map[string]interface{}{
										"name": "offset",
										"type": map[string]interface{}{
											"inputFields": interface{}(nil),
											"name":        "Int",
										},
									},
									map[string]interface{}{
										"name": "order",
										"type": map[string]interface{}{
											"inputFields": interface{}(nil),
											"name":        interface{}(nil),
										},
									},
								},
								"name": "_group",
							},
							map[string]interface{}{
								"args": []interface{}{
									map[string]interface{}{
										"name": "_avg",
										"type": map[string]interface{}{
											"inputFields": []interface{}{},
											"name":        "",
										},
									},
									map[string]interface{}{
										"name": "_count",
										"type": map[string]interface{}{
											"inputFields": []interface{}{},
											"name":        "",
										},
									},
									map[string]interface{}{
										"name": "_deleted",
										"type": map[string]interface{}{
											"inputFields": []interface{}{},
											"name":        "",
										},
									},
									map[string]interface{}{
										"name": "_docID",
										"type": map[string]interface{}{
											"inputFields": []interface{}{},
											"name":        "",
										},
									},
									map[string]interface{}{
										"name": "_group",
										"type": map[string]interface{}{
											"inputFields": []interface{}{
												map[string]interface{}{
													"name": "field",
													"type": map[string]interface{}{
														"kind": "NON_NULL",
														"name": interface{}(nil),
														"ofType": map[string]interface{}{
															"name": "UsersNumericFieldsArg",
														},
													},
												},
												map[string]interface{}{
													"name": "filter",
													"type": map[string]interface{}{
														"kind":   "INPUT_OBJECT",
														"name":   "UsersFilterArg",
														"ofType": interface{}(nil),
													},
												},
												map[string]interface{}{
													"name": "limit",
													"type": map[string]interface{}{
														"kind":   "SCALAR",
														"name":   "Int",
														"ofType": interface{}(nil),
													},
												},
												map[string]interface{}{
													"name": "offset",
													"type": map[string]interface{}{
														"kind":   "SCALAR",
														"name":   "Int",
														"ofType": interface{}(nil),
													},
												},
												map[string]interface{}{
													"name": "order",
													"type": map[string]interface{}{
														"kind": "LIST",
														"name": interface{}(nil),
														"ofType": map[string]interface{}{
															"name": "UsersOrderArg",
														},
													},
												},
											},
											"name": "Users__NumericSelector",
										},
									},
									map[string]interface{}{
										"name": "_sum",
										"type": map[string]interface{}{
											"inputFields": []interface{}{},
											"name":        "",
										},
									},
								},
								"name": "_max",
							},
							map[string]interface{}{
								"args": []interface{}{
									map[string]interface{}{
										"name": "_avg",
										"type": map[string]interface{}{
											"inputFields": []interface{}{},
											"name":        "",
										},
									},
									map[string]interface{}{
										"name": "_count",
										"type": map[string]interface{}{
											"inputFields": []interface{}{},
											"name":        "",
										},
									},
									map[string]interface{}{
										"name": "_deleted",
										"type": map[string]interface{}{
											"inputFields": []interface{}{},
											"name":        "",
										},
									},
									map[string]interface{}{
										"name": "_docID",
										"type": map[string]interface{}{
											"inputFields": []interface{}{},
											"name":        "",
										},
									},
									map[string]interface{}{
										"name": "_group",
										"type": map[string]interface{}{
											"inputFields": []interface{}{
												map[string]interface{}{
													"name": "field",
													"type": map[string]interface{}{
														"kind": "NON_NULL",
														"name": interface{}(nil),
														"ofType": map[string]interface{}{
															"name": "UsersNumericFieldsArg",
														},
													},
												},
												map[string]interface{}{
													"name": "filter",
													"type": map[string]interface{}{
														"kind":   "INPUT_OBJECT",
														"name":   "UsersFilterArg",
														"ofType": interface{}(nil),
													},
												},
												map[string]interface{}{
													"name": "limit",
													"type": map[string]interface{}{
														"kind":   "SCALAR",
														"name":   "Int",
														"ofType": interface{}(nil),
													},
												},
												map[string]interface{}{
													"name": "offset",
													"type": map[string]interface{}{
														"kind":   "SCALAR",
														"name":   "Int",
														"ofType": interface{}(nil),
													},
												},
												map[string]interface{}{
													"name": "order",
													"type": map[string]interface{}{
														"kind": "LIST",
														"name": interface{}(nil),
														"ofType": map[string]interface{}{
															"name": "UsersOrderArg",
														},
													},
												},
											},
											"name": "Users__NumericSelector",
										},
									},
									map[string]interface{}{
										"name": "_max",
										"type": map[string]interface{}{
											"inputFields": []interface{}{},
											"name":        "",
										},
									},
									map[string]interface{}{
										"name": "_sum",
										"type": map[string]interface{}{
											"inputFields": []interface{}{},
											"name":        "",
										},
									},
								},
								"name": "_min",
							},
							map[string]interface{}{
								"args": []interface{}{},
								"name": "_similarity",
							},
							map[string]interface{}{
								"args": []interface{}{
									map[string]interface{}{
										"name": "_count",
										"type": map[string]interface{}{
											"inputFields": []interface{}{},
											"name":        "",
										},
									},
									map[string]interface{}{
										"name": "_deleted",
										"type": map[string]interface{}{
											"inputFields": []interface{}{},
											"name":        "",
										},
									},
									map[string]interface{}{
										"name": "_docID",
										"type": map[string]interface{}{
											"inputFields": []interface{}{},
											"name":        "",
										},
									},
									map[string]interface{}{
										"name": "_group",
										"type": map[string]interface{}{
											"inputFields": []interface{}{
												map[string]interface{}{
													"name": "field",
													"type": map[string]interface{}{
														"kind": "NON_NULL",
														"name": interface{}(nil),
														"ofType": map[string]interface{}{
															"name": "UsersNumericFieldsArg",
														},
													},
												},
												map[string]interface{}{
													"name": "filter",
													"type": map[string]interface{}{
														"kind":   "INPUT_OBJECT",
														"name":   "UsersFilterArg",
														"ofType": interface{}(nil),
													},
												},
												map[string]interface{}{
													"name": "limit",
													"type": map[string]interface{}{
														"kind":   "SCALAR",
														"name":   "Int",
														"ofType": interface{}(nil),
													},
												},
												map[string]interface{}{
													"name": "offset",
													"type": map[string]interface{}{
														"kind":   "SCALAR",
														"name":   "Int",
														"ofType": interface{}(nil),
													},
												},
												map[string]interface{}{
													"name": "order",
													"type": map[string]interface{}{
														"kind": "LIST",
														"name": interface{}(nil),
														"ofType": map[string]interface{}{
															"name": "UsersOrderArg",
														},
													},
												},
											},
											"name": "Users__NumericSelector",
										},
									},
								},
								"name": "_sum",
							},
							map[string]interface{}{
								"args": []interface{}{},
								"name": "_version",
							},
						},
						"name": "Users",
					},
				},

				// End of ContainsData
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
