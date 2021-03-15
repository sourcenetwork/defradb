// Copyright 2020 Source Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package parser

import (
	"testing"

	"github.com/graphql-go/graphql/language/ast"
	gqlp "github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
	"github.com/stretchr/testify/assert"
)

func getQueryFilterObject(query string) (*Filter, error) {
	source := source.NewSource(&source.Source{
		Body: []byte(query),
		Name: "",
	})

	doc, err := gqlp.Parse(gqlp.ParseParams{Source: source})
	if err != nil {
		return nil, err
	}

	filterObj := doc.Definitions[0].(*ast.OperationDefinition).SelectionSet.Selections[0].(*ast.Field).Arguments[0].Value.(*ast.ObjectValue)
	return NewFilter(filterObj)
}

func TestNewFilterFromString(t *testing.T) {
	_, err := NewFilterFromString(`name: {_eq: "bob"}`)
	assert.NoError(t, err)

	// fmt.Println(f.Conditions)
	// assert.True(t, false)
}

func TestParseConditions_Empty(t *testing.T) {
	var query = (`
	query {
		users(filter: {})
	}`)

	filter, err := getQueryFilterObject(query)
	assert.NoError(t, err)

	assert.Equal(t, filter.Conditions, map[string]interface{}{})
}

func TestParseConditions_Single_Int(t *testing.T) {
	runParseConditionTest(t, (`
		query {
			users(filter: {x: 1})
		}`),
		map[string]interface{}{
			"x": int64(1),
		},
	)
}

func TestParseConditions_Single_Float(t *testing.T) {
	runParseConditionTest(t, (`
		query {
			users(filter: {x: 1.1})
		}`),
		map[string]interface{}{
			"x": 1.1,
		},
	)
}

func TestParseConditions_Single_String(t *testing.T) {
	runParseConditionTest(t, (`
		query {
			users(filter: {x: "test"})
		}`),
		map[string]interface{}{
			"x": "test",
		},
	)
}

func TestParseConditions_Single_Enum(t *testing.T) {
	runParseConditionTest(t, (`
		query {
			users(filter: {x: ASC})
		}`),
		map[string]interface{}{
			"x": "ASC",
		},
	)
}

func TestParseConditions_Single_Boolean(t *testing.T) {
	runParseConditionTest(t, (`
		query {
			users(filter: {x: true})
		}`),
		map[string]interface{}{
			"x": true,
		},
	)
}

// Ignoring Null input, see:
// - https://github.com/graphql-go/graphql/issues/178
// - https://github.com/99designs/gqlgen/issues/1416
// func TestParseConditions_Single_Null(t *testing.T) {
// 	runParseConditionTest(t, (`
// 		query {
// 			users(filter: {x: null})
// 		}`),
// 		map[string]interface{}{
// 			"x": nil,
// 		},
// 	)
// }

func TestParseConditions_List_Int(t *testing.T) {
	runParseConditionTest(t, (`
		query {
			users(filter: {x: [1,2,3]})
		}`),
		map[string]interface{}{
			"x": []interface{}{int64(1), int64(2), int64(3)},
		},
	)
}

func TestParseConditions_List_Float(t *testing.T) {
	runParseConditionTest(t, (`
		query {
			users(filter: {x: [1.1,2.2,3.3]})
		}`),
		map[string]interface{}{
			"x": []interface{}{1.1, 2.2, 3.3},
		},
	)
}

func TestParseConditions_List_String(t *testing.T) {
	runParseConditionTest(t, (`
		query {
			users(filter: {x: ["hello", "world", "bye"]})
		}`),
		map[string]interface{}{
			"x": []interface{}{"hello", "world", "bye"},
		},
	)
}

func TestParseConditions_Single_Object_Simple(t *testing.T) {
	runParseConditionTest(t, (`
		query {
			users(filter: {x: {y: 1}})
		}`),
		map[string]interface{}{
			"x": map[string]interface{}{
				"y": int64(1),
			},
		},
	)
}

func TestParseConditions_Single_Object_Multiple(t *testing.T) {
	runParseConditionTest(t, (`
		query {
			users(filter: {x: {y: 1, z: 2}})
		}`),
		map[string]interface{}{
			"x": map[string]interface{}{
				"y": int64(1),
				"z": int64(2),
			},
		},
	)
}

func TestParseConditions_Single_Object_Nested(t *testing.T) {
	runParseConditionTest(t, (`
		query {
			users(filter: {x: {y: 1, z: {a: 2}}})
		}`),
		map[string]interface{}{
			"x": map[string]interface{}{
				"y": int64(1),
				"z": map[string]interface{}{
					"a": int64(2),
				},
			},
		},
	)
}

func TestParseConditions_Single_List_Objects(t *testing.T) {
	runParseConditionTest(t, (`
		query {
			users(filter: {_and: [{x: 1}, {y: 1}]})
		}`),
		map[string]interface{}{
			"$and": []interface{}{
				map[string]interface{}{
					"x": int64(1),
				},
				map[string]interface{}{
					"y": int64(1),
				},
			},
		},
	)
}

func runParseConditionTest(t *testing.T, query string, target map[string]interface{}) {
	filter, err := getQueryFilterObject(query)
	assert.NoError(t, err)

	assert.Equal(t, target, filter.Conditions)
}

// func TestRunFilter_Single_Int_Match(t *testing.T) {
// 	runRunFilterTest(t, (`
// 		query {
// 			users(filter: {x: 1})
// 		}`),
// 		map[string]interface{}{
// 			"x": int64(1),
// 		},
// 		true,
// 	)
// }

// func TestRunFilter_Single_Int_FilterNoMatch(t *testing.T) {
// 	runRunFilterTest(t, (`
// 		query {
// 			users(filter: {x: 2})
// 		}`),
// 		map[string]interface{}{
// 			"x": int64(1),
// 		},
// 		false,
// 	)
// }

// func TestRunFilter_Single_Int_DataNoMatch(t *testing.T) {
// 	runRunFilterTest(t, (`
// 		query {
// 			users(filter: {x: 1})
// 		}`),
// 		map[string]interface{}{
// 			"x": int64(2),
// 		},
// 		false,
// 	)
// }

// func TestRunFilter_Single_Int_Match_MultiData(t *testing.T) {
// 	runRunFilterTest(t, (`
// 		query {
// 			users(filter: {x: 2})
// 		}`),
// 		map[string]interface{}{
// 			"x": int64(1),
// 			"y": "hello",
// 		},
// 		false,
// 	)
// }

// func TestRunFilter_Single_String_Match(t *testing.T) {
// 	runRunFilterTest(t, (`
// 		query {
// 			users(filter: {x: "test"})
// 		}`),
// 		map[string]interface{}{
// 			"x": "test",
// 		},
// 		true,
// 	)
// }

// func TestRunFilter_Single_String_NoMatch(t *testing.T) {
// 	runRunFilterTest(t, (`
// 		query {
// 			users(filter: {x: "test"})
// 		}`),
// 		map[string]interface{}{
// 			"x": "somethingelse",
// 		},
// 		false,
// 	)
// }

// func TestRunFilter_Single_Float_Match(t *testing.T) {
// 	runRunFilterTest(t, (`
// 		query {
// 			users(filter: {x: 1.1})
// 		}`),
// 		map[string]interface{}{
// 			"x": 1.1,
// 		},
// 		true,
// 	)
// }

// func TestRunFilter_Single_Float_NoMatch(t *testing.T) {
// 	runRunFilterTest(t, (`
// 		query {
// 			users(filter: {x: 1.1})
// 		}`),
// 		map[string]interface{}{
// 			"x": 2.2,
// 		},
// 		false,
// 	)
// }

type testCase struct {
	description string
	query       string
	data        map[string]interface{}
	shouldPass  bool
}

func TestRunFilter_TestCases(t *testing.T) {
	testCases := []testCase{
		{
			"Single Int, match",
			(`
			query {
				users(filter: {x: 1})
			}`),
			map[string]interface{}{
				"x": uint64(1),
			},
			true,
		},
		{
			"Single Int, no match",
			(`
			query {
				users(filter: {x: 1})
			}`),
			map[string]interface{}{
				"x": int64(2),
			},
			false,
		},
		{
			"Single Int, match, multiple data fields",
			(`
			query {
				users(filter: {x: 1})
			}`),
			map[string]interface{}{
				"x": int64(1),
				"y": int64(2),
			},
			true,
		},
		{
			"Single string, match",
			(`
			query {
				users(filter: {x: "test"})
			}`),
			map[string]interface{}{
				"x": "test",
			},
			true,
		},
		{
			"Single string, no match",
			(`
			query {
				users(filter: {x: "test"})
			}`),
			map[string]interface{}{
				"x": "nothing",
			},
			false,
		},
		{
			"Single float, match",
			(`
			query {
				users(filter: {x: 1.1})
			}`),
			map[string]interface{}{
				"x": 1.1,
			},
			true,
		},
		{
			"Single float, no match",
			(`
			query {
				users(filter: {x: 1.1})
			}`),
			map[string]interface{}{
				"x": 2.2,
			},
			false,
		},
		{
			"Single boolean, match",
			(`
			query {
				users(filter: {x: true})
			}`),
			map[string]interface{}{
				"x": true,
			},
			true,
		},
		// @todo: implement list equality filter
		// {
		// 	"Single list, match",
		// 	(`
		// 	query {
		// 		users(filter: {x: [1,2,3]})
		// 	}`),
		// 	map[string]interface{}{
		// 		"x": []interface{}{
		// 			int64(1),
		// 			int64(2),
		// 			int64(3),
		// 		},
		// 	},
		// 	false,
		// },
		{
			"Multi condition match",
			(`
			query {
				users(filter: {x: 1, y: 1})
			}`),
			map[string]interface{}{
				"x": int64(1),
				"y": int64(1),
			},
			true,
		},
		{
			"Object, match",
			(`
			query {
				users(filter: {x: {y: 1}})
			}`),
			map[string]interface{}{
				"x": map[string]interface{}{
					"y": int64(1),
				},
			},
			true,
		},
		{
			"Nested object, match",
			(`
			query {
				users(filter: {x: {y: {z: 1}}})
			}`),
			map[string]interface{}{
				"x": map[string]interface{}{
					"y": map[string]interface{}{
						"z": int64(1),
					},
				},
			},
			true,
		},
		{
			"Nested object, multi match",
			(`
			query {
				users(filter: {a: 1, x: {y: {z: 1}}})
			}`),
			map[string]interface{}{
				"x": map[string]interface{}{
					"y": map[string]interface{}{
						"z": int64(1),
					},
				},
				"a": int64(1),
			},
			true,
		},
		{
			"Explicit condition match: Int _eq",
			(`
			query {
				users(filter: {x: {_eq: 1}})
			}`),
			map[string]interface{}{
				"x": int64(1),
			},
			true,
		},
		{
			"Explicit condition match: Int _ne",
			(`
			query {
				users(filter: {x: {_ne: 1}})
			}`),
			map[string]interface{}{
				"x": int64(2),
			},
			true,
		},
		{
			"Explicit condition match: Float _ne",
			(`
			query {
				users(filter: {x: {_ne: 1.1}})
			}`),
			map[string]interface{}{
				"x": int64(2),
			},
			true,
		},
		{
			"Explicit condition match: String _ne",
			(`
			query {
				users(filter: {x: {_ne: "test"}})
			}`),
			map[string]interface{}{
				"x": int64(2),
			},
			true,
		},
		{
			"Explicit condition match: Int greater than (_gt)",
			(`
			query {
				users(filter: {x: {_gt: 1}})
			}`),
			map[string]interface{}{
				"x": int64(2),
			},
			true,
		},
		{
			"Explicit condition no match: Int greater than (_gt)",
			(`
			query {
				users(filter: {x: {_gt: 1}})
			}`),
			map[string]interface{}{
				"x": int64(0),
			},
			false,
		},
		{
			"Explicit muti condition match: Int greater than (_gt) less than (_lt)",
			(`
			query {
				users(filter: {x: {_gt: 1}, y: {_lt: 2}})
			}`),
			map[string]interface{}{
				"x": int64(3),
				"y": int64(1),
			},
			true,
		},
		{
			"Explicit condition match: Int & Float greater than (_gt)",
			(`
			query {
				users(filter: {x: {_gt: 1}})
			}`),
			map[string]interface{}{
				"x": 1.1,
			},
			true,
		},
		{
			"Explicit condition match: In set (_in)",
			(`
			query {
				users(filter: {x: {_in: [1,2,3]}})
			}`),
			map[string]interface{}{
				"x": int64(1),
			},
			true,
		},
		{
			"Explicit condition match: Not in set (_nin)",
			(`
			query {
				users(filter: {x: {_nin: [1,2,3]}})
			}`),
			map[string]interface{}{
				"x": int64(4),
			},
			true,
		},
		{
			"Compound logical condition match: _and",
			(`
			query {
				users(filter: {_and: [ {x: {_lt: 10}}, {x: {_gt: 5}} ]})
			}`),
			map[string]interface{}{
				"x": int64(6),
			},
			true,
		},
		{
			"Compound logical condition no match: _and",
			(`
			query {
				users(filter: {_and: [ {x: {_lt: 10}}, {x: {_gt: 5}} ]})
			}`),
			map[string]interface{}{
				"x": int64(11),
			},
			false,
		},
		{
			"Compound logical condition match: _or",
			(`
			query {
				users(filter: {_or: [ {x: 10}, {x: 5} ]})
			}`),
			map[string]interface{}{
				"x": int64(5),
			},
			true,
		},
		{
			"Compound nested logical condition match: _or #2",
			(`
			query {
				users(filter: {_or: 
								[ 
									{_and: [ {x: {_lt: 10}}, {x: {_gt: 5}} ]}, 
									{_and: [ {y: {_lt: 1}}, {y: {_gt: 0}} ]}, 
								]})
			}`),
			map[string]interface{}{
				"x": int64(6),
			},
			true,
		},
		{
			"Compound nested logical condition match: _or #2",
			(`
			query {
				users(filter: {_or: 
								[ 
									{_and: [ {x: {_lt: 10}}, {x: {_gt: 5}} ]}, 
									{_and: [ {y: {_lt: 1}}, {y: {_gt: 0}} ]}, 
								]})
			}`),
			map[string]interface{}{
				"y": 0.1,
			},
			true,
		},
		{
			"Nested object logical condition match: _or",
			(`
			query {
				users(filter: {_or: [ {x: 10}, {x: 5} ]})
			}`),
			map[string]interface{}{
				"x": int64(5),
			},
			true,
		},
	}
	for _, test := range testCases {
		runRunFilterTest(t, test)
	}
}

func runRunFilterTest(t *testing.T, test testCase) {
	filter, err := getQueryFilterObject(test.query)
	assert.NoError(t, err)

	passed, err := RunFilter(test.data, filter, EvalContext{})
	assert.NoError(t, err)
	assert.True(t, passed == test.shouldPass, "Test Case faild: %s", test.description)
}

func getQuerySortObject(query string) (*OrderBy, error) {
	source := source.NewSource(&source.Source{
		Body: []byte(query),
		Name: "",
	})

	doc, err := gqlp.Parse(gqlp.ParseParams{Source: source})
	if err != nil {
		return nil, err
	}

	sortObj := doc.Definitions[0].(*ast.OperationDefinition).SelectionSet.Selections[0].(*ast.Field).Arguments[0].Value.(*ast.ObjectValue)
	conditions, err := ParseConditionsInOrder(sortObj)
	if err != nil {
		return nil, err
	}
	return &OrderBy{
		Conditions: conditions,
		Statement:  sortObj,
	}, nil
}

func TestParseConditionsInOrder_Empty(t *testing.T) {
	runParseConditionInOrderTest(t, (`
		query {
			users(order: {})
		}`),
		[]SortCondition{},
	)
}

func TestParseConditionsInOrder_Simple(t *testing.T) {
	runParseConditionInOrderTest(t, (`
		query {
			users(order: {name: ASC})
		}`),
		[]SortCondition{
			{
				Field:     "name",
				Direction: ASC,
			},
		},
	)
}

func TestParseConditionsInOrder_Simple_Multiple(t *testing.T) {
	runParseConditionInOrderTest(t, (`
		query {
			users(order: {name: ASC, date: DESC})
		}`),
		[]SortCondition{
			{
				Field:     "name",
				Direction: ASC,
			},
			{
				Field:     "date",
				Direction: DESC,
			},
		},
	)
}

func TestParseConditionsInOrder_Embedded(t *testing.T) {
	runParseConditionInOrderTest(t, (`
		query {
			users(order: {author: {name: ASC}})
		}`),
		[]SortCondition{
			{
				Field:     "author.name",
				Direction: ASC,
			},
		},
	)
}

func TestParseConditionsInOrder_Embedded_Multiple(t *testing.T) {
	runParseConditionInOrderTest(t, (`
		query {
			users(order: {author: {name: ASC, birthday: DESC}})
		}`),
		[]SortCondition{
			{
				Field:     "author.name",
				Direction: ASC,
			},
			{
				Field:     "author.birthday",
				Direction: DESC,
			},
		},
	)
}

func TestParseConditionsInOrder_Embedded_Nested(t *testing.T) {
	runParseConditionInOrderTest(t, (`
		query {
			users(order: {author: {address: {street_name: DESC}}})
		}`),
		[]SortCondition{
			{
				Field:     "author.address.street_name",
				Direction: DESC,
			},
		},
	)
}

func TestParseConditionsInOrder_Complex(t *testing.T) {
	runParseConditionInOrderTest(t, (`
		query {
			users(order: {name: ASC, author: {birthday: ASC, address: {street_name: DESC}}})
		}`),
		[]SortCondition{
			{
				Field:     "name",
				Direction: ASC,
			},
			{
				Field:     "author.birthday",
				Direction: ASC,
			},
			{
				Field:     "author.address.street_name",
				Direction: DESC,
			},
		},
	)
}

func runParseConditionInOrderTest(t *testing.T, query string, target []SortCondition) {
	sort, err := getQuerySortObject(query)
	assert.NoError(t, err)

	assert.Equal(t, target, sort.Conditions)
}
