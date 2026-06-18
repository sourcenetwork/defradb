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

// TestDocClone_MutatingCloneDoesNotAffectOriginal tests that mutating the clone's fields does not affect the original.
func TestDocClone_MutatingCloneDoesNotAffectOriginal(t *testing.T) {
	original := Doc{
		Hidden:              true,
		Status:              client.Deleted,
		CollectionVersionID: "v1",
		Fields:              DocFields{"id1", "value"},
	}

	// The following lines intentionally mutate the clone to confirm the original is not affected.
	// These are "unused writes", but this is intentional.
	clone := original.Clone()
	clone.Hidden = false             //nolint:ineffassign
	clone.Status = client.Active     //nolint:ineffassign
	clone.CollectionVersionID = "v2" //nolint:ineffassign
	clone.Fields[1] = "mutated"      //nolint:ineffassign

	assert.True(t, original.Hidden)
	assert.Equal(t, client.Deleted, original.Status)
	assert.Equal(t, "v1", original.CollectionVersionID)
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
	innerClone := clone.Fields[1].(Doc)
	innerClone.Fields[1] = "mutated"

	assert.Equal(t, "inner-value", original.Fields[1].(Doc).Fields[1])
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
	innerClone := clone.Fields[1].([]Doc)
	innerClone[0].Fields[1] = "mutated"

	assert.Equal(t, "inner-value", original.Fields[1].([]Doc)[0].Fields[1])
}
