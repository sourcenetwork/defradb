// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/immutable"
)

var (
	testJSONObj = []byte(`{
		"Name": "John",
		"Age": 26
	}`)

	def = CollectionVersion{
		Name: "User",
		Fields: []CollectionFieldDescription{
			{
				Name: "Name",
				Typ:  LWW_REGISTER,
				Kind: FieldKind_NILLABLE_STRING,
			},
			{
				Name: "Age",
				Typ:  LWW_REGISTER,
				Kind: FieldKind_NILLABLE_INT,
			},
			{
				Name: "Custom",
				Typ:  LWW_REGISTER,
				Kind: FieldKind_NILLABLE_JSON,
			},
		},
	}
)

func TestNewFromJSON(t *testing.T) {
	ctx := context.Background()
	doc, err := NewDocFromJSON(ctx, testJSONObj, def)
	if err != nil {
		t.Error("Error creating new doc from JSON:", err)
		return
	}

	// check field/value
	// fields
	assert.Equal(t, doc.fields["Name"].Name(), "Name")
	assert.Equal(t, doc.fields["Name"].Type(), LWW_REGISTER)
	assert.Equal(t, doc.fields["Age"].Name(), "Age")
	assert.Equal(t, doc.fields["Age"].Type(), LWW_REGISTER)

	//values
	assert.Equal(t, doc.values[doc.fields["Name"]].Value(), "John")
	assert.Equal(t, doc.values[doc.fields["Name"]].IsDocument(), false)
	assert.Equal(t, doc.values[doc.fields["Age"]].Value(), int64(26))
	assert.Equal(t, doc.values[doc.fields["Age"]].IsDocument(), false)
}

func TestSetWithJSON(t *testing.T) {
	ctx := context.Background()
	doc, err := NewDocFromJSON(ctx, testJSONObj, def)
	if err != nil {
		t.Error("Error creating new doc from JSON:", err)
		return
	}

	updatePatch := []byte(`{
		"Name": "Alice",
		"Age": 27
	}`)
	err = doc.SetWithJSON(ctx, updatePatch)
	if err != nil {
		t.Error(err)
	}

	// check field/value
	// fields
	assert.Equal(t, doc.fields["Name"].Name(), "Name")
	assert.Equal(t, doc.fields["Name"].Type(), LWW_REGISTER)
	assert.Equal(t, doc.fields["Age"].Name(), "Age")
	assert.Equal(t, doc.fields["Age"].Type(), LWW_REGISTER)

	//values
	assert.Equal(t, doc.values[doc.fields["Name"]].Value(), "Alice")
	assert.Equal(t, doc.values[doc.fields["Name"]].IsDocument(), false)
	assert.Equal(t, doc.values[doc.fields["Age"]].Value(), int64(27))
	assert.Equal(t, doc.values[doc.fields["Age"]].IsDocument(), false)
}

func TestNewDocsFromJSON_WithObjectInsteadOfArray_Error(t *testing.T) {
	ctx := context.Background()
	_, err := NewDocsFromJSON(ctx, testJSONObj, def)
	require.ErrorContains(t, err, "value doesn't contain array; it contains object")
}

func TestNewFromJSON_WithValidJSONFieldValue_NoError(t *testing.T) {
	ctx := context.Background()
	objWithJSONField := []byte(`{
		"Name": "John",
		"Age": 26,
		"Custom": {
			"string": "maple", 
			"int": 260,
			"float": 3.14,
			"false": false,
			"true": true,
			"null": null,
			"array": ["one", 1],
			"object": {"one": 1}
		}
	}`)
	doc, err := NewDocFromJSON(ctx, objWithJSONField, def)
	if err != nil {
		t.Error("Error creating new doc from JSON:", err)
		return
	}

	// check field/value
	// fields
	assert.Equal(t, doc.fields["Name"].Name(), "Name")
	assert.Equal(t, doc.fields["Name"].Type(), LWW_REGISTER)
	assert.Equal(t, doc.fields["Age"].Name(), "Age")
	assert.Equal(t, doc.fields["Age"].Type(), LWW_REGISTER)
	assert.Equal(t, doc.fields["Custom"].Name(), "Custom")
	assert.Equal(t, doc.fields["Custom"].Type(), LWW_REGISTER)

	//values
	assert.Equal(t, doc.values[doc.fields["Name"]].Value(), "John")
	assert.Equal(t, doc.values[doc.fields["Name"]].IsDocument(), false)
	assert.Equal(t, doc.values[doc.fields["Age"]].Value(), int64(26))
	assert.Equal(t, doc.values[doc.fields["Age"]].IsDocument(), false)
	assert.Equal(t, doc.values[doc.fields["Custom"]].Value(), map[string]any{
		"string": "maple",
		"int":    float64(260),
		"float":  float64(3.14),
		"false":  false,
		"true":   true,
		"null":   nil,
		"array":  []any{"one", float64(1)},
		"object": map[string]any{"one": float64(1)},
	})
	assert.Equal(t, doc.values[doc.fields["Custom"]].IsDocument(), false)
}

func TestNewFromJSON_WithInvalidJSONFieldValue_Error(t *testing.T) {
	ctx := context.Background()
	objWithJSONField := []byte(`{
		"Name": "John",
		"Age": 26,
		"Custom": {"tree":"maple, "age": 260}
	}`)
	_, err := NewDocFromJSON(ctx, objWithJSONField, def)
	require.ErrorContains(t, err, "cannot parse JSON")
}

func TestNewFromJSON_WithJSONFieldValueSimpleString_Succeed(t *testing.T) {
	ctx := context.Background()
	objWithJSONField := []byte(`{
		"Name": "John",
		"Age": 26,
		"Custom": "blah"
	}`)
	_, err := NewDocFromJSON(ctx, objWithJSONField, def)
	require.NoError(t, err)
}

func TestIsJSONArray(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected bool
	}{
		{
			name:     "Valid JSON Array",
			input:    []byte(`[{"name":"John","age":21},{"name":"Islam","age":33}]`),
			expected: true,
		},
		{
			name:     "Valid Empty JSON Array",
			input:    []byte(`[]`),
			expected: true,
		},
		{
			name:     "Valid JSON Object",
			input:    []byte(`{"name":"John","age":21}`),
			expected: false,
		},
		{
			name:     "Invalid JSON String",
			input:    []byte(`{"name":"John","age":21`),
			expected: false,
		},
		{
			name:     "Non-JSON String",
			input:    []byte(`Hello, World!`),
			expected: false,
		},
		{
			name:     "Array of Primitives",
			input:    []byte(`[1, 2, 3, 4]`),
			expected: true,
		},
		{
			name:     "Nested JSON Array",
			input:    []byte(`[[1, 2], [3, 4]]`),
			expected: true,
		},
		{
			name: "Valid JSON Array with Whitespace",
			input: []byte(`
				[
					{	"name": "John", "age": 21	},
					{	"name": "Islam", "age": 33	}
				]
			`),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := IsJSONArray(tt.input)
			if actual != tt.expected {
				t.Errorf("IsJSONArray(%s) = %v; expected %v", tt.input, actual, tt.expected)
			}
		})
	}
}

func TestSet_WithNilElementInSliceForNonNillableArray_Error(t *testing.T) {
	ctx := context.Background()
	boolArrayDef := CollectionVersion{
		Name: "User",
		Fields: []CollectionFieldDescription{
			{
				Name: "flags",
				Typ:  LWW_REGISTER,
				Kind: FieldKind_BOOL_ARRAY,
			},
		},
	}
	doc, err := NewDocFromJSON(ctx, []byte(`{}`), boolArrayDef)
	require.NoError(t, err)

	err = doc.Set(ctx, "flags", []any{true, nil, false})
	require.ErrorContains(t, err, errNullValueForNonNillableField)
}

func TestNewDocFromJSON_OmittedNonNillableField_Error(t *testing.T) {
	ctx := context.Background()
	nonNillableDef := CollectionVersion{
		Name: "User",
		Fields: []CollectionFieldDescription{
			{
				Name: "score",
				Typ:  LWW_REGISTER,
				Kind: FieldKind_INT,
			},
		},
	}
	_, err := NewDocFromJSON(ctx, []byte(`{}`), nonNillableDef)
	require.ErrorContains(t, err, errMissingRequiredField)
}

func TestNewDocFromJSON_OmittedNonNillableArrayField_NoError(t *testing.T) {
	ctx := context.Background()
	nillableArrayDef := CollectionVersion{
		Name: "User",
		Fields: []CollectionFieldDescription{
			{
				Name: "scores",
				Typ:  LWW_REGISTER,
				Kind: FieldKind_INT_ARRAY,
			},
		},
	}
	_, err := NewDocFromJSON(ctx, []byte(`{}`), nillableArrayDef)
	require.NoError(t, err)
}

func TestNewDocFromMap_TypedStringSliceForNillableArray_Persisted(t *testing.T) {
	ctx := context.Background()
	nillableArrayDef := CollectionVersion{
		Name: "Block",
		Fields: []CollectionFieldDescription{
			{
				Name: "cids",
				Typ:  LWW_REGISTER,
				Kind: FieldKind_NILLABLE_STRING_ARRAY,
			},
		},
	}

	doc, err := NewDocFromMap(ctx, map[string]any{
		"cids": []string{"bafyA", "bafyB", "bafyC"},
	}, nillableArrayDef)
	require.NoError(t, err)

	val, err := doc.Get("cids")
	require.NoError(t, err)
	assert.Equal(t, []immutable.Option[string]{
		immutable.Some("bafyA"),
		immutable.Some("bafyB"),
		immutable.Some("bafyC"),
	}, val)
}

func TestNewDocFromMap_TypedTimeSliceForNillableArray_Persisted(t *testing.T) {
	ctx := context.Background()
	nillableArrayDef := CollectionVersion{
		Name: "Block",
		Fields: []CollectionFieldDescription{
			{
				Name: "times",
				Typ:  LWW_REGISTER,
				Kind: FieldKind_NILLABLE_DATETIME_ARRAY,
			},
		},
	}

	first := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	second := time.Date(2026, 7, 8, 12, 0, 0, 0, time.UTC)
	doc, err := NewDocFromMap(ctx, map[string]any{
		"times": []time.Time{first, second},
	}, nillableArrayDef)
	require.NoError(t, err)

	val, err := doc.Get("times")
	require.NoError(t, err)
	assert.Equal(t, []immutable.Option[time.Time]{
		immutable.Some(first),
		immutable.Some(second),
	}, val)
}
