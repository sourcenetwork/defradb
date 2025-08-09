// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/state"
)

// mockTestState is a simple mock implementation of TestState interface for testing
type mockTestState struct {
	clientType    state.ClientType
	currentNodeID int
}

func (m *mockTestState) GetClientType() state.ClientType {
	return m.clientType
}

func (m *mockTestState) GetCurrentNodeID() int {
	return m.currentNodeID
}

func (m *mockTestState) GetIdentity(_ state.Identity) acpIdentity.Identity {
	return nil
}

func (m *mockTestState) GetDocID(_ int, _ int) client.DocID {
	return client.DocID{}
}

func TestAnyOfMatcher(t *testing.T) {
	tests := []struct {
		name       string
		clientType state.ClientType
		values     []any
		actual     any
		want       bool
	}{
		{
			name:       "HTTP client with matching value",
			clientType: HTTPClientType,
			values:     []any{"value1", "value2", "value3"},
			actual:     "value2",
			want:       true,
		},
		{
			name:       "CLI client with matching value",
			clientType: CLIClientType,
			values:     []any{1, 2, 3},
			actual:     2,
			want:       true,
		},
		{
			name:       "Go client with matching value",
			clientType: GoClientType,
			values:     []any{"value1", "value2", "value3"},
			actual:     "value2",
			want:       true,
		},
		{
			name:       "HTTP client with non-matching value",
			clientType: HTTPClientType,
			values:     []any{"value1", "value2", "value3"},
			actual:     "value4",
			want:       false,
		},
		{
			name:       "Go client with non-matching value",
			clientType: GoClientType,
			values:     []any{1, 2, 3},
			actual:     4,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockState := &mockTestState{clientType: tt.clientType}
			matcher := AnyOf(tt.values...)
			matcher.SetTestState(mockState)

			got, err := matcher.Match(tt.actual)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUniqueValueMatcher(t *testing.T) {
	tests := []struct {
		name      string
		nodeIDs   []int
		values    []any
		wantMatch []bool
	}{
		{
			name:      "All unique values for same node",
			nodeIDs:   []int{0, 0, 0},
			values:    []any{"a", "b", "c"},
			wantMatch: []bool{true, true, true},
		},
		{
			name:      "Duplicate values for same node",
			nodeIDs:   []int{0, 0, 0},
			values:    []any{"a", "a", "b"},
			wantMatch: []bool{true, false, true},
		},
		{
			name:      "Same values across different nodes should be allowed",
			nodeIDs:   []int{0, 1, 0, 1, 2},
			values:    []any{"a", "a", "b", "b", "a"},
			wantMatch: []bool{true, true, true, true, true},
		},
		{
			name:      "Duplicate values for different nodes",
			nodeIDs:   []int{0, 1, 0, 1, 1, 1},
			values:    []any{"a", "a", "a", "c", "b", "c"},
			wantMatch: []bool{true, true, false, true, true, false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := NewUniqueValue()

			for i := range tt.values {
				mockState := &mockTestState{currentNodeID: tt.nodeIDs[i]}
				matcher.SetTestState(mockState)

				got, err := matcher.Match(tt.values[i])
				assert.NoError(t, err)
				assert.Equal(t, tt.wantMatch[i], got)
			}
		})
	}
}

func TestSameValueMatcher(t *testing.T) {
	tests := []struct {
		name      string
		values    []any
		wantMatch []bool
	}{
		{
			name:      "First value always matches",
			values:    []any{"value1"},
			wantMatch: []bool{true},
		},
		{
			name:      "All same values should match",
			values:    []any{"value1", "value1", "value1"},
			wantMatch: []bool{true, true, true},
		},
		{
			name:      "Different values shouldn't match after the first",
			values:    []any{"value1", "value2", "value3"},
			wantMatch: []bool{true, false, false},
		},
		{
			name:      "Mixed values - only matching ones should return true",
			values:    []any{42, 42, 43, 42},
			wantMatch: []bool{true, true, false, true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := NewSameValue()

			for i, val := range tt.values {
				got, err := matcher.Match(val)
				assert.NoError(t, err)
				assert.Equal(t, tt.wantMatch[i], got, "Value: %v, Index: %d", val, i)
			}
		})
	}
}

func TestSameValueMatcher_ResetState_ShouldReset(t *testing.T) {
	matcher := NewSameValue()

	got, err := matcher.Match(1)
	assert.NoError(t, err)
	assert.True(t, got)

	got, err = matcher.Match(2)
	assert.NoError(t, err)
	assert.False(t, got)

	matcher.ResetMatcherState()

	got, err = matcher.Match(2)
	assert.NoError(t, err)
	assert.True(t, got)
}
