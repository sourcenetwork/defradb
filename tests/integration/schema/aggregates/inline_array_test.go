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

	testUtils "github.com/sourcenetwork/defradb/tests/integration/schema"
)

func TestSchemaAggregateInlineArrayCreatesUsersCount(t *testing.T) {
	test := testUtils.QueryTestCase{
		Schema: []string{
			`
				type users {
					FavouriteIntegers: [Int!]
				}
			`,
		},
		IntrospectionQuery: `
			query IntrospectionQuery {
				__type (name: "users") {
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
		ContainsData: map[string]interface{}{
			"__type": map[string]interface{}{
				"name": "users",
				"fields": []interface{}{
					map[string]interface{}{
						"name": "_count",
						"args": []interface{}{
							map[string]interface{}{
								"name": "FavouriteIntegers",
								"type": map[string]interface{}{
									"name": "users__FavouriteIntegers__CountSelector",
									"inputFields": []interface{}{
										map[string]interface{}{
											"name": "filter",
											"type": map[string]interface{}{
												"name": "NotNullIntFilterArg",
											},
										},
										map[string]interface{}{
											"name": "limit",
											"type": map[string]interface{}{
												"name": "Int",
											},
										},
										map[string]interface{}{
											"name": "offset",
											"type": map[string]interface{}{
												"name": "Int",
											},
										},
									},
								},
							},
							map[string]interface{}{
								"name": "_group",
								"type": map[string]interface{}{
									"name": "users__CountSelector",
									"inputFields": []interface{}{
										map[string]interface{}{
											"name": "filter",
											"type": map[string]interface{}{
												"name": "usersFilterArg",
											},
										},
										map[string]interface{}{
											"name": "limit",
											"type": map[string]interface{}{
												"name": "Int",
											},
										},
									},
								},
							},
							map[string]interface{}{
								"name": "_version",
								"type": map[string]interface{}{
									"name": "users___version__CountSelector",
									"inputFields": []interface{}{
										map[string]interface{}{
											"name": "limit",
											"type": map[string]interface{}{
												"name": "Int",
											},
										},
										map[string]interface{}{
											"name": "offset",
											"type": map[string]interface{}{
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
	}

	testUtils.ExecuteQueryTestCase(t, test)
}

func TestSchemaAggregateInlineArrayCreatesUsersSum(t *testing.T) {
	test := testUtils.QueryTestCase{
		Schema: []string{
			`
				type users {
					FavouriteFloats: [Float!]
				}
			`,
		},
		IntrospectionQuery: `
			query IntrospectionQuery {
				__type (name: "users") {
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
		ContainsData: map[string]interface{}{
			"__type": map[string]interface{}{
				"name": "users",
				"fields": []interface{}{
					map[string]interface{}{
						"name": "_sum",
						"args": []interface{}{
							map[string]interface{}{
								"name": "FavouriteFloats",
								"type": map[string]interface{}{
									"name": "users__FavouriteFloats__NumericSelector",
									"inputFields": []interface{}{
										map[string]interface{}{
											"name": "filter",
											"type": map[string]interface{}{
												"name": "NotNullFloatFilterArg",
											},
										},
										map[string]interface{}{
											"name": "limit",
											"type": map[string]interface{}{
												"name": "Int",
											},
										},
									},
								},
							},
							map[string]interface{}{
								"name": "_group",
								"type": map[string]interface{}{
									"name": "users__NumericSelector",
									"inputFields": []interface{}{
										map[string]interface{}{
											"name": "field",
											"type": map[string]interface{}{
												"name": "usersNumericFieldsArg",
											},
										},
										map[string]interface{}{
											"name": "filter",
											"type": map[string]interface{}{
												"name": "usersFilterArg",
											},
										},
										map[string]interface{}{
											"name": "limit",
											"type": map[string]interface{}{
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
	}

	testUtils.ExecuteQueryTestCase(t, test)
}

func TestSchemaAggregateInlineArrayCreatesUsersAverage(t *testing.T) {
	test := testUtils.QueryTestCase{
		Schema: []string{
			`
				type users {
					FavouriteIntegers: [Int!]
				}
			`,
		},
		IntrospectionQuery: `
			query IntrospectionQuery {
				__type (name: "users") {
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
		ContainsData: map[string]interface{}{
			"__type": map[string]interface{}{
				"name": "users",
				"fields": []interface{}{
					map[string]interface{}{
						"name": "_avg",
						"args": []interface{}{
							map[string]interface{}{
								"name": "FavouriteIntegers",
								"type": map[string]interface{}{
									"name": "users__FavouriteIntegers__NumericSelector",
									"inputFields": []interface{}{
										map[string]interface{}{
											"name": "filter",
											"type": map[string]interface{}{
												"name": "NotNullIntFilterArg",
											},
										},
										map[string]interface{}{
											"name": "limit",
											"type": map[string]interface{}{
												"name": "Int",
											},
										},
									},
								},
							},
							map[string]interface{}{
								"name": "_group",
								"type": map[string]interface{}{
									"name": "users__NumericSelector",
									"inputFields": []interface{}{
										map[string]interface{}{
											"name": "field",
											"type": map[string]interface{}{
												"name": "usersNumericFieldsArg",
											},
										},
										map[string]interface{}{
											"name": "filter",
											"type": map[string]interface{}{
												"name": "usersFilterArg",
											},
										},
										map[string]interface{}{
											"name": "limit",
											"type": map[string]interface{}{
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
	}

	testUtils.ExecuteQueryTestCase(t, test)
}

var aggregateGroupArg = map[string]interface{}{
	"name": "_group",
	"type": map[string]interface{}{
		"name": "users__CountSelector",
		"inputFields": []interface{}{
			map[string]interface{}{
				"name": "filter",
				"type": map[string]interface{}{
					"name": "usersFilterArg",
					"inputFields": []interface{}{
						map[string]interface{}{
							"name": "_and",
							"type": map[string]interface{}{
								"name": nil,
							},
						},
						map[string]interface{}{
							"name": "_key",
							"type": map[string]interface{}{
								"name": "IDOperatorBlock",
							},
						},
						map[string]interface{}{
							"name": "_not",
							"type": map[string]interface{}{
								"name": "usersFilterArg",
							},
						},
						map[string]interface{}{
							"name": "_or",
							"type": map[string]interface{}{
								"name": nil,
							},
						},
					},
				},
			},
			map[string]interface{}{
				"name": "limit",
				"type": map[string]interface{}{
					"name":        "Int",
					"inputFields": nil,
				},
			},
		},
	},
}

var aggregateVersionArg = map[string]interface{}{
	"name": "_version",
	"type": map[string]interface{}{
		"name": "users___version__CountSelector",
		"inputFields": []interface{}{
			map[string]interface{}{
				"name": "limit",
				"type": map[string]interface{}{
					"name":        "Int",
					"inputFields": nil,
				},
			},
		},
	},
}

func TestSchemaAggregateInlineArrayCreatesUsersNillableBooleanCountFilter(t *testing.T) {
	test := testUtils.QueryTestCase{
		Schema: []string{
			`
				type users {
					Favourites: [Boolean]
				}
			`,
		},
		IntrospectionQuery: `
			query IntrospectionQuery {
				__type (name: "users") {
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
		ContainsData: map[string]interface{}{
			"__type": map[string]interface{}{
				"name": "users",
				"fields": []interface{}{
					map[string]interface{}{
						"name": "_count",
						"args": []interface{}{
							map[string]interface{}{
								"name": "Favourites",
								"type": map[string]interface{}{
									"name": "users__Favourites__CountSelector",
									"inputFields": []interface{}{
										map[string]interface{}{
											"name": "filter",
											"type": map[string]interface{}{
												"name": "BooleanFilterArg",
												"inputFields": []interface{}{
													map[string]interface{}{
														"name": "_and",
														"type": map[string]interface{}{
															"name": nil,
														},
													},
													map[string]interface{}{
														"name": "_eq",
														"type": map[string]interface{}{
															"name": "Boolean",
														},
													},
													map[string]interface{}{
														"name": "_in",
														"type": map[string]interface{}{
															"name": nil,
														},
													},
													map[string]interface{}{
														"name": "_like",
														"type": map[string]interface{}{
															"name": "Boolean",
														},
													},
													map[string]interface{}{
														"name": "_ne",
														"type": map[string]interface{}{
															"name": "Boolean",
														},
													},
													map[string]interface{}{
														"name": "_nin",
														"type": map[string]interface{}{
															"name": nil,
														},
													},
													map[string]interface{}{
														"name": "_or",
														"type": map[string]interface{}{
															"name": nil,
														},
													},
												},
											},
										},
										map[string]interface{}{
											"name": "limit",
											"type": map[string]interface{}{
												"name":        "Int",
												"inputFields": nil,
											},
										},
									},
								},
							},
							aggregateGroupArg,
							aggregateVersionArg,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteQueryTestCase(t, test)
}

func TestSchemaAggregateInlineArrayCreatesUsersBooleanCountFilter(t *testing.T) {
	test := testUtils.QueryTestCase{
		Schema: []string{
			`
				type users {
					Favourites: [Boolean!]
				}
			`,
		},
		IntrospectionQuery: `
			query IntrospectionQuery {
				__type (name: "users") {
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
		ContainsData: map[string]interface{}{
			"__type": map[string]interface{}{
				"name": "users",
				"fields": []interface{}{
					map[string]interface{}{
						"name": "_count",
						"args": []interface{}{
							map[string]interface{}{
								"name": "Favourites",
								"type": map[string]interface{}{
									"name": "users__Favourites__CountSelector",
									"inputFields": []interface{}{
										map[string]interface{}{
											"name": "filter",
											"type": map[string]interface{}{
												"name": "NotNullBooleanFilterArg",
												"inputFields": []interface{}{
													map[string]interface{}{
														"name": "_and",
														"type": map[string]interface{}{
															"name": nil,
														},
													},
													map[string]interface{}{
														"name": "_eq",
														"type": map[string]interface{}{
															"name": "Boolean",
														},
													},
													map[string]interface{}{
														"name": "_in",
														"type": map[string]interface{}{
															"name": nil,
														},
													},
													map[string]interface{}{
														"name": "_ne",
														"type": map[string]interface{}{
															"name": "Boolean",
														},
													},
													map[string]interface{}{
														"name": "_nin",
														"type": map[string]interface{}{
															"name": nil,
														},
													},
													map[string]interface{}{
														"name": "_or",
														"type": map[string]interface{}{
															"name": nil,
														},
													},
												},
											},
										},
										map[string]interface{}{
											"name": "limit",
											"type": map[string]interface{}{
												"name":        "Int",
												"inputFields": nil,
											},
										},
									},
								},
							},
							aggregateGroupArg,
							aggregateVersionArg,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteQueryTestCase(t, test)
}

func TestSchemaAggregateInlineArrayCreatesUsersNillableIntegerCountFilter(t *testing.T) {
	test := testUtils.QueryTestCase{
		Schema: []string{
			`
				type users {
					Favourites: [Int]
				}
			`,
		},
		IntrospectionQuery: `
			query IntrospectionQuery {
				__type (name: "users") {
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
		ContainsData: map[string]interface{}{
			"__type": map[string]interface{}{
				"name": "users",
				"fields": []interface{}{
					map[string]interface{}{
						"name": "_count",
						"args": []interface{}{
							map[string]interface{}{
								"name": "Favourites",
								"type": map[string]interface{}{
									"name": "users__Favourites__CountSelector",
									"inputFields": []interface{}{
										map[string]interface{}{
											"name": "filter",
											"type": map[string]interface{}{
												"name": "IntFilterArg",
												"inputFields": []interface{}{
													map[string]interface{}{
														"name": "_and",
														"type": map[string]interface{}{
															"name": nil,
														},
													},
													map[string]interface{}{
														"name": "_eq",
														"type": map[string]interface{}{
															"name": "Int",
														},
													},
													map[string]interface{}{
														"name": "_ge",
														"type": map[string]interface{}{
															"name": "Int",
														},
													},
													map[string]interface{}{
														"name": "_gt",
														"type": map[string]interface{}{
															"name": "Int",
														},
													},
													map[string]interface{}{
														"name": "_in",
														"type": map[string]interface{}{
															"name": nil,
														},
													},
													map[string]interface{}{
														"name": "_le",
														"type": map[string]interface{}{
															"name": "Int",
														},
													},
													map[string]interface{}{
														"name": "_lt",
														"type": map[string]interface{}{
															"name": "Int",
														},
													},
													map[string]interface{}{
														"name": "_ne",
														"type": map[string]interface{}{
															"name": "Int",
														},
													},
													map[string]interface{}{
														"name": "_nin",
														"type": map[string]interface{}{
															"name": nil,
														},
													},
													map[string]interface{}{
														"name": "_or",
														"type": map[string]interface{}{
															"name": nil,
														},
													},
												},
											},
										},
										map[string]interface{}{
											"name": "limit",
											"type": map[string]interface{}{
												"name":        "Int",
												"inputFields": nil,
											},
										},
									},
								},
							},
							aggregateGroupArg,
							aggregateVersionArg,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteQueryTestCase(t, test)
}

func TestSchemaAggregateInlineArrayCreatesUsersIntegerCountFilter(t *testing.T) {
	test := testUtils.QueryTestCase{
		Schema: []string{
			`
				type users {
					Favourites: [Int!]
				}
			`,
		},
		IntrospectionQuery: `
			query IntrospectionQuery {
				__type (name: "users") {
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
		ContainsData: map[string]interface{}{
			"__type": map[string]interface{}{
				"name": "users",
				"fields": []interface{}{
					map[string]interface{}{
						"name": "_count",
						"args": []interface{}{
							map[string]interface{}{
								"name": "Favourites",
								"type": map[string]interface{}{
									"name": "users__Favourites__CountSelector",
									"inputFields": []interface{}{
										map[string]interface{}{
											"name": "filter",
											"type": map[string]interface{}{
												"name": "NotNullIntFilterArg",
												"inputFields": []interface{}{
													map[string]interface{}{
														"name": "_and",
														"type": map[string]interface{}{
															"name": nil,
														},
													},
													map[string]interface{}{
														"name": "_eq",
														"type": map[string]interface{}{
															"name": "Int",
														},
													},
													map[string]interface{}{
														"name": "_ge",
														"type": map[string]interface{}{
															"name": "Int",
														},
													},
													map[string]interface{}{
														"name": "_gt",
														"type": map[string]interface{}{
															"name": "Int",
														},
													},
													map[string]interface{}{
														"name": "_in",
														"type": map[string]interface{}{
															"name": nil,
														},
													},
													map[string]interface{}{
														"name": "_le",
														"type": map[string]interface{}{
															"name": "Int",
														},
													},
													map[string]interface{}{
														"name": "_lt",
														"type": map[string]interface{}{
															"name": "Int",
														},
													},
													map[string]interface{}{
														"name": "_ne",
														"type": map[string]interface{}{
															"name": "Int",
														},
													},
													map[string]interface{}{
														"name": "_nin",
														"type": map[string]interface{}{
															"name": nil,
														},
													},
													map[string]interface{}{
														"name": "_or",
														"type": map[string]interface{}{
															"name": nil,
														},
													},
												},
											},
										},
										map[string]interface{}{
											"name": "limit",
											"type": map[string]interface{}{
												"name":        "Int",
												"inputFields": nil,
											},
										},
									},
								},
							},
							aggregateGroupArg,
							aggregateVersionArg,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteQueryTestCase(t, test)
}

func TestSchemaAggregateInlineArrayCreatesUsersNillableFloatCountFilter(t *testing.T) {
	test := testUtils.QueryTestCase{
		Schema: []string{
			`
				type users {
					Favourites: [Float]
				}
			`,
		},
		IntrospectionQuery: `
			query IntrospectionQuery {
				__type (name: "users") {
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
		ContainsData: map[string]interface{}{
			"__type": map[string]interface{}{
				"name": "users",
				"fields": []interface{}{
					map[string]interface{}{
						"name": "_count",
						"args": []interface{}{
							map[string]interface{}{
								"name": "Favourites",
								"type": map[string]interface{}{
									"name": "users__Favourites__CountSelector",
									"inputFields": []interface{}{
										map[string]interface{}{
											"name": "filter",
											"type": map[string]interface{}{
												"name": "FloatFilterArg",
												"inputFields": []interface{}{
													map[string]interface{}{
														"name": "_and",
														"type": map[string]interface{}{
															"name": nil,
														},
													},
													map[string]interface{}{
														"name": "_eq",
														"type": map[string]interface{}{
															"name": "Float",
														},
													},
													map[string]interface{}{
														"name": "_ge",
														"type": map[string]interface{}{
															"name": "Float",
														},
													},
													map[string]interface{}{
														"name": "_gt",
														"type": map[string]interface{}{
															"name": "Float",
														},
													},
													map[string]interface{}{
														"name": "_in",
														"type": map[string]interface{}{
															"name": nil,
														},
													},
													map[string]interface{}{
														"name": "_le",
														"type": map[string]interface{}{
															"name": "Float",
														},
													},
													map[string]interface{}{
														"name": "_lt",
														"type": map[string]interface{}{
															"name": "Float",
														},
													},
													map[string]interface{}{
														"name": "_ne",
														"type": map[string]interface{}{
															"name": "Float",
														},
													},
													map[string]interface{}{
														"name": "_nin",
														"type": map[string]interface{}{
															"name": nil,
														},
													},
													map[string]interface{}{
														"name": "_or",
														"type": map[string]interface{}{
															"name": nil,
														},
													},
												},
											},
										},
										map[string]interface{}{
											"name": "limit",
											"type": map[string]interface{}{
												"name":        "Int",
												"inputFields": nil,
											},
										},
									},
								},
							},
							aggregateGroupArg,
							aggregateVersionArg,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteQueryTestCase(t, test)
}

func TestSchemaAggregateInlineArrayCreatesUsersFloatCountFilter(t *testing.T) {
	test := testUtils.QueryTestCase{
		Schema: []string{
			`
				type users {
					Favourites: [Float!]
				}
			`,
		},
		IntrospectionQuery: `
			query IntrospectionQuery {
				__type (name: "users") {
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
		ContainsData: map[string]interface{}{
			"__type": map[string]interface{}{
				"name": "users",
				"fields": []interface{}{
					map[string]interface{}{
						"name": "_count",
						"args": []interface{}{
							map[string]interface{}{
								"name": "Favourites",
								"type": map[string]interface{}{
									"name": "users__Favourites__CountSelector",
									"inputFields": []interface{}{
										map[string]interface{}{
											"name": "filter",
											"type": map[string]interface{}{
												"name": "NotNullFloatFilterArg",
												"inputFields": []interface{}{
													map[string]interface{}{
														"name": "_and",
														"type": map[string]interface{}{
															"name": nil,
														},
													},
													map[string]interface{}{
														"name": "_eq",
														"type": map[string]interface{}{
															"name": "Float",
														},
													},
													map[string]interface{}{
														"name": "_ge",
														"type": map[string]interface{}{
															"name": "Float",
														},
													},
													map[string]interface{}{
														"name": "_gt",
														"type": map[string]interface{}{
															"name": "Float",
														},
													},
													map[string]interface{}{
														"name": "_in",
														"type": map[string]interface{}{
															"name": nil,
														},
													},
													map[string]interface{}{
														"name": "_le",
														"type": map[string]interface{}{
															"name": "Float",
														},
													},
													map[string]interface{}{
														"name": "_lt",
														"type": map[string]interface{}{
															"name": "Float",
														},
													},
													map[string]interface{}{
														"name": "_ne",
														"type": map[string]interface{}{
															"name": "Float",
														},
													},
													map[string]interface{}{
														"name": "_nin",
														"type": map[string]interface{}{
															"name": nil,
														},
													},
													map[string]interface{}{
														"name": "_or",
														"type": map[string]interface{}{
															"name": nil,
														},
													},
												},
											},
										},
										map[string]interface{}{
											"name": "limit",
											"type": map[string]interface{}{
												"name":        "Int",
												"inputFields": nil,
											},
										},
									},
								},
							},
							aggregateGroupArg,
							aggregateVersionArg,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteQueryTestCase(t, test)
}

func TestSchemaAggregateInlineArrayCreatesUsersNillableStringCountFilter(t *testing.T) {
	test := testUtils.QueryTestCase{
		Schema: []string{
			`
				type users {
					Favourites: [String]
				}
			`,
		},
		IntrospectionQuery: `
			query IntrospectionQuery {
				__type (name: "users") {
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
		ContainsData: map[string]interface{}{
			"__type": map[string]interface{}{
				"name": "users",
				"fields": []interface{}{
					map[string]interface{}{
						"name": "_count",
						"args": []interface{}{
							map[string]interface{}{
								"name": "Favourites",
								"type": map[string]interface{}{
									"name": "users__Favourites__CountSelector",
									"inputFields": []interface{}{
										map[string]interface{}{
											"name": "filter",
											"type": map[string]interface{}{
												"name": "StringFilterArg",
												"inputFields": []interface{}{
													map[string]interface{}{
														"name": "_and",
														"type": map[string]interface{}{
															"name": nil,
														},
													},
													map[string]interface{}{
														"name": "_eq",
														"type": map[string]interface{}{
															"name": "String",
														},
													},
													map[string]interface{}{
														"name": "_in",
														"type": map[string]interface{}{
															"name": nil,
														},
													},
													map[string]interface{}{
														"name": "_like",
														"type": map[string]interface{}{
															"name": "String",
														},
													},
													map[string]interface{}{
														"name": "_ne",
														"type": map[string]interface{}{
															"name": "String",
														},
													},
													map[string]interface{}{
														"name": "_nin",
														"type": map[string]interface{}{
															"name": nil,
														},
													},
													map[string]interface{}{
														"name": "_or",
														"type": map[string]interface{}{
															"name": nil,
														},
													},
												},
											},
										},
										map[string]interface{}{
											"name": "limit",
											"type": map[string]interface{}{
												"name":        "Int",
												"inputFields": nil,
											},
										},
									},
								},
							},
							aggregateGroupArg,
							aggregateVersionArg,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteQueryTestCase(t, test)
}

func TestSchemaAggregateInlineArrayCreatesUsersStringCountFilter(t *testing.T) {
	test := testUtils.QueryTestCase{
		Schema: []string{
			`
				type users {
					Favourites: [String!]
				}
			`,
		},
		IntrospectionQuery: `
			query IntrospectionQuery {
				__type (name: "users") {
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
		ContainsData: map[string]interface{}{
			"__type": map[string]interface{}{
				"name": "users",
				"fields": []interface{}{
					map[string]interface{}{
						"name": "_count",
						"args": []interface{}{
							map[string]interface{}{
								"name": "Favourites",
								"type": map[string]interface{}{
									"name": "users__Favourites__CountSelector",
									"inputFields": []interface{}{
										map[string]interface{}{
											"name": "filter",
											"type": map[string]interface{}{
												"name": "NotNullStringFilterArg",
												"inputFields": []interface{}{
													map[string]interface{}{
														"name": "_and",
														"type": map[string]interface{}{
															"name": nil,
														},
													},
													map[string]interface{}{
														"name": "_eq",
														"type": map[string]interface{}{
															"name": "String",
														},
													},
													map[string]interface{}{
														"name": "_in",
														"type": map[string]interface{}{
															"name": nil,
														},
													},
													map[string]interface{}{
														"name": "_ne",
														"type": map[string]interface{}{
															"name": "String",
														},
													},
													map[string]interface{}{
														"name": "_nin",
														"type": map[string]interface{}{
															"name": nil,
														},
													},
													map[string]interface{}{
														"name": "_or",
														"type": map[string]interface{}{
															"name": nil,
														},
													},
												},
											},
										},
										map[string]interface{}{
											"name": "limit",
											"type": map[string]interface{}{
												"name":        "Int",
												"inputFields": nil,
											},
										},
									},
								},
							},
							aggregateGroupArg,
							aggregateVersionArg,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteQueryTestCase(t, test)
}
