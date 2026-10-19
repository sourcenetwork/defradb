// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package keys

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
)

func TestActionStatusKey_ToString_CollectionWide(t *testing.T) {
	key := NewActionStatusKey("col1", client.TruncateAction)
	assert.Equal(t, ACTION_STATUS+"/col1/1", key.ToString())
}

func TestActionStatusKey_ToString_WithSubject(t *testing.T) {
	key := NewActionStatusSubjectKey("col1", client.BackfillIndexAction, "42")
	assert.Equal(t, ACTION_STATUS+"/col1/3/42", key.ToString())
}

func TestActionStatusKey_ToString_Empty(t *testing.T) {
	assert.Equal(t, ACTION_STATUS+"/", NewEmptyActionStatusKey().ToString())
}

func TestActionStatusKey_CollectionPrefix(t *testing.T) {
	key := NewActionStatusSubjectKey("col1", client.BackfillIndexAction, "42")
	assert.Equal(t, []byte(ACTION_STATUS+"/col1/"), key.CollectionPrefix())
}

func TestNewActionStatusKeyString(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    ActionStatusKey
		expectError bool
	}{
		{
			name:  "collection-wide key round-trips",
			input: ACTION_STATUS + "/col1/1",
			expected: ActionStatusKey{
				CollectionID: "col1",
				Action:       client.TruncateAction,
			},
		},
		{
			name:  "subject key round-trips",
			input: ACTION_STATUS + "/col1/3/42",
			expected: ActionStatusKey{
				CollectionID: "col1",
				Action:       client.BackfillIndexAction,
				Subject:      "42",
			},
		},
		{
			name:        "missing action component",
			input:       ACTION_STATUS + "/col1",
			expectError: true,
		},
		{
			name:        "too many components",
			input:       ACTION_STATUS + "/col1/3/42/extra",
			expectError: true,
		},
		{
			name:        "non-numeric action",
			input:       ACTION_STATUS + "/col1/notanumber",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewActionStatusKeyString(tt.input)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestActionStatusKey_RoundTrip(t *testing.T) {
	for _, original := range []ActionStatusKey{
		NewActionStatusKey("testcollection", client.RefreshDatastoreAction),
		NewActionStatusSubjectKey("testcollection", client.BackfillIndexAction, "99"),
	} {
		parsed, err := NewActionStatusKeyString(original.ToString())
		require.NoError(t, err)
		assert.Equal(t, original, parsed)
	}
}
