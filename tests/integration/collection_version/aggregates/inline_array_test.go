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

package aggregates

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestCollectionVersionAggregateInlineArrayAddsUsersCount(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
											"name": "Users___version__CountSelector",
										},
									},
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

func TestCollectionVersionAggregateInlineArrayAddsUsersSum(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
								"name": "SUM",
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
										"name": "GROUP",
										"type": map[string]any{
											"name": "Users__NumericSelector",
											"inputFields": []any{
												map[string]any{
													"name": "field",
													"type": map[string]any{
														"name": any(nil),
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
														"name": any(nil),
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

func TestCollectionVersionAggregateInlineArrayAddsUsersAverage(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
								"name": "AVG",
								"args": []any{
									map[string]any{
										"name": "GROUP",
										"type": map[string]any{
											"inputFields": []any{
												map[string]any{
													"name": "field",
													"type": map[string]any{
														"name": any(nil),
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
														"name": any(nil),
													},
												},
											},
											"name": "Users__NumericSelector",
										},
									},
									map[string]any{
										"name": "favouriteIntegers",
										"type": map[string]any{
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
											"name": "Users__favouriteIntegers__NumericSelector",
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

func aggregateGroupArg(fieldType string) map[string]any {
	return map[string]any{
		"name": "GROUP",
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
								"name": "_alias",
								"type": map[string]any{
									"name": "JSON",
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

func TestCollectionVersionAggregateInlineArrayAddsUsersNillableBooleanCountFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
								"name": "COUNT",
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
																"name": "_neq",
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

func TestCollectionVersionAggregateInlineArrayAddsUsersBooleanCountFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
								"name": "COUNT",
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
																"name": "_neq",
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

func TestCollectionVersionAggregateInlineArrayAddsUsersNillableIntegerCountFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
								"name": "COUNT",
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
																"name": "_geq",
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
																"name": "_leq",
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
																"name": "_neq",
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

func TestCollectionVersionAggregateInlineArrayAddsUsersIntegerCountFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
								"name": "COUNT",
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
																"name": "_geq",
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
																"name": "_leq",
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
																"name": "_neq",
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

func TestCollectionVersionAggregateInlineArrayAddsUsersNillableFloatCountFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
								"name": "COUNT",
								"args": []any{
									map[string]any{
										"name": "Favourites",
										"type": map[string]any{
											"name": "Users__Favourites__CountSelector",
											"inputFields": []any{
												map[string]any{
													"name": "filter",
													"type": map[string]any{
														"name": "Float64FilterArg",
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
																	"name": "Float64",
																},
															},
															map[string]any{
																"name": "_geq",
																"type": map[string]any{
																	"name": "Float64",
																},
															},
															map[string]any{
																"name": "_gt",
																"type": map[string]any{
																	"name": "Float64",
																},
															},
															map[string]any{
																"name": "_in",
																"type": map[string]any{
																	"name": nil,
																},
															},
															map[string]any{
																"name": "_leq",
																"type": map[string]any{
																	"name": "Float64",
																},
															},
															map[string]any{
																"name": "_lt",
																"type": map[string]any{
																	"name": "Float64",
																},
															},
															map[string]any{
																"name": "_neq",
																"type": map[string]any{
																	"name": "Float64",
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
									aggregateGroupArg("Float64"),
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

func TestCollectionVersionAggregateInlineArrayAddsUsersFloatCountFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
								"name": "COUNT",
								"args": []any{
									map[string]any{
										"name": "Favourites",
										"type": map[string]any{
											"name": "Users__Favourites__CountSelector",
											"inputFields": []any{
												map[string]any{
													"name": "filter",
													"type": map[string]any{
														"name": "NotNullFloat64FilterArg",
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
																	"name": "Float64",
																},
															},
															map[string]any{
																"name": "_geq",
																"type": map[string]any{
																	"name": "Float64",
																},
															},
															map[string]any{
																"name": "_gt",
																"type": map[string]any{
																	"name": "Float64",
																},
															},
															map[string]any{
																"name": "_in",
																"type": map[string]any{
																	"name": nil,
																},
															},
															map[string]any{
																"name": "_leq",
																"type": map[string]any{
																	"name": "Float64",
																},
															},
															map[string]any{
																"name": "_lt",
																"type": map[string]any{
																	"name": "Float64",
																},
															},
															map[string]any{
																"name": "_neq",
																"type": map[string]any{
																	"name": "Float64",
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
									aggregateGroupArg("NotNullFloat64"),
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

func TestCollectionVersionAggregateInlineArrayAddsUsersNillableStringCountFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
								"name": "COUNT",
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
																"name": "_neq",
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

func TestCollectionVersionAggregateInlineArrayAddsUsersStringCountFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
								"name": "COUNT",
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
																"name": "_neq",
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
