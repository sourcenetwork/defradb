// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package crdt

import (
	"github.com/sourcenetwork/defradb/client"
)

// DocField is a struct that holds the field value being applied.
type DocField struct {
	// FieldName is the name of the field.
	FieldName string
	// FieldValue is the field value.
	FieldValue *client.FieldValue
}

// NewDocField creates a new DocField instance.
func NewDocField(fieldName string, fieldValue *client.FieldValue) *DocField {
	return &DocField{
		FieldName:  fieldName,
		FieldValue: fieldValue,
	}
}
