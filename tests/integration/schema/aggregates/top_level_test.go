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

func TestSchemaAggregateTopLevelCreatesCountGivenSchema(t *testing.T) {
	test := testUtils.QueryTestCase{
		Schema: []string{
			`
				type users {}
			`,
		},
		IntrospectionQuery: `
			query IntrospectionQuery {
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
		ContainsData: map[string]interface{}{
			"__schema": map[string]interface{}{
				"queryType": map[string]interface{}{
					"fields": []interface{}{
						map[string]interface{}{
							"name": "_count",
							"args": []interface{}{
								map[string]interface{}{
									"name": "users",
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
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteQueryTestCase(t, test)
}

func TestSchemaAggregateTopLevelCreatesSumGivenSchema(t *testing.T) {
	test := testUtils.QueryTestCase{
		Schema: []string{
			`
				type users {}
			`,
		},
		IntrospectionQuery: `
			query IntrospectionQuery {
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
		ContainsData: map[string]interface{}{
			"__schema": map[string]interface{}{
				"queryType": map[string]interface{}{
					"fields": []interface{}{
						map[string]interface{}{
							"name": "_sum",
							"args": []interface{}{
								map[string]interface{}{
									"name": "users",
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
		},
	}

	testUtils.ExecuteQueryTestCase(t, test)
}

func TestSchemaAggregateTopLevelCreatesAverageGivenSchema(t *testing.T) {
	test := testUtils.QueryTestCase{
		Schema: []string{
			`
				type users {}
			`,
		},
		IntrospectionQuery: `
			query IntrospectionQuery {
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
		ContainsData: map[string]interface{}{
			"__schema": map[string]interface{}{
				"queryType": map[string]interface{}{
					"fields": []interface{}{
						map[string]interface{}{
							"name": "_avg",
							"args": []interface{}{
								map[string]interface{}{
									"name": "users",
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
		},
	}

	testUtils.ExecuteQueryTestCase(t, test)
}
