// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package schema

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestFilterForSimpleSchema(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
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
									"name": "Users",
									"args": append(
										defaultUserArgsWithoutFilter,
										map[string]any{
											"name": "filter",
											"type": map[string]any{
												"name": "UsersFilterArg",
												"inputFields": []any{
													map[string]any{
														"name": "_and",
														"type": map[string]any{
															"name": nil,
															"ofType": map[string]any{
																"name": "UsersFilterArg",
															},
														},
													},
													map[string]any{
														"name": "_docID",
														"type": map[string]any{
															"name":   "IDOperatorBlock",
															"ofType": nil,
														},
													},
													map[string]any{
														"name": "_not",
														"type": map[string]any{
															"name":   "UsersFilterArg",
															"ofType": nil,
														},
													},
													map[string]any{
														"name": "_or",
														"type": map[string]any{
															"name": nil,
															"ofType": map[string]any{
																"name": "UsersFilterArg",
															},
														},
													},
													map[string]any{
														"name": "name",
														"type": map[string]any{
															"name":   "StringOperatorBlock",
															"ofType": nil,
														},
													},
												},
											},
										},
									).Tidy(),
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

var testFilterForSimpleSchemaArgProps = map[string]any{
	"name": struct{}{},
	"type": map[string]any{
		"name":        struct{}{},
		"inputFields": struct{}{},
	},
}

var defaultUserArgsWithoutFilter = trimFields(
	fields{
		cidArg,
		docIDArg,
		showDeletedArg,
		groupByArg,
		limitArg,
		offsetArg,
		buildOrderArg("Users"),
	},
	testFilterForSimpleSchemaArgProps,
)

func TestFilterForOneToOneSchema(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Book {
						name: String
						author: Author
					}

					type Author {
						age: Int
						wrote: Book @primary
					}
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
									"name": "Book",
									"args": append(
										defaultBookArgsWithoutFilter,
										map[string]any{
											"name": "filter",
											"type": map[string]any{
												"name": "BookFilterArg",
												"inputFields": []any{
													map[string]any{
														"name": "_and",
														"type": map[string]any{
															"name": nil,
															"ofType": map[string]any{
																"name": "BookFilterArg",
															},
														},
													},
													map[string]any{
														"name": "_docID",
														"type": map[string]any{
															"name":   "IDOperatorBlock",
															"ofType": nil,
														},
													},
													map[string]any{
														"name": "_not",
														"type": map[string]any{
															"name":   "BookFilterArg",
															"ofType": nil,
														},
													},
													map[string]any{
														"name": "_or",
														"type": map[string]any{
															"name": nil,
															"ofType": map[string]any{
																"name": "BookFilterArg",
															},
														},
													},
													map[string]any{
														"name": "author",
														"type": map[string]any{
															"name":   "AuthorFilterArg",
															"ofType": nil,
														},
													},
													map[string]any{
														"name": "author_id",
														"type": map[string]any{
															"name":   "IDOperatorBlock",
															"ofType": nil,
														},
													},
													map[string]any{
														"name": "name",
														"type": map[string]any{
															"name":   "StringOperatorBlock",
															"ofType": nil,
														},
													},
												},
											},
										},
									).Tidy(),
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

var testFilterForOneToOneSchemaArgProps = map[string]any{
	"name": struct{}{},
	"type": map[string]any{
		"name":        struct{}{},
		"inputFields": struct{}{},
	},
}

var defaultBookArgsWithoutFilter = trimFields(
	fields{
		cidArg,
		docIDArg,
		showDeletedArg,
		groupByArg,
		limitArg,
		offsetArg,
		buildOrderArg("Book"),
	},
	testFilterForOneToOneSchemaArgProps,
)
