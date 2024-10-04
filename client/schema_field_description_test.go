// Copyright 2024 Democratized Data Foundation
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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestField_ScalarArray_HasSubKind(t *testing.T) {
	tests := []struct {
		name    string
		arrKind ScalarArrayKind
		subKind ScalarKind
	}{
		{
			name:    "bool array",
			arrKind: FieldKind_BOOL_ARRAY,
			subKind: FieldKind_NILLABLE_BOOL,
		},
		{
			name:    "int array",
			arrKind: FieldKind_INT_ARRAY,
			subKind: FieldKind_NILLABLE_INT,
		},
		{
			name:    "float array",
			arrKind: FieldKind_FLOAT_ARRAY,
			subKind: FieldKind_NILLABLE_FLOAT,
		},
		{
			name:    "string array",
			arrKind: FieldKind_STRING_ARRAY,
			subKind: FieldKind_NILLABLE_STRING,
		},
		{
			name:    "nillable bool array",
			arrKind: FieldKind_NILLABLE_BOOL_ARRAY,
			subKind: FieldKind_NILLABLE_BOOL,
		},
		{
			name:    "nillable int array",
			arrKind: FieldKind_NILLABLE_INT_ARRAY,
			subKind: FieldKind_NILLABLE_INT,
		},
		{
			name:    "nillable float array",
			arrKind: FieldKind_NILLABLE_FLOAT_ARRAY,
			subKind: FieldKind_NILLABLE_FLOAT,
		},
		{
			name:    "nillable string array",
			arrKind: FieldKind_NILLABLE_STRING_ARRAY,
			subKind: FieldKind_NILLABLE_STRING,
		},
		{
			name:    "nillable string array",
			arrKind: ScalarArrayKind(0),
			subKind: FieldKind_None,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.subKind, tt.arrKind.SubKind())
		})
	}
}
