// Copyright 2023 Democratized Data Foundation
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
	test := testUtils.RequestTestCase{
		Schema: []string{
			`
				type users {
					FavouriteIntegers: [Int!]
				}
			`,
		},
		IntrospectionRequest: `
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
		ContainsData: map[string]any{
			"__type": map[string]any{
				"name": "users",
				"fields": []any{
					map[string]any{
						"name": "_count",
						"args": []any{
							map[string]any{
								"name": "FavouriteIntegers",
								"type": map[string]any{
									"name": "users__FavouriteIntegers__CountSelector",
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
									"name": "users__CountSelector",
									"inputFields": []any{
										map[string]any{
											"name": "filter",
											"type": map[string]any{
												"name": "usersFilterArg",
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
									"name": "users___version__CountSelector",
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
	}

	testUtils.ExecuteRequestTestCase(t, test)
}

func TestSchemaAggregateInlineArrayCreatesUsersSum(t *testing.T) {
	test := testUtils.RequestTestCase{
		Schema: []string{
			`
				type users {
					FavouriteFloats: [Float!]
				}
			`,
		},
		IntrospectionRequest: `
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
		ContainsData: map[string]any{
			"__type": map[string]any{
				"name": "users",
				"fields": []any{
					map[string]any{
						"name": "_sum",
						"args": []any{
							map[string]any{
								"name": "FavouriteFloats",
								"type": map[string]any{
									"name": "users__FavouriteFloats__NumericSelector",
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
									"name": "users__NumericSelector",
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
												"name": "usersFilterArg",
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
												"name": "usersOrderArg",
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

	testUtils.ExecuteRequestTestCase(t, test)
}

func TestSchemaAggregateInlineArrayCreatesUsersAverage(t *testing.T) {
	test := testUtils.RequestTestCase{
		Schema: []string{
			`
				type users {
					FavouriteIntegers: [Int!]
				}
			`,
		},
		IntrospectionRequest: `
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
		ContainsData: map[string]any{
			"__type": map[string]any{
				"name": "users",
				"fields": []any{
					map[string]any{
						"name": "_avg",
						"args": []any{
							map[string]any{
								"name": "FavouriteIntegers",
								"type": map[string]any{
									"name": "users__FavouriteIntegers__NumericSelector",
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
									"name": "users__NumericSelector",
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
												"name": "usersFilterArg",
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
												"name": "usersOrderArg",
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

	testUtils.ExecuteRequestTestCase(t, test)
}

var aggregateGroupArg = map[string]any{
	"name": "_group",
	"type": map[string]any{
		"name": "users__CountSelector",
		"inputFields": []any{
			map[string]any{
				"name": "filter",
				"type": map[string]any{
					"name": "usersFilterArg",
					"inputFields": []any{
						map[string]any{
							"name": "_and",
							"type": map[string]any{
								"name": nil,
							},
						},
						map[string]any{
							"name": "_key",
							"type": map[string]any{
								"name": "IDOperatorBlock",
							},
						},
						map[string]any{
							"name": "_not",
							"type": map[string]any{
								"name": "usersFilterArg",
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

var aggregateVersionArg = map[string]any{
	"name": "_version",
	"type": map[string]any{
		"name": "users___version__CountSelector",
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
	test := testUtils.RequestTestCase{
		Schema: []string{
			`
				type users {
					Favourites: [Boolean]
				}
			`,
		},
		IntrospectionRequest: `
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
		ContainsData: map[string]any{
			"__type": map[string]any{
				"name": "users",
				"fields": []any{
					map[string]any{
						"name": "_count",
						"args": []any{
							map[string]any{
								"name": "Favourites",
								"type": map[string]any{
									"name": "users__Favourites__CountSelector",
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
							aggregateGroupArg,
							aggregateVersionArg,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteRequestTestCase(t, test)
}

func TestSchemaAggregateInlineArrayCreatesUsersBooleanCountFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Schema: []string{
			`
				type users {
					Favourites: [Boolean!]
				}
			`,
		},
		IntrospectionRequest: `
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
		ContainsData: map[string]any{
			"__type": map[string]any{
				"name": "users",
				"fields": []any{
					map[string]any{
						"name": "_count",
						"args": []any{
							map[string]any{
								"name": "Favourites",
								"type": map[string]any{
									"name": "users__Favourites__CountSelector",
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
							aggregateGroupArg,
							aggregateVersionArg,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteRequestTestCase(t, test)
}

func TestSchemaAggregateInlineArrayCreatesUsersNillableIntegerCountFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Schema: []string{
			`
				type users {
					Favourites: [Int]
				}
			`,
		},
		IntrospectionRequest: `
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
		ContainsData: map[string]any{
			"__type": map[string]any{
				"name": "users",
				"fields": []any{
					map[string]any{
						"name": "_count",
						"args": []any{
							map[string]any{
								"name": "Favourites",
								"type": map[string]any{
									"name": "users__Favourites__CountSelector",
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
							aggregateGroupArg,
							aggregateVersionArg,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteRequestTestCase(t, test)
}

func TestSchemaAggregateInlineArrayCreatesUsersIntegerCountFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Schema: []string{
			`
				type users {
					Favourites: [Int!]
				}
			`,
		},
		IntrospectionRequest: `
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
		ContainsData: map[string]any{
			"__type": map[string]any{
				"name": "users",
				"fields": []any{
					map[string]any{
						"name": "_count",
						"args": []any{
							map[string]any{
								"name": "Favourites",
								"type": map[string]any{
									"name": "users__Favourites__CountSelector",
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
							aggregateGroupArg,
							aggregateVersionArg,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteRequestTestCase(t, test)
}

func TestSchemaAggregateInlineArrayCreatesUsersNillableFloatCountFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Schema: []string{
			`
				type users {
					Favourites: [Float]
				}
			`,
		},
		IntrospectionRequest: `
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
		ContainsData: map[string]any{
			"__type": map[string]any{
				"name": "users",
				"fields": []any{
					map[string]any{
						"name": "_count",
						"args": []any{
							map[string]any{
								"name": "Favourites",
								"type": map[string]any{
									"name": "users__Favourites__CountSelector",
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
							aggregateGroupArg,
							aggregateVersionArg,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteRequestTestCase(t, test)
}

func TestSchemaAggregateInlineArrayCreatesUsersFloatCountFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Schema: []string{
			`
				type users {
					Favourites: [Float!]
				}
			`,
		},
		IntrospectionRequest: `
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
		ContainsData: map[string]any{
			"__type": map[string]any{
				"name": "users",
				"fields": []any{
					map[string]any{
						"name": "_count",
						"args": []any{
							map[string]any{
								"name": "Favourites",
								"type": map[string]any{
									"name": "users__Favourites__CountSelector",
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
							aggregateGroupArg,
							aggregateVersionArg,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteRequestTestCase(t, test)
}

func TestSchemaAggregateInlineArrayCreatesUsersNillableStringCountFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Schema: []string{
			`
				type users {
					Favourites: [String]
				}
			`,
		},
		IntrospectionRequest: `
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
		ContainsData: map[string]any{
			"__type": map[string]any{
				"name": "users",
				"fields": []any{
					map[string]any{
						"name": "_count",
						"args": []any{
							map[string]any{
								"name": "Favourites",
								"type": map[string]any{
									"name": "users__Favourites__CountSelector",
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
														"name": "_in",
														"type": map[string]any{
															"name": nil,
														},
													},
													map[string]any{
														"name": "_ne",
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
							aggregateGroupArg,
							aggregateVersionArg,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteRequestTestCase(t, test)
}

func TestSchemaAggregateInlineArrayCreatesUsersStringCountFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Schema: []string{
			`
				type users {
					Favourites: [String!]
				}
			`,
		},
		IntrospectionRequest: `
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
		ContainsData: map[string]any{
			"__type": map[string]any{
				"name": "users",
				"fields": []any{
					map[string]any{
						"name": "_count",
						"args": []any{
							map[string]any{
								"name": "Favourites",
								"type": map[string]any{
									"name": "users__Favourites__CountSelector",
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
														"name": "_in",
														"type": map[string]any{
															"name": nil,
														},
													},
													map[string]any{
														"name": "_ne",
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
							aggregateGroupArg,
							aggregateVersionArg,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteRequestTestCase(t, test)
}
