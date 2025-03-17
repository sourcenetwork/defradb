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
	"math"
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

func boolPtr(b bool) *bool {
	return &b
}

func intPtr(i int) *int {
	return &i
}

func int8Ptr(i int8) *int8 {
	return &i
}
func int16Ptr(i int16) *int16 {
	return &i
}
func int32Ptr(i int32) *int32 {
	return &i
}
func int64Ptr(i int64) *int64 {
	return &i
}
func uint8Ptr(i uint8) *uint8 {
	return &i
}
func uint16Ptr(i uint16) *uint16 {
	return &i
}
func uint32Ptr(i uint32) *uint32 {
	return &i
}
func uint64Ptr(i uint64) *uint64 {
	return &i
}
func uintPtr(i uint) *uint {
	return &i
}
func float32Ptr(i float32) *float32 {
	return &i
}
func float64Ptr(i float64) *float64 {
	return &i
}
func stringPtr(i string) *string {
	return &i
}

func TestCoerceFloat32(t *testing.T) {
	tests := []struct {
		in   any
		want any
	}{
		{
			in:   false,
			want: float32(0.0),
		},
		{
			in:   true,
			want: float32(1.0),
		},
		{
			in:   boolPtr(false),
			want: float32(0.0),
		},
		{
			in:   boolPtr(true),
			want: float32(1.0),
		},
		{
			in:   (*bool)(nil),
			want: nil,
		},
		{
			in:   int(math.MinInt32),
			want: float32(math.MinInt32),
		},
		{
			in:   int(math.MaxInt32),
			want: float32(math.MaxInt32),
		},
		{
			in:   intPtr(12),
			want: float32(12),
		},
		{
			in:   (*int)(nil),
			want: nil,
		},
		{
			in:   int8(13),
			want: float32(13),
		},
		{
			in:   int8Ptr(14),
			want: float32(14),
		},
		{
			in:   (*int8)(nil),
			want: nil,
		},
		{
			in:   int16(15),
			want: float32(15),
		},
		{
			in:   int16Ptr(16),
			want: float32(16),
		},
		{
			in:   (*int16)(nil),
			want: nil,
		},
		{
			in:   int32(17),
			want: float32(17),
		},
		{
			in:   int32Ptr(18),
			want: float32(18),
		},
		{
			in:   (*int32)(nil),
			want: nil,
		},
		{
			in:   int64(19),
			want: float32(19),
		},
		{
			in:   int64Ptr(20),
			want: float32(20),
		},
		{
			in:   (*int64)(nil),
			want: nil,
		},
		{
			in:   uint8(21),
			want: float32(21),
		},
		{
			in:   uint8Ptr(22),
			want: float32(22),
		},
		{
			in:   (*uint8)(nil),
			want: nil,
		},
		{
			in:   uint16(23),
			want: float32(23),
		},
		{
			in:   uint16Ptr(24),
			want: float32(24),
		},
		{
			in:   (*uint16)(nil),
			want: nil,
		},
		{
			in:   uint32(25),
			want: float32(25),
		},
		{
			in:   uint32Ptr(26),
			want: float32(26),
		},
		{
			in:   (*uint32)(nil),
			want: nil,
		},
		{
			in:   uint64(27),
			want: float32(27),
		},
		{
			in:   uint64Ptr(28),
			want: float32(28),
		},
		{
			in:   (*uint64)(nil),
			want: nil,
		},
		{
			in:   uintPtr(29),
			want: float32(29),
		},
		{
			in:   (*uint)(nil),
			want: nil,
		},
		{
			in:   float32(30),
			want: float32(30),
		},
		{
			in:   float32Ptr(31),
			want: float32(31),
		},
		{
			in:   (*float32)(nil),
			want: nil,
		},
		{
			in:   float32(32),
			want: float32(32),
		},
		{
			in:   float64Ptr(33.2),
			want: float32(33.2),
		},
		{
			in:   (*float64)(nil),
			want: nil,
		},
		{
			in:   "34",
			want: float32(34),
		},
		{
			in:   stringPtr("35.2"),
			want: float32(35.2),
		},
		{
			in:   (*string)(nil),
			want: nil,
		},
		{
			in:   "I'm not a number",
			want: nil,
		},
		{
			in:   make(map[string]any),
			want: nil,
		},
	}

	for i, tt := range tests {
		if got, want := coerceFloat32(tt.in), tt.want; got != want {
			t.Errorf("%d: in=%v, got=%v, want=%v", i, tt.in, got, want)
		}
	}
}

func TestCoerceFloat64(t *testing.T) {
	tests := []struct {
		in   any
		want any
	}{
		{
			in:   false,
			want: float64(0.0),
		},
		{
			in:   true,
			want: float64(1.0),
		},
		{
			in:   boolPtr(false),
			want: float64(0.0),
		},
		{
			in:   boolPtr(true),
			want: float64(1.0),
		},
		{
			in:   (*bool)(nil),
			want: nil,
		},
		{
			in:   int(math.MinInt32),
			want: float64(math.MinInt32),
		},
		{
			in:   int(math.MaxInt32),
			want: float64(math.MaxInt32),
		},
		{
			in:   intPtr(12),
			want: float64(12),
		},
		{
			in:   (*int)(nil),
			want: nil,
		},
		{
			in:   int8(13),
			want: float64(13),
		},
		{
			in:   int8Ptr(14),
			want: float64(14),
		},
		{
			in:   (*int8)(nil),
			want: nil,
		},
		{
			in:   int16(15),
			want: float64(15),
		},
		{
			in:   int16Ptr(16),
			want: float64(16),
		},
		{
			in:   (*int16)(nil),
			want: nil,
		},
		{
			in:   int32(17),
			want: float64(17),
		},
		{
			in:   int32Ptr(18),
			want: float64(18),
		},
		{
			in:   (*int32)(nil),
			want: nil,
		},
		{
			in:   int64(19),
			want: float64(19),
		},
		{
			in:   int64Ptr(20),
			want: float64(20),
		},
		{
			in:   (*int64)(nil),
			want: nil,
		},
		{
			in:   uint8(21),
			want: float64(21),
		},
		{
			in:   uint8Ptr(22),
			want: float64(22),
		},
		{
			in:   (*uint8)(nil),
			want: nil,
		},
		{
			in:   uint16(23),
			want: float64(23),
		},
		{
			in:   uint16Ptr(24),
			want: float64(24),
		},
		{
			in:   (*uint16)(nil),
			want: nil,
		},
		{
			in:   uint32(25),
			want: float64(25),
		},
		{
			in:   uint32Ptr(26),
			want: float64(26),
		},
		{
			in:   (*uint32)(nil),
			want: nil,
		},
		{
			in:   uint64(27),
			want: float64(27),
		},
		{
			in:   uint64Ptr(28),
			want: float64(28),
		},
		{
			in:   (*uint64)(nil),
			want: nil,
		},
		{
			in:   uintPtr(29),
			want: float64(29),
		},
		{
			in:   (*uint)(nil),
			want: nil,
		},
		{
			in:   float32(30),
			want: float64(30),
		},
		{
			in:   float32Ptr(31),
			want: float64(31),
		},
		{
			in:   (*float32)(nil),
			want: nil,
		},
		{
			in:   float32(32),
			want: float64(32),
		},
		{
			in:   float64Ptr(33.2),
			want: float64(33.2),
		},
		{
			in:   (*float64)(nil),
			want: nil,
		},
		{
			in:   "34",
			want: float64(34),
		},
		{
			in:   stringPtr("35.2"),
			want: float64(35.2),
		},
		{
			in:   (*string)(nil),
			want: nil,
		},
		{
			in:   "I'm not a number",
			want: nil,
		},
		{
			in:   make(map[string]any),
			want: nil,
		},
	}

	for i, tt := range tests {
		if got, want := coerceFloat64(tt.in), tt.want; got != want {
			t.Errorf("%d: in=%v, got=%v, want=%v", i, tt.in, got, want)
		}
	}
}
