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
		result := BlobScalarType.ParseValue(c.input)
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
		result := BlobScalarType.ParseLiteral(c.input)
		assert.Equal(t, c.expect, result)
	}
}

func TestBigIntScalarTypeParseValue(t *testing.T) {
	stringInput := "123456"
	intInput := int(123456)
	int16Input := int16(12345)
	int32Input := int32(123456)
	int64Input := int64(123456)
	uintInput := uint(123456)
	uint16Input := uint16(12345)
	uint32Input := uint32(123456)
	uint64Input := uint64(123456)
	float32Input := float32(123456)
	float64Input := float64(123456)
	// invalid string containing non-number characters
	invalidString := "!@#$%^&*"

	cases := []struct {
		input  any
		expect any
	}{
		{stringInput, "123456"},
		{&stringInput, "123456"},
		{intInput, "123456"},
		{&intInput, "123456"},
		{int16Input, "12345"},
		{&int16Input, "12345"},
		{int32Input, "123456"},
		{&int32Input, "123456"},
		{int64Input, "123456"},
		{&int64Input, "123456"},
		{uintInput, "123456"},
		{&uintInput, "123456"},
		{uint16Input, "12345"},
		{&uint16Input, "12345"},
		{uint32Input, "123456"},
		{&uint32Input, "123456"},
		{uint64Input, "123456"},
		{&uint64Input, "123456"},
		{float32Input, "123456"},
		{&float32Input, "123456"},
		{float64Input, "123456"},
		{&float64Input, "123456"},
		{invalidString, nil},
		{&invalidString, nil},
		{nil, nil},
		{false, nil},
	}
	for _, c := range cases {
		result := BigIntScalarType.ParseValue(c.input)
		assert.Equal(t, c.expect, result)
	}
}

func TestBigIntScalarTypeParseLiteral(t *testing.T) {
	cases := []struct {
		input  ast.Value
		expect any
	}{
		{&ast.StringValue{Value: "123456"}, "123456"},
		{&ast.StringValue{Value: "00!@#$%^&*"}, nil},
		{&ast.StringValue{Value: "!@#$%^&*00"}, nil},
		{&ast.IntValue{Value: "123456"}, "123456"},
		{&ast.BooleanValue{}, nil},
		{&ast.NullValue{}, nil},
		{&ast.EnumValue{}, nil},
		{&ast.FloatValue{Value: "123456"}, "123456"},
		{&ast.ListValue{}, nil},
		{&ast.ObjectValue{}, nil},
	}
	for _, c := range cases {
		result := BigIntScalarType.ParseLiteral(c.input)
		assert.Equal(t, c.expect, result)
	}
}

func TestBigFloatScalarTypeParseValue(t *testing.T) {
	stringInput := "123456.123456"
	exponentInput := "123456.123456e±12"
	intInput := int(123456)
	int16Input := int16(12345)
	int32Input := int32(123456)
	int64Input := int64(123456)
	uintInput := uint(123456)
	uint16Input := uint16(12345)
	uint32Input := uint32(123456)
	uint64Input := uint64(123456)
	float32Input := float32(123456.123456)
	float64Input := float64(123456.123456)
	// invalid string containing non-number characters
	invalidString := "!@#$%^&*"

	cases := []struct {
		input  any
		expect any
	}{
		{stringInput, "123456.123456"},
		{&stringInput, "123456.123456"},
		{exponentInput, "123456.123456e±12"},
		{&exponentInput, "123456.123456e±12"},
		{intInput, "123456"},
		{&intInput, "123456"},
		{int16Input, "12345"},
		{&int16Input, "12345"},
		{int32Input, "123456"},
		{&int32Input, "123456"},
		{int64Input, "123456"},
		{&int64Input, "123456"},
		{uintInput, "123456"},
		{&uintInput, "123456"},
		{uint16Input, "12345"},
		{&uint16Input, "12345"},
		{uint32Input, "123456"},
		{&uint32Input, "123456"},
		{uint64Input, "123456"},
		{&uint64Input, "123456"},
		{float32Input, "123456.125"},
		{&float32Input, "123456.125"},
		{float64Input, "123456.1235"},
		{&float64Input, "123456.1235"},
		{invalidString, nil},
		{&invalidString, nil},
		{nil, nil},
		{false, nil},
	}
	for _, c := range cases {
		result := BigFloatScalarType.ParseValue(c.input)
		assert.Equal(t, c.expect, result)
	}
}

func TestBigFloatScalarTypeParseLiteral(t *testing.T) {
	cases := []struct {
		input  ast.Value
		expect any
	}{
		{&ast.StringValue{Value: "123456.123456"}, "123456.123456"},
		{&ast.StringValue{Value: "123456.123456e±12"}, "123456.123456e±12"},
		{&ast.StringValue{Value: "00!@#$%^&*"}, nil},
		{&ast.StringValue{Value: "!@#$%^&*00"}, nil},
		{&ast.IntValue{Value: "123456"}, "123456"},
		{&ast.BooleanValue{}, nil},
		{&ast.NullValue{}, nil},
		{&ast.EnumValue{}, nil},
		{&ast.FloatValue{Value: "123456.123456"}, "123456.123456"},
		{&ast.ListValue{}, nil},
		{&ast.ObjectValue{}, nil},
	}
	for _, c := range cases {
		result := BigFloatScalarType.ParseLiteral(c.input)
		assert.Equal(t, c.expect, result)
	}
}
