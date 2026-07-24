// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// Package math holds small, reusable math helpers shared across DefraDB. It is deliberately general:
// vector helpers live here now, but any other shared math can join them. The package is named math;
// import it explicitly (e.g. deframath "github.com/.../internal/utils/math") in files that also need
// the standard library math package.
package math

import "math"

// Number is the set of numeric element types the vector helpers accept.
type Number interface {
	int64 | float64 | float32
}

// Dot returns the dot product of two equal-length vectors, accumulated in float64. The caller is
// responsible for passing equal-length slices; extra elements of the longer slice are ignored.
func Dot[T Number](a, b []T) float64 {
	var sum float64
	n := min(len(a), len(b))
	for i := range n {
		sum += float64(a[i]) * float64(b[i])
	}
	return sum
}

// CosineSimilarity returns the cosine similarity of two equal-length vectors: their dot product
// divided by the product of their magnitudes, in the range [-1, 1] where 1 means the same direction.
// Dividing by the magnitudes is what makes this a true cosine, so the result does not depend on the
// vectors' lengths (a long vector and a short one pointing the same way score 1).
//
// A zero-length vector has no direction, so its cosine similarity to anything is defined as 0 rather
// than a division by zero. The caller is responsible for passing equal-length slices.
func CosineSimilarity[T Number](a, b []T) float64 {
	var dot, aNorm, bNorm float64
	n := min(len(a), len(b))
	for i := range n {
		x, y := float64(a[i]), float64(b[i])
		dot += x * y
		aNorm += x * x
		bNorm += y * y
	}
	if aNorm == 0 || bNorm == 0 {
		return 0
	}
	return dot / (math.Sqrt(aNorm) * math.Sqrt(bNorm))
}
