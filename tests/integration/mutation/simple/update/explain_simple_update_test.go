// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package update

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	simpleTests "github.com/sourcenetwork/defradb/tests/integration/mutation/simple"
)

type dataMap = map[string]interface{}

func TestExplainSimpleMutationUpdateWithBooleanFilter(t *testing.T) {
	tests := []testUtils.QueryTestCase{

		{
			Description: "Explain simple update mutation with boolean equals filter, multiple rows",

			Query: `mutation @explain {
						update_user(
							filter: {
								verified: {
									_eq: true
								}
							},
							data: "{\"points\": 59}"
						) {
							_key
							name
							points
						}
					}`,

			Docs: map[int][]string{
				0: {
					(`{
					"name": "John",
					"age": 27,
					"verified": true,
					"points": 42.1
				}`),
					(`{
					"name": "Bob",
					"age": 39,
					"verified": true,
					"points": 66.6
				}`)},
			},

			Results: []dataMap{
				{
					"explain": dataMap{
						"updateNode": dataMap{
							"data": dataMap{
								"points": float64(59),
							},
							"filter": dataMap{
								"verified": dataMap{
									"$eq": true,
								},
							},
							"ids": []string(nil),
							"selectTopNode": dataMap{
								"renderNode": dataMap{
									"selectNode": dataMap{
										"filter": nil,
										"scanNode": dataMap{
											"collectionID":   "1",
											"collectionName": "user",
											"filter":         nil,
											"spans":          []dataMap{},
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

	for _, test := range tests {
		simpleTests.ExecuteTestCase(t, test)
	}
}

func TestExplainSimpleMutationUpdateWithIdInFilter(t *testing.T) {
	test := testUtils.QueryTestCase{

		Description: "Explain simple update mutation with id in filter, multiple rows",

		Query: `mutation @explain {
					update_user(
						ids: [
							"bae-0a24cf29-b2c2-5861-9d00-abd6250c475d",
							"bae-958c9334-73cf-5695-bf06-cf06826babfa"
						],
						data: "{\"points\": 59}"
					) {
						_key
						name
						points
					}
				}`,

		Docs: map[int][]string{
			0: {
				(`{
				"name": "John",
				"age": 27,
				"verified": true,
				"points": 42.1
			}`),
				(`{
				"name": "Bob",
				"age": 39,
				"verified": false,
				"points": 66.6
			}`)},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"updateNode": dataMap{
						"data": dataMap{
							"points": float64(59),
						},
						"filter": nil,
						"ids": []string{
							"bae-0a24cf29-b2c2-5861-9d00-abd6250c475d",
							"bae-958c9334-73cf-5695-bf06-cf06826babfa",
						},
						"selectTopNode": dataMap{
							"renderNode": dataMap{
								"selectNode": dataMap{
									"filter": nil,
									"scanNode": dataMap{
										"collectionID":   "1",
										"collectionName": "user",
										"filter":         nil,
										"spans":          []dataMap{},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	simpleTests.ExecuteTestCase(t, test)
}

func TestExplainSimpleMutationUpdateWithIdEqualsFilter(t *testing.T) {
	test := testUtils.QueryTestCase{

		Description: "Explain simple update mutation with id equals filter, multiple rows but single match",

		Query: `mutation @explain {
					update_user(
						id: "bae-0a24cf29-b2c2-5861-9d00-abd6250c475d",
						data: "{\"points\": 59}"
					) {
						_key
						name
						points
					}
				}`,

		Docs: map[int][]string{
			0: {
				(`{
				"name": "John",
				"age": 27,
				"verified": true,
				"points": 42.1
			}`),
				(`{
				"name": "Bob",
				"age": 39,
				"verified": false,
				"points": 66.6
			}`)},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"updateNode": dataMap{
						"data": dataMap{
							"points": float64(59),
						},
						"filter": nil,
						"ids": []string{
							"bae-0a24cf29-b2c2-5861-9d00-abd6250c475d",
						},
						"selectTopNode": dataMap{
							"renderNode": dataMap{
								"selectNode": dataMap{
									"filter": nil,
									"scanNode": dataMap{
										"collectionID":   "1",
										"collectionName": "user",
										"filter":         nil,
										"spans":          []dataMap{},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	simpleTests.ExecuteTestCase(t, test)
}

func TestExplainSimpleMutationUpdateWithIdAndFilter(t *testing.T) {
	test := testUtils.QueryTestCase{

		Description: "Explain simple update mutation with ids and filter, multiple rows",

		Query: `mutation @explain {
					update_user(
						filter: {
							verified: {
								_eq: true
							}
						},
						ids: [
							"bae-0a24cf29-b2c2-5861-9d00-abd6250c475d",
							"bae-958c9334-73cf-5695-bf06-cf06826babfa"
						],
						data: "{\"points\": 59}"
					) {
						_key
						name
						points
					}
				}`,

		Docs: map[int][]string{
			0: {
				(`{
				"name": "John",
				"age": 27,
				"verified": true,
				"points": 42.1
			}`),
				(`{
				"name": "Bob",
				"age": 39,
				"verified": false,
				"points": 66.6
			}`)},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"updateNode": dataMap{
						"data": dataMap{
							"points": float64(59),
						},
						"filter": dataMap{
							"verified": dataMap{
								"$eq": true,
							},
						},
						"ids": []string{
							"bae-0a24cf29-b2c2-5861-9d00-abd6250c475d",
							"bae-958c9334-73cf-5695-bf06-cf06826babfa",
						},
						"selectTopNode": dataMap{
							"renderNode": dataMap{
								"selectNode": dataMap{
									"filter": nil,
									"scanNode": dataMap{
										"collectionID":   "1",
										"collectionName": "user",
										"filter":         nil,
										"spans":          []dataMap{},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	simpleTests.ExecuteTestCase(t, test)
}
