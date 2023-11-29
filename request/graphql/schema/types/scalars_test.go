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

func TestBytesScalarTypeSerialize(t *testing.T) {
	input := []byte{0, 255}
	output := "00ff"

	cases := []struct {
		input  any
		expect any
	}{
		{input, output},
		{&input, output},
		{nil, nil},
		{0, nil},
		{false, nil},
	}
	for _, c := range cases {
		result := BytesScalarType.Serialize(c.input)
		assert.Equal(t, c.expect, result)
	}
}

func TestBytesScalarTypeParseValue(t *testing.T) {
	input := "00ff"
	output := []byte{0, 255}
	invalid := "invalid"

	cases := []struct {
		input  any
		expect any
	}{
		{input, output},
		{&input, output},
		{invalid, nil},
		{&invalid, nil},
		{nil, nil},
		{0, nil},
		{false, nil},
	}
	for _, c := range cases {
		result := BytesScalarType.ParseValue(c.input)
		assert.Equal(t, c.expect, result)
	}
}

func TestBytesScalarTypeParseLiteral(t *testing.T) {
	cases := []struct {
		input  ast.Value
		expect any
	}{
		{&ast.StringValue{Value: "00ff"}, []byte{0, 255}},
		{&ast.StringValue{Value: "invalid"}, nil},
		{&ast.IntValue{}, nil},
		{&ast.BooleanValue{}, nil},
		{&ast.NullValue{}, nil},
		{&ast.EnumValue{}, nil},
		{&ast.FloatValue{}, nil},
		{&ast.ListValue{}, nil},
		{&ast.ObjectValue{}, nil},
	}
	for _, c := range cases {
		result := BytesScalarType.ParseLiteral(c.input)
		assert.Equal(t, c.expect, result)
	}
}
