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

func TestCollectionVersionAggregateTopLevelAddsCountGivenCollection(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {}
				`,
			},
			testUtils.IntrospectionRequest{
				Request: `
					query {
						__schema {
							queryType {
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
					}
				`,
				ContainsData: map[string]any{
					"__schema": map[string]any{
						"queryType": map[string]any{
							"fields": []any{
								map[string]any{
									"name": "COUNT",
									"args": []any{
										map[string]any{
											"name": "Users",
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

func TestCollectionVersionAggregateTopLevelAddsSumGivenCollection(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {}
				`,
			},
			testUtils.IntrospectionRequest{
				Request: `
					query {
						__schema {
							queryType {
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
					}
				`,
				ContainsData: map[string]any{
					"__schema": map[string]any{
						"queryType": map[string]any{
							"fields": []any{
								map[string]any{
									"name": "SUM",
									"args": []any{
										map[string]any{
											"name": "Users",
											"type": map[string]any{
												"name": "Users__NumericSelector",
												"inputFields": []any{
													map[string]any{
														"name": "field",
														"type": map[string]any{
															"name": nil,
															"kind": "NON_NULL",
															"ofType": map[string]any{
																"name": "UsersNumericFieldsArg",
															},
														},
													},
													map[string]any{
														"name": "filter",
														"type": map[string]any{
															"name":   "UsersFilterArg",
															"kind":   "INPUT_OBJECT",
															"ofType": nil,
														},
													},
													map[string]any{
														"name": "limit",
														"type": map[string]any{
															"name":   "Int",
															"kind":   "SCALAR",
															"ofType": nil,
														},
													},
													map[string]any{
														"name": "offset",
														"type": map[string]any{
															"name":   "Int",
															"kind":   "SCALAR",
															"ofType": nil,
														},
													},
													map[string]any{
														"name": "order",
														"type": map[string]any{
															"name": nil,
															"kind": "LIST",
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionAggregateTopLevelAddsAverageGivenCollection(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {}
				`,
			},
			testUtils.IntrospectionRequest{
				Request: `
					query {
						__schema {
							queryType {
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
					}
				`,
				ContainsData: map[string]any{
					"__schema": map[string]any{
						"queryType": map[string]any{
							"fields": []any{
								map[string]any{
									"name": "AVG",
									"args": []any{
										map[string]any{
											"name": "Users",
											"type": map[string]any{
												"name": "Users__NumericSelector",
												"inputFields": []any{
													map[string]any{
														"name": "field",
														"type": map[string]any{
															"name": nil,
															"kind": "NON_NULL",
															"ofType": map[string]any{
																"name": "UsersNumericFieldsArg",
															},
														},
													},
													map[string]any{
														"name": "filter",
														"type": map[string]any{
															"name":   "UsersFilterArg",
															"kind":   "INPUT_OBJECT",
															"ofType": nil,
														},
													},
													map[string]any{
														"name": "limit",
														"type": map[string]any{
															"name":   "Int",
															"kind":   "SCALAR",
															"ofType": nil,
														},
													},
													map[string]any{
														"name": "offset",
														"type": map[string]any{
															"name":   "Int",
															"kind":   "SCALAR",
															"ofType": nil,
														},
													},
													map[string]any{
														"name": "order",
														"type": map[string]any{
															"name": nil,
															"kind": "LIST",
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
