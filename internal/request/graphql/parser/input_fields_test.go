// Copyright 2026 Democratized Data Foundation
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

	gqlp "github.com/sourcenetwork/graphql-go/language/parser"
	gqls "github.com/sourcenetwork/graphql-go/language/source"
	"github.com/stretchr/testify/require"
)

func TestValidateUniqueInputFieldNames(t *testing.T) {
	tests := []struct {
		name        string
		request     string
		expectedLen int
	}{
		{
			name:        "no duplicates",
			request:     `query { User(filter: {_or: [{name: {_eq: "a"}}, {age: {_eq: 1}}]}) { name } }`,
			expectedLen: 0,
		},
		{
			name:        "duplicate at top level of an input object",
			request:     `query { User(filter: {name: {_eq: "a"}, name: {_eq: "b"}}) { name } }`,
			expectedLen: 1,
		},
		{
			name:        "duplicate nested inside another field's value",
			request:     `query { _commits(filter: {_or: {fieldName: {_eq: "title"}, fieldName: {_eq: "plot"}}}) { cid } }`,
			expectedLen: 1,
		},
		{
			name:        "duplicate nested inside a list element",
			request:     `query { User(filter: {_or: [{name: {_eq: "a"}, name: {_eq: "b"}}]}) { name } }`,
			expectedLen: 1,
		},
		{
			name:        "duplicate in a variable default value",
			request:     `query ($f: Foo = {name: {_eq: "a"}, name: {_eq: "b"}}) { User(filter: $f) { name } }`,
			expectedLen: 1,
		},
		{
			name:        "multiple duplicates reported separately",
			request:     `query { User(filter: {_and: [{x: 1, x: 2}, {y: 1, y: 2}]}) { name } }`,
			expectedLen: 2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			src := gqls.NewSource(&gqls.Source{Body: []byte(test.request)})
			doc, err := gqlp.Parse(gqlp.ParseParams{Source: src})
			require.NoError(t, err)

			errs := ValidateUniqueInputFieldNames(doc)
			require.Len(t, errs, test.expectedLen)
		})
	}
}
