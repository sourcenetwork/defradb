// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package math

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDot_ReturnsSumOfProducts(t *testing.T) {
	assert.Equal(t, float64(10), Dot([]int64{2, 4, 1}, []int64{1, 2, 0}))
	assert.Equal(t, float64(0), Dot([]float64{1, 0}, []float64{0, 1}))
}

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		a, b     []float64
		expected float64
	}{
		{"identical direction is 1", []float64{1, 0, 0}, []float64{1, 0, 0}, 1},
		{"magnitude does not matter", []float64{10, 0, 0}, []float64{1, 0, 0}, 1},
		{"orthogonal is 0", []float64{1, 0}, []float64{0, 1}, 0},
		{"opposite direction is -1", []float64{1, 0}, []float64{-1, 0}, -1},
		{"zero vector is 0", []float64{0, 0}, []float64{1, 1}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.InDelta(t, tt.expected, CosineSimilarity(tt.a, tt.b), 1e-9)
		})
	}
}

// The generic element type is honoured: an int vector and a float vector with the same directions
// give the same cosine.
func TestCosineSimilarity_IsGenericOverElementType(t *testing.T) {
	assert.InDelta(t,
		CosineSimilarity([]int64{2, 4, 1}, []int64{1, 2, 0}),
		CosineSimilarity([]float64{2, 4, 1}, []float64{1, 2, 0}),
		1e-9,
	)
}
