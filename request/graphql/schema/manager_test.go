// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SchemaManager_NewNoErrs(t *testing.T) {
	_, err := NewSchemaManager()
	assert.NoError(t, err, "NewSchemaManager returned an error")
}

func Test_SchemaManager_HasDefaultTypes(t *testing.T) {
	s, err := NewSchemaManager()
	assert.NoError(t, err, "NewSchemaManager returned an error")

	tm := s.schema.TypeMap()
	for _, ty := range defaultTypes() {
		_, ok := tm[ty.Name()]
		assert.True(t, ok, "TypeMap missing default type %s", ty.Name())
	}
}

func Test_SchemaManager_ResolveTypes(t *testing.T) {
	s, _ := NewSchemaManager()
	err := s.ResolveTypes()
	assert.NoError(t, err, "Failed to ResolveTypes on a brand new SchemaManager")
}
