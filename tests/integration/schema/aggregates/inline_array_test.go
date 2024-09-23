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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

/* WIP
func TestSchemaAggregateInlineArrayCreatesUsersCount(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						favouriteIntegers: [Int!]
					}
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
										"name": "favouriteIntegers",
										"type": map[string]any{
											"name": "Users__favouriteIntegers__CountSelector",
											"inputFields": []any{
												map[string]any{
													"name": "filter",
													"type": map[string]any{
														"name": "NotNullIntFilterArg",
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
*/

func TestSchemaAggregateInlineArrayCreatesUsersSum(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						FavouriteFloats: [Float!]
					}
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
														"name": "NotNullFloatFilterArg",
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
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

/* WIP
func TestSchemaAggregateInlineArrayCreatesUsersAverage(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						favouriteIntegers: [Int!]
					}
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
								"name": "_avg",
								"args": []any{
									map[string]any{
										"name": "favouriteIntegers",
										"type": map[string]any{
											"name": "Users__favouriteIntegers__NumericSelector",
											"inputFields": []any{
												map[string]any{
													"name": "filter",
													"type": map[string]any{
														"name": "NotNullIntFilterArg",
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
	}

	testUtils.ExecuteTestCase(t, test)
}
*/

func aggregateGroupArg(fieldType string) map[string]any {
	return map[string]any{
		"name": "_group",
		"type": map[string]any{
			"name": "Users__CountSelector",
			"inputFields": []any{
				map[string]any{
					"name": "filter",
					"type": map[string]any{
						"name": "UsersFilterArg",
						"inputFields": []any{
							map[string]any{
								"name": "Favourites",
								"type": map[string]any{
									"name": fieldType + "ListOperatorBlock",
								},
							},
							map[string]any{
								"name": "_and",
								"type": map[string]any{
									"name": nil,
								},
							},
							map[string]any{
								"name": "_docID",
								"type": map[string]any{
									"name": "IDOperatorBlock",
								},
							},
							map[string]any{
								"name": "_not",
								"type": map[string]any{
									"name": "UsersFilterArg",
								},
							},
							map[string]any{
								"name": "_or",
								"type": map[string]any{
									"name": nil,
								},
							},
						},
					},
				},
				map[string]any{
					"name": "limit",
					"type": map[string]any{
						"name":        "Int",
						"inputFields": nil,
					},
				},
				map[string]any{
					"name": "offset",
					"type": map[string]any{
						"name":        "Int",
						"inputFields": nil,
					},
				},
			},
		},
	}
}

var aggregateVersionArg = map[string]any{
	"name": "_version",
	"type": map[string]any{
		"name": "Users___version__CountSelector",
		"inputFields": []any{
			map[string]any{
				"name": "limit",
				"type": map[string]any{
					"name":        "Int",
					"inputFields": nil,
				},
			},
			map[string]any{
				"name": "offset",
				"type": map[string]any{
					"name":        "Int",
					"inputFields": nil,
				},
			},
		},
	},
}

func TestSchemaAggregateInlineArrayCreatesUsersNillableBooleanCountFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Favourites: [Boolean]
					}
				`,
			},
			testUtils.IntrospectionRequest{
				Request: `
					query IntrospectionQuery {
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
										"name": "Favourites",
										"type": map[string]any{
											"name": "Users__Favourites__CountSelector",
											"inputFields": []any{
												map[string]any{
													"name": "filter",
													"type": map[string]any{
														"name": "BooleanFilterArg",
														"inputFields": []any{
															map[string]any{
																"name": "_and",
																"type": map[string]any{
																	"name": nil,
																},
															},
															map[string]any{
																"name": "_eq",
																"type": map[string]any{
																	"name": "Boolean",
																},
															},
															map[string]any{
																"name": "_in",
																"type": map[string]any{
																	"name": nil,
																},
															},
															map[string]any{
																"name": "_ne",
																"type": map[string]any{
																	"name": "Boolean",
																},
															},
															map[string]any{
																"name": "_nin",
																"type": map[string]any{
																	"name": nil,
																},
															},
															map[string]any{
																"name": "_or",
																"type": map[string]any{
																	"name": nil,
																},
															},
														},
													},
												},
												map[string]any{
													"name": "limit",
													"type": map[string]any{
														"name":        "Int",
														"inputFields": nil,
													},
												},
												map[string]any{
													"name": "offset",
													"type": map[string]any{
														"name":        "Int",
														"inputFields": nil,
													},
												},
											},
										},
									},
									aggregateGroupArg("Boolean"),
									aggregateVersionArg,
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

func TestSchemaAggregateInlineArrayCreatesUsersBooleanCountFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Favourites: [Boolean!]
					}
				`,
			},
			testUtils.IntrospectionRequest{
				Request: `
					query IntrospectionQuery {
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
										"name": "Favourites",
										"type": map[string]any{
											"name": "Users__Favourites__CountSelector",
											"inputFields": []any{
												map[string]any{
													"name": "filter",
													"type": map[string]any{
														"name": "NotNullBooleanFilterArg",
														"inputFields": []any{
															map[string]any{
																"name": "_and",
																"type": map[string]any{
																	"name": nil,
																},
															},
															map[string]any{
																"name": "_eq",
																"type": map[string]any{
																	"name": "Boolean",
																},
															},
															map[string]any{
																"name": "_in",
																"type": map[string]any{
																	"name": nil,
																},
															},
															map[string]any{
																"name": "_ne",
																"type": map[string]any{
																	"name": "Boolean",
																},
															},
															map[string]any{
																"name": "_nin",
																"type": map[string]any{
																	"name": nil,
																},
															},
															map[string]any{
																"name": "_or",
																"type": map[string]any{
																	"name": nil,
																},
															},
														},
													},
												},
												map[string]any{
													"name": "limit",
													"type": map[string]any{
														"name":        "Int",
														"inputFields": nil,
													},
												},
												map[string]any{
													"name": "offset",
													"type": map[string]any{
														"name":        "Int",
														"inputFields": nil,
													},
												},
											},
										},
									},
									aggregateGroupArg("NotNullBoolean"),
									aggregateVersionArg,
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

func TestSchemaAggregateInlineArrayCreatesUsersNillableIntegerCountFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Favourites: [Int]
					}
				`,
			},
			testUtils.IntrospectionRequest{
				Request: `
					query IntrospectionQuery {
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
										"name": "Favourites",
										"type": map[string]any{
											"name": "Users__Favourites__CountSelector",
											"inputFields": []any{
												map[string]any{
													"name": "filter",
													"type": map[string]any{
														"name": "IntFilterArg",
														"inputFields": []any{
															map[string]any{
																"name": "_and",
																"type": map[string]any{
																	"name": nil,
																},
															},
															map[string]any{
																"name": "_eq",
																"type": map[string]any{
																	"name": "Int",
																},
															},
															map[string]any{
																"name": "_ge",
																"type": map[string]any{
																	"name": "Int",
																},
															},
															map[string]any{
																"name": "_gt",
																"type": map[string]any{
																	"name": "Int",
																},
															},
															map[string]any{
																"name": "_in",
																"type": map[string]any{
																	"name": nil,
																},
															},
															map[string]any{
																"name": "_le",
																"type": map[string]any{
																	"name": "Int",
																},
															},
															map[string]any{
																"name": "_lt",
																"type": map[string]any{
																	"name": "Int",
																},
															},
															map[string]any{
																"name": "_ne",
																"type": map[string]any{
																	"name": "Int",
																},
															},
															map[string]any{
																"name": "_nin",
																"type": map[string]any{
																	"name": nil,
																},
															},
															map[string]any{
																"name": "_or",
																"type": map[string]any{
																	"name": nil,
																},
															},
														},
													},
												},
												map[string]any{
													"name": "limit",
													"type": map[string]any{
														"name":        "Int",
														"inputFields": nil,
													},
												},
												map[string]any{
													"name": "offset",
													"type": map[string]any{
														"name":        "Int",
														"inputFields": nil,
													},
												},
											},
										},
									},
									aggregateGroupArg("Int"),
									aggregateVersionArg,
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

func TestSchemaAggregateInlineArrayCreatesUsersIntegerCountFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Favourites: [Int!]
					}
				`,
			},
			testUtils.IntrospectionRequest{
				Request: `
					query IntrospectionQuery {
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
										"name": "Favourites",
										"type": map[string]any{
											"name": "Users__Favourites__CountSelector",
											"inputFields": []any{
												map[string]any{
													"name": "filter",
													"type": map[string]any{
														"name": "NotNullIntFilterArg",
														"inputFields": []any{
															map[string]any{
																"name": "_and",
																"type": map[string]any{
																	"name": nil,
																},
															},
															map[string]any{
																"name": "_eq",
																"type": map[string]any{
																	"name": "Int",
																},
															},
															map[string]any{
																"name": "_ge",
																"type": map[string]any{
																	"name": "Int",
																},
															},
															map[string]any{
																"name": "_gt",
																"type": map[string]any{
																	"name": "Int",
																},
															},
															map[string]any{
																"name": "_in",
																"type": map[string]any{
																	"name": nil,
																},
															},
															map[string]any{
																"name": "_le",
																"type": map[string]any{
																	"name": "Int",
																},
															},
															map[string]any{
																"name": "_lt",
																"type": map[string]any{
																	"name": "Int",
																},
															},
															map[string]any{
																"name": "_ne",
																"type": map[string]any{
																	"name": "Int",
																},
															},
															map[string]any{
																"name": "_nin",
																"type": map[string]any{
																	"name": nil,
																},
															},
															map[string]any{
																"name": "_or",
																"type": map[string]any{
																	"name": nil,
																},
															},
														},
													},
												},
												map[string]any{
													"name": "limit",
													"type": map[string]any{
														"name":        "Int",
														"inputFields": nil,
													},
												},
												map[string]any{
													"name": "offset",
													"type": map[string]any{
														"name":        "Int",
														"inputFields": nil,
													},
												},
											},
										},
									},
									aggregateGroupArg("NotNullInt"),
									aggregateVersionArg,
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

func TestSchemaAggregateInlineArrayCreatesUsersNillableFloatCountFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Favourites: [Float]
					}
				`,
			},
			testUtils.IntrospectionRequest{
				Request: `
					query IntrospectionQuery {
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
										"name": "Favourites",
										"type": map[string]any{
											"name": "Users__Favourites__CountSelector",
											"inputFields": []any{
												map[string]any{
													"name": "filter",
													"type": map[string]any{
														"name": "FloatFilterArg",
														"inputFields": []any{
															map[string]any{
																"name": "_and",
																"type": map[string]any{
																	"name": nil,
																},
															},
															map[string]any{
																"name": "_eq",
																"type": map[string]any{
																	"name": "Float",
																},
															},
															map[string]any{
																"name": "_ge",
																"type": map[string]any{
																	"name": "Float",
																},
															},
															map[string]any{
																"name": "_gt",
																"type": map[string]any{
																	"name": "Float",
																},
															},
															map[string]any{
																"name": "_in",
																"type": map[string]any{
																	"name": nil,
																},
															},
															map[string]any{
																"name": "_le",
																"type": map[string]any{
																	"name": "Float",
																},
															},
															map[string]any{
																"name": "_lt",
																"type": map[string]any{
																	"name": "Float",
																},
															},
															map[string]any{
																"name": "_ne",
																"type": map[string]any{
																	"name": "Float",
																},
															},
															map[string]any{
																"name": "_nin",
																"type": map[string]any{
																	"name": nil,
																},
															},
															map[string]any{
																"name": "_or",
																"type": map[string]any{
																	"name": nil,
																},
															},
														},
													},
												},
												map[string]any{
													"name": "limit",
													"type": map[string]any{
														"name":        "Int",
														"inputFields": nil,
													},
												},
												map[string]any{
													"name": "offset",
													"type": map[string]any{
														"name":        "Int",
														"inputFields": nil,
													},
												},
											},
										},
									},
									aggregateGroupArg("Float"),
									aggregateVersionArg,
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

func TestSchemaAggregateInlineArrayCreatesUsersFloatCountFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Favourites: [Float!]
					}
				`,
			},
			testUtils.IntrospectionRequest{
				Request: `
					query IntrospectionQuery {
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
										"name": "Favourites",
										"type": map[string]any{
											"name": "Users__Favourites__CountSelector",
											"inputFields": []any{
												map[string]any{
													"name": "filter",
													"type": map[string]any{
														"name": "NotNullFloatFilterArg",
														"inputFields": []any{
															map[string]any{
																"name": "_and",
																"type": map[string]any{
																	"name": nil,
																},
															},
															map[string]any{
																"name": "_eq",
																"type": map[string]any{
																	"name": "Float",
																},
															},
															map[string]any{
																"name": "_ge",
																"type": map[string]any{
																	"name": "Float",
																},
															},
															map[string]any{
																"name": "_gt",
																"type": map[string]any{
																	"name": "Float",
																},
															},
															map[string]any{
																"name": "_in",
																"type": map[string]any{
																	"name": nil,
																},
															},
															map[string]any{
																"name": "_le",
																"type": map[string]any{
																	"name": "Float",
																},
															},
															map[string]any{
																"name": "_lt",
																"type": map[string]any{
																	"name": "Float",
																},
															},
															map[string]any{
																"name": "_ne",
																"type": map[string]any{
																	"name": "Float",
																},
															},
															map[string]any{
																"name": "_nin",
																"type": map[string]any{
																	"name": nil,
																},
															},
															map[string]any{
																"name": "_or",
																"type": map[string]any{
																	"name": nil,
																},
															},
														},
													},
												},
												map[string]any{
													"name": "limit",
													"type": map[string]any{
														"name":        "Int",
														"inputFields": nil,
													},
												},
												map[string]any{
													"name": "offset",
													"type": map[string]any{
														"name":        "Int",
														"inputFields": nil,
													},
												},
											},
										},
									},
									aggregateGroupArg("NotNullFloat"),
									aggregateVersionArg,
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

func TestSchemaAggregateInlineArrayCreatesUsersNillableStringCountFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Favourites: [String]
					}
				`,
			},
			testUtils.IntrospectionRequest{
				Request: `
					query IntrospectionQuery {
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
										"name": "Favourites",
										"type": map[string]any{
											"name": "Users__Favourites__CountSelector",
											"inputFields": []any{
												map[string]any{
													"name": "filter",
													"type": map[string]any{
														"name": "StringFilterArg",
														"inputFields": []any{
															map[string]any{
																"name": "_and",
																"type": map[string]any{
																	"name": nil,
																},
															},
															map[string]any{
																"name": "_eq",
																"type": map[string]any{
																	"name": "String",
																},
															},
															map[string]any{
																"name": "_ilike",
																"type": map[string]any{
																	"name": "String",
																},
															},
															map[string]any{
																"name": "_in",
																"type": map[string]any{
																	"name": nil,
																},
															},
															map[string]any{
																"name": "_like",
																"type": map[string]any{
																	"name": "String",
																},
															},
															map[string]any{
																"name": "_ne",
																"type": map[string]any{
																	"name": "String",
																},
															},
															map[string]any{
																"name": "_nilike",
																"type": map[string]any{
																	"name": "String",
																},
															},
															map[string]any{
																"name": "_nin",
																"type": map[string]any{
																	"name": nil,
																},
															},
															map[string]any{
																"name": "_nlike",
																"type": map[string]any{
																	"name": "String",
																},
															},
															map[string]any{
																"name": "_or",
																"type": map[string]any{
																	"name": nil,
																},
															},
														},
													},
												},
												map[string]any{
													"name": "limit",
													"type": map[string]any{
														"name":        "Int",
														"inputFields": nil,
													},
												},
												map[string]any{
													"name": "offset",
													"type": map[string]any{
														"name":        "Int",
														"inputFields": nil,
													},
												},
											},
										},
									},
									aggregateGroupArg("String"),
									aggregateVersionArg,
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

func TestSchemaAggregateInlineArrayCreatesUsersStringCountFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Favourites: [String!]
					}
				`,
			},
			testUtils.IntrospectionRequest{
				Request: `
					query IntrospectionQuery {
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
										"name": "Favourites",
										"type": map[string]any{
											"name": "Users__Favourites__CountSelector",
											"inputFields": []any{
												map[string]any{
													"name": "filter",
													"type": map[string]any{
														"name": "NotNullStringFilterArg",
														"inputFields": []any{
															map[string]any{
																"name": "_and",
																"type": map[string]any{
																	"name": nil,
																},
															},
															map[string]any{
																"name": "_eq",
																"type": map[string]any{
																	"name": "String",
																},
															},
															map[string]any{
																"name": "_ilike",
																"type": map[string]any{
																	"name": "String",
																},
															},
															map[string]any{
																"name": "_in",
																"type": map[string]any{
																	"name": nil,
																},
															},
															map[string]any{
																"name": "_like",
																"type": map[string]any{
																	"name": "String",
																},
															},
															map[string]any{
																"name": "_ne",
																"type": map[string]any{
																	"name": "String",
																},
															},
															map[string]any{
																"name": "_nilike",
																"type": map[string]any{
																	"name": "String",
																},
															},
															map[string]any{
																"name": "_nin",
																"type": map[string]any{
																	"name": nil,
																},
															},
															map[string]any{
																"name": "_nlike",
																"type": map[string]any{
																	"name": "String",
																},
															},
															map[string]any{
																"name": "_or",
																"type": map[string]any{
																	"name": nil,
																},
															},
														},
													},
												},
												map[string]any{
													"name": "limit",
													"type": map[string]any{
														"name":        "Int",
														"inputFields": nil,
													},
												},
												map[string]any{
													"name": "offset",
													"type": map[string]any{
														"name":        "Int",
														"inputFields": nil,
													},
												},
											},
										},
									},
									aggregateGroupArg("NotNullString"),
									aggregateVersionArg,
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
