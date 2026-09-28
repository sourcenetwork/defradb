// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package vectorindex

import (
	"github.com/sourcenetwork/defradb/client"
	dbmath "github.com/sourcenetwork/defradb/internal/utils/math"
)

// Score returns how near two vectors are under the given metric, where a larger result is nearer.
//
// Callers rank documents by this without knowing which metric produced it, which is what lets a query
// be answered by an index of any metric. Cosine is already a "larger is nearer" score; the other two
// are distances and are negated. Negation is monotonic, so each metric keeps its own ordering and
// only its direction changes.
//
// Only the shared leading elements are compared; callers that require equal lengths must check first.
func Score[T dbmath.Number](metric client.DistanceMetric, a, b []T) float64 {
	switch metric {
	case client.DistanceMetricEuclidean:
		return dbmath.NegativeSquaredEuclidean(a, b)
	case client.DistanceMetricDotProduct:
		return dbmath.Dot(a, b)
	default:
		return dbmath.CosineSimilarity(a, b)
	}
}
