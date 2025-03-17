// Copyright 2025 Democratized Data Foundation
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

func TestSchemaInstrospection_SimilarityCapableFieldIntArray(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						someVector: [Int!]
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
												kind
												ofType {
													name
													kind
													ofType {
														name
														kind
														ofType {
															name
															kind
														}
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
								"name": "_similarity",
								"args": []any{
									map[string]any{
										"name": "someVector",
										"type": map[string]any{
											"inputFields": []any{
												map[string]any{
													"name": "vector",
													"type": map[string]any{
														"kind": "NON_NULL",
														"name": nil,
														"ofType": map[string]any{
															"kind": "LIST",
															"name": nil,
															"ofType": map[string]any{
																"name": nil,
																"kind": "NON_NULL",
																"ofType": map[string]any{
																	"name": "Int",
																	"kind": "SCALAR",
																},
															},
														},
													},
												},
											},
											"name": "Users__someVector__SimilaritySelector",
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

func TestSchemaInstrospection_SimilarityCapableFieldFloat32Array(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						someVector: [Float32!]
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
												kind
												ofType {
													name
													kind
													ofType {
														name
														kind
														ofType {
															name
															kind
														}
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
								"name": "_similarity",
								"args": []any{
									map[string]any{
										"name": "someVector",
										"type": map[string]any{
											"inputFields": []any{
												map[string]any{
													"name": "vector",
													"type": map[string]any{
														"kind": "NON_NULL",
														"name": nil,
														"ofType": map[string]any{
															"kind": "LIST",
															"name": nil,
															"ofType": map[string]any{
																"name": nil,
																"kind": "NON_NULL",
																"ofType": map[string]any{
																	"name": "Float32",
																	"kind": "SCALAR",
																},
															},
														},
													},
												},
											},
											"name": "Users__someVector__SimilaritySelector",
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

func TestSchemaInstrospection_SimilarityCapableFieldsIntArrayAndFloat32Array(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						someVectorInt: [Int!]
						someVectorFloat32: [Float32!]
						someOtherNumber: Int
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
												kind
												ofType {
													name
													kind
													ofType {
														name
														kind
														ofType {
															name
															kind
														}
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
								"name": "_similarity",
								"args": []any{
									map[string]any{
										"name": "someVectorFloat32",
										"type": map[string]any{
											"inputFields": []any{
												map[string]any{
													"name": "vector",
													"type": map[string]any{
														"kind": "NON_NULL",
														"name": nil,
														"ofType": map[string]any{
															"kind": "LIST",
															"name": nil,
															"ofType": map[string]any{
																"name": nil,
																"kind": "NON_NULL",
																"ofType": map[string]any{
																	"name": "Float32",
																	"kind": "SCALAR",
																},
															},
														},
													},
												},
											},
											"name": "Users__someVectorFloat32__SimilaritySelector",
										},
									},
									map[string]any{
										"name": "someVectorInt",
										"type": map[string]any{
											"inputFields": []any{
												map[string]any{
													"name": "vector",
													"type": map[string]any{
														"kind": "NON_NULL",
														"name": nil,
														"ofType": map[string]any{
															"kind": "LIST",
															"name": nil,
															"ofType": map[string]any{
																"name": nil,
																"kind": "NON_NULL",
																"ofType": map[string]any{
																	"name": "Int",
																	"kind": "SCALAR",
																},
															},
														},
													},
												},
											},
											"name": "Users__someVectorInt__SimilaritySelector",
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
