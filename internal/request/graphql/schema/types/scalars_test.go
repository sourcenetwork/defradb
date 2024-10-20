// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package types

import (
	"testing"

	"github.com/sourcenetwork/graphql-go/language/ast"
	"github.com/stretchr/testify/assert"
)

func TestBlobScalarTypeSerialize(t *testing.T) {
	stringInput := "00ff"
	bytesInput := []byte{0, 255}

	cases := []struct {
		input  any
		expect any
	}{
		{stringInput, "00ff"},
		{&stringInput, "00ff"},
		{bytesInput, "00ff"},
		{&bytesInput, "00ff"},
		{nil, nil},
		{0, nil},
		{false, nil},
	}
	for _, c := range cases {
		result := BlobScalarType().Serialize(c.input)
		assert.Equal(t, c.expect, result)
	}
}

func TestBlobScalarTypeParseValue(t *testing.T) {
	stringInput := "00ff"
	bytesInput := []byte{0, 255}
	// invalid string containing non-hex characters
	invalidHexString := "!@#$%^&*"

	cases := []struct {
		input  any
		expect any
	}{
		{stringInput, "00ff"},
		{&stringInput, "00ff"},
		{bytesInput, "00ff"},
		{&bytesInput, "00ff"},
		{invalidHexString, nil},
		{&invalidHexString, nil},
		{nil, nil},
		{0, nil},
		{false, nil},
	}
	for _, c := range cases {
		result := BlobScalarType().ParseValue(c.input)
		assert.Equal(t, c.expect, result)
	}
}

func TestBlobScalarTypeParseLiteral(t *testing.T) {
	cases := []struct {
		input  ast.Value
		expect any
	}{
		{&ast.StringValue{Value: "00ff"}, "00ff"},
		{&ast.StringValue{Value: "00!@#$%^&*"}, nil},
		{&ast.StringValue{Value: "!@#$%^&*00"}, nil},
		{&ast.IntValue{}, nil},
		{&ast.BooleanValue{}, nil},
		{&ast.NullValue{}, nil},
		{&ast.EnumValue{}, nil},
		{&ast.FloatValue{}, nil},
		{&ast.ListValue{}, nil},
		{&ast.ObjectValue{}, nil},
	}
	for _, c := range cases {
		result := BlobScalarType().ParseLiteral(c.input, nil)
		assert.Equal(t, c.expect, result)
	}
}

func TestJSONScalarTypeParseLiteral(t *testing.T) {
	cases := []struct {
		input  ast.Value
		expect any
	}{
		{&ast.StringValue{Value: "hello"}, "hello"},
		{&ast.IntValue{Value: "10"}, int32(10)},
		{&ast.BooleanValue{Value: true}, true},
		{&ast.NullValue{}, nil},
		{&ast.EnumValue{Value: "DESC"}, "DESC"},
		{&ast.Variable{Name: &ast.Name{Value: "message"}}, "hello"},
		{&ast.Variable{Name: &ast.Name{Value: "invalid"}}, nil},
		{&ast.FloatValue{Value: "3.14"}, 3.14},
		{&ast.ListValue{Values: []ast.Value{
			&ast.StringValue{Value: "hello"},
			&ast.IntValue{Value: "10"},
		}}, []any{"hello", int32(10)}},
		{&ast.ObjectValue{
			Fields: []*ast.ObjectField{
				{
					Name:  &ast.Name{Value: "int"},
					Value: &ast.IntValue{Value: "10"},
				},
				{
					Name:  &ast.Name{Value: "string"},
					Value: &ast.StringValue{Value: "hello"},
				},
			},
		}, map[string]any{
			"int":    int32(10),
			"string": "hello",
		}},
	}
	variables := map[string]any{
		"message": "hello",
	}
	for _, c := range cases {
		result := JSONScalarType().ParseLiteral(c.input, variables)
		assert.Equal(t, c.expect, result)
	}
}
