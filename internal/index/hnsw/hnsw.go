// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// Package hnsw implements a standalone, pure-Go Hierarchical Navigable Small
// World (HNSW) approximate-nearest-neighbour graph, as described in
// Malkov & Yashunin, "Efficient and robust approximate nearest neighbor
// search using Hierarchical Navigable Small World graphs" (arXiv:1603.09320).
//
// This package is designed to be extraction-ready: it imports only the Go
// standard library. Persistence is abstracted behind the NodeStore
// interface defined in store.go, so callers can plug in any storage
// backend (an in-memory implementation is provided for tests and
// standalone use).
package hnsw

import "math"

// NodeID identifies a single vector/node in the graph. Callers (e.g. a
// database adapter) are responsible for mapping their own identifiers
// (such as document ids) to NodeID values.
type NodeID uint64

// Metric identifies a distance function used to compare vectors.
type Metric int

const (
	// Cosine is the cosine distance metric: 1 - dot(a, b) for unit-length
	// vectors a and b. It compares direction only.
	Cosine Metric = iota
	// Euclidean is the straight-line (L2) distance metric. It compares direction and magnitude.
	Euclidean
	// DotProduct is the (negated) dot product metric. It grows with magnitude, so a longer vector
	// pointing the same way is nearer.
	DotProduct
)

// requiresNormalized reports whether the metric needs unit-length vectors. Only cosine does; the
// others read magnitude, so scaling would change their answers.
func (m Metric) requiresNormalized() bool {
	return m == Cosine
}

// Params holds the tunable construction/search parameters of the graph.
type Params struct {
	// M is the number of bidirectional links created per node per layer
	// (except layer 0, which uses Mmax0).
	M int
	// Mmax0 is the maximum number of links a node may have at layer 0.
	Mmax0 int
	// EfConstruction is the size of the dynamic candidate list used
	// while inserting new nodes.
	EfConstruction int
	// EfSearch is the default size of the dynamic candidate list used
	// while searching, when the caller does not override it.
	EfSearch int
	// ML is the level-generation normalization factor used to draw a
	// node's top layer at insertion time.
	ML float64
}

// DefaultParams returns sensible defaults for the given M (the number of
// bidirectional links created per node per layer), following the
// recommendations in the HNSW paper.
func DefaultParams(m int) Params {
	// m == 1 gives ML = 1/ln(1) = +Inf, so randomLevel returns garbage and a node tries to allocate a
	// huge number of layers. The HNSW math needs m >= 2.
	if m <= 1 {
		m = 16
	}
	return Params{
		M:              m,
		Mmax0:          2 * m,
		EfConstruction: 128,
		EfSearch:       64,
		ML:             1 / math.Log(float64(m)),
	}
}

// normThreshold is the smallest squared norm normalize will scale. Below it a vector is treated as
// having negligible norm and returned unchanged: dividing by such a tiny norm would overflow float32
// to ±Inf rather than yield a meaningful unit vector.
const normThreshold = 1e-30

// normalize returns a unit-length copy of v. If v is the zero vector or has a negligible norm (its
// squared norm is below normThreshold), normalize returns a copy of v unchanged, since it cannot be
// meaningfully scaled to unit length; callers should avoid indexing such vectors when using the
// Cosine metric.
func normalize(v []float32) []float32 {
	// The squared norm is the vector's dot product with itself.
	sumSq := dot(v, v)
	if sumSq < normThreshold {
		out := make([]float32, len(v))
		copy(out, v)
		return out
	}

	norm := math.Sqrt(sumSq)
	out := make([]float32, len(v))
	for i, x := range v {
		out[i] = float32(float64(x) / norm)
	}
	return out
}

// vectorForMetric returns the form of v the metric compares. Both branches copy, matching what
// normalize already did for cosine. Nothing here needs the copy today (Insert hands the vector to a
// store that serializes it, and search only reads it), but it keeps a NodeStore that retains the
// slice from aliasing the caller's.
func vectorForMetric(metric Metric, v []float32) []float32 {
	if metric.requiresNormalized() {
		return normalize(v)
	}
	out := make([]float32, len(v))
	copy(out, v)
	return out
}

// dot returns the dot product of a and b, accumulated in float64 to avoid float32 rounding/underflow.
// If the lengths differ, only the shared leading elements are used.
func dot(a, b []float32) float64 {
	var sum float64
	n := min(len(a), len(b))
	for i := range n {
		sum += float64(a[i]) * float64(b[i])
	}
	return sum
}

// distance returns the distance between two vectors under the given metric, smaller being nearer.
// Both must have come through vectorForMetric for this metric. An unrecognised metric is a
// programming error rather than a silent fallback, so it panics.
func distance(metric Metric, a, b []float32) float32 {
	switch metric {
	case Cosine:
		return cosineDistance(a, b)
	case Euclidean:
		return squaredEuclideanDistance(a, b)
	case DotProduct:
		return dotProductDistance(a, b)
	default:
		panic("hnsw: unsupported distance metric")
	}
}

// cosineDistance computes 1 - dot(a, b) for unit-length vectors a and b.
func cosineDistance(a, b []float32) float32 {
	return float32(1 - dot(a, b))
}

// squaredEuclideanDistance computes the squared straight-line distance between a and b. The square
// root is skipped: it is monotonic, and distances are only ever compared against each other. If the
// lengths differ, only the shared leading elements are used, matching dot.
func squaredEuclideanDistance(a, b []float32) float32 {
	var sum float64
	n := min(len(a), len(b))
	for i := range n {
		d := float64(a[i]) - float64(b[i])
		sum += d * d
	}
	return float32(sum)
}

// dotProductDistance computes -dot(a, b). The dot product is a similarity (bigger is nearer), so the
// sign is flipped to keep smaller nearer.
func dotProductDistance(a, b []float32) float32 {
	return float32(-dot(a, b))
}
