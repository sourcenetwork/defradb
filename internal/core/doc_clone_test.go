// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
)

// TestDocClone_CopiesScalarFields tests that Hidden, Status, and CollectionVersionID are copied to the clone.
func TestDocClone_CopiesScalarFields(t *testing.T) {
	original := Doc{
		Hidden:              true,
		Status:              client.Deleted,
		CollectionVersionID: "v1",
		Fields:              DocFields{"id1", "value"},
	}

	clone := original.Clone()

	assert.Equal(t, original.Hidden, clone.Hidden)
	assert.Equal(t, original.Status, clone.Status)
	assert.Equal(t, original.CollectionVersionID, clone.CollectionVersionID)
}

// TestDocClone_MutatingCloneFieldsDoesNotAffectOriginal tests that mutating the clone's Fields slice does not affect the original.
func TestDocClone_MutatingCloneFieldsDoesNotAffectOriginal(t *testing.T) {
	original := Doc{
		Fields: DocFields{"id1", "value"},
	}

	clone := original.Clone()
	clone.Fields[1] = "mutated"

	assert.Equal(t, "value", original.Fields[1])
}

// TestDocClone_DeepCopiesNestedDoc tests that a Doc nested inside Fields is deep copied, not shared.
func TestDocClone_DeepCopiesNestedDoc(t *testing.T) {
	inner := Doc{
		Hidden: true,
		Fields: DocFields{"inner-id", "inner-value"},
	}
	original := Doc{
		Fields: DocFields{"id1", inner},
	}

	clone := original.Clone()
	innerClone, ok := clone.Fields[1].(Doc)
	require.True(t, ok)
	innerClone.Fields[1] = "mutated"

	innerOriginal, ok := original.Fields[1].(Doc)
	require.True(t, ok)
	assert.Equal(t, "inner-value", innerOriginal.Fields[1])
}

// TestDocClone_DeepCopiesNestedDocSlice tests that a []Doc nested inside Fields is deep copied, not shared.
func TestDocClone_DeepCopiesNestedDocSlice(t *testing.T) {
	inner := Doc{
		Hidden: true,
		Fields: DocFields{"inner-id", "inner-value"},
	}
	original := Doc{
		Fields: DocFields{"id1", []Doc{inner}},
	}

	clone := original.Clone()
	innerClone, ok := clone.Fields[1].([]Doc)
	require.True(t, ok)
	innerClone[0].Fields[1] = "mutated"

	innerOriginal, ok := original.Fields[1].([]Doc)
	require.True(t, ok)
	assert.Equal(t, "inner-value", innerOriginal[0].Fields[1])
}
