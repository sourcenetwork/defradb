// Copyright 2022 Democratized Data Foundation
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

	"github.com/sourcenetwork/defradb/internal/keys"
)

func TestMergeAscending_ReturnsEmpty_GivenEmpty(t *testing.T) {
	input := []Span{}

	result := MergeAscending(input)

	assert.Empty(t, result)
}

func TestMergeAscending_ReturnsSingle_GivenSingle(t *testing.T) {
	start1 := keys.MustNewDataStoreKey("/1/p/0/k1")
	end1 := keys.MustNewDataStoreKey("/1/p/0/k2")
	input := []Span{NewSpan(start1, end1)}

	result := MergeAscending(input)

	assert.Len(t, result, 1)
	assert.Equal(t, start1, result[0].Start())
	assert.Equal(t, end1, result[0].End())
}

func TestMergeAscending_ReturnsSecondBeforeFirst_GivenKeysInReverseOrder(t *testing.T) {
	start1 := keys.MustNewDataStoreKey("/1/p/0/k4")
	end1 := keys.MustNewDataStoreKey("/1/p/0/k5")
	start2 := keys.MustNewDataStoreKey("/1/p/0/k1")
	end2 := keys.MustNewDataStoreKey("/1/p/0/k2")

	input := []Span{
		NewSpan(start1, end1),
		NewSpan(start2, end2),
	}

	result := MergeAscending(input)

	assert.Len(t, result, 2)
	assert.Equal(t, start2, result[0].Start())
	assert.Equal(t, end2, result[0].End())
	assert.Equal(t, start1, result[1].Start())
	assert.Equal(t, end1, result[1].End())
}

func TestMergeAscending_ReturnsItemsInOrder_GivenKeysInMixedOrder(t *testing.T) {
	start1 := keys.MustNewDataStoreKey("/1/p/0/k1")
	end1 := keys.MustNewDataStoreKey("/1/p/0/k2")
	start2 := keys.MustNewDataStoreKey("/1/p/0/k7")
	end2 := keys.MustNewDataStoreKey("/1/p/0/k8")
	start3 := keys.MustNewDataStoreKey("/1/p/0/k4")
	end3 := keys.MustNewDataStoreKey("/1/p/0/k5")

	input := []Span{
		NewSpan(start1, end1),
		NewSpan(start2, end2),
		NewSpan(start3, end3),
	}

	result := MergeAscending(input)

	assert.Len(t, result, 3)
	assert.Equal(t, start1, result[0].Start())
	assert.Equal(t, end1, result[0].End())
	// Span 3 should be returned between one and two
	assert.Equal(t, start3, result[1].Start())
	assert.Equal(t, end3, result[1].End())
	assert.Equal(t, start2, result[2].Start())
	assert.Equal(t, end2, result[2].End())
}

func TestMergeAscending_ReturnsSingle_GivenStartBeforeEndEqualToStart(t *testing.T) {
	start1 := keys.MustNewDataStoreKey("/1/p/0/k3")
	end1 := keys.MustNewDataStoreKey("/1/p/0/k4")
	start2 := keys.MustNewDataStoreKey("/1/p/0/k1")
	end2 := keys.MustNewDataStoreKey("/1/p/0/k3")
	input := []Span{
		NewSpan(start1, end1),
		NewSpan(start2, end2),
	}

	result := MergeAscending(input)

	assert.Len(t, result, 1)
	assert.Equal(t, start2, result[0].Start())
	assert.Equal(t, end1, result[0].End())
}

func TestMergeAscending_ReturnsSingle_GivenStartBeforeEndAdjacentToStart(t *testing.T) {
	start1 := keys.MustNewDataStoreKey("/1/p/0/k3")
	end1 := keys.MustNewDataStoreKey("/1/p/0/k4")
	start2 := keys.MustNewDataStoreKey("/1/p/0/k1")
	end2 := keys.MustNewDataStoreKey("/1/p/0/k2")
	input := []Span{
		NewSpan(start1, end1),
		NewSpan(start2, end2),
	}

	result := MergeAscending(input)

	assert.Len(t, result, 1)
	assert.Equal(t, start2, result[0].Start())
	assert.Equal(t, end1, result[0].End())
}

func TestMergeAscending_ReturnsSingle_GivenStartBeforeEndWithin(t *testing.T) {
	start1 := keys.MustNewDataStoreKey("/1/p/0/k3")
	end1 := keys.MustNewDataStoreKey("/1/p/0/k4")
	start2 := keys.MustNewDataStoreKey("/1/p/0/k1")
	end2 := keys.MustNewDataStoreKey("/1/p/0/k3.5")
	input := []Span{
		NewSpan(start1, end1),
		NewSpan(start2, end2),
	}

	result := MergeAscending(input)

	assert.Len(t, result, 1)
	assert.Equal(t, start2, result[0].Start())
	assert.Equal(t, end1, result[0].End())
}

func TestMergeAscending_ReturnsSingle_GivenStartPrefixesEndWithin(t *testing.T) {
	start1 := keys.MustNewDataStoreKey("/1/p/0/k1.1")
	end1 := keys.MustNewDataStoreKey("/1/p/0/k3")
	start2 := keys.MustNewDataStoreKey("/1/p/0/k1")
	end2 := keys.MustNewDataStoreKey("/1/p/0/k2.5")
	input := []Span{
		NewSpan(start1, end1),
		NewSpan(start2, end2),
	}

	result := MergeAscending(input)

	assert.Len(t, result, 1)
	assert.Equal(t, start2, result[0].Start())
	assert.Equal(t, end1, result[0].End())
}

func TestMergeAscending_ReturnsSingle_GivenStartBeforeEndWithinEndPrefix(t *testing.T) {
	start1 := keys.MustNewDataStoreKey("/1/p/0/k3")
	end1 := keys.MustNewDataStoreKey("/1/p/0/k4")
	start2 := keys.MustNewDataStoreKey("/1/p/0/k1")
	end2 := keys.MustNewDataStoreKey("/1/p/0/k4.5")
	input := []Span{
		NewSpan(start1, end1),
		NewSpan(start2, end2),
	}

	result := MergeAscending(input)

	assert.Len(t, result, 1)
	assert.Equal(t, start2, result[0].Start())
	assert.Equal(t, end1, result[0].End())
}

func TestMergeAscending_ReturnsSingle_GivenStartPrefixesEndWithinEndPrefix(t *testing.T) {
	start1 := keys.MustNewDataStoreKey("/1/p/0/k1.1")
	end1 := keys.MustNewDataStoreKey("/1/p/0/k3")
	start2 := keys.MustNewDataStoreKey("/1/p/0/k1")
	end2 := keys.MustNewDataStoreKey("/1/p/0/k3.5")
	input := []Span{
		NewSpan(start1, end1),
		NewSpan(start2, end2),
	}

	result := MergeAscending(input)

	assert.Len(t, result, 1)
	assert.Equal(t, start2, result[0].Start())
	assert.Equal(t, end1, result[0].End())
}

func TestMergeAscending_ReturnsSingle_GivenStartBeforeEndEqual(t *testing.T) {
	start1 := keys.MustNewDataStoreKey("/1/p/0/k3")
	end1 := keys.MustNewDataStoreKey("/1/p/0/k4")
	start2 := keys.MustNewDataStoreKey("/1/p/0/k1")
	end2 := keys.MustNewDataStoreKey("/1/p/0/k4")
	input := []Span{
		NewSpan(start1, end1),
		NewSpan(start2, end2),
	}

	result := MergeAscending(input)

	assert.Len(t, result, 1)
	assert.Equal(t, start2, result[0].Start())
	assert.Equal(t, end1, result[0].End())
}

func TestMergeAscending_ReturnsSingle_GivenStartBeforeEndAdjacentAndBefore(t *testing.T) {
	start1 := keys.MustNewDataStoreKey("/1/p/0/k3")
	end1 := keys.MustNewDataStoreKey("/1/p/0/k5")
	start2 := keys.MustNewDataStoreKey("/1/p/0/k1")
	end2 := keys.MustNewDataStoreKey("/1/p/0/k4")
	input := []Span{
		NewSpan(start1, end1),
		NewSpan(start2, end2),
	}

	result := MergeAscending(input)

	assert.Len(t, result, 1)
	assert.Equal(t, start2, result[0].Start())
	assert.Equal(t, end1, result[0].End())
}

func TestMergeAscending_ReturnsSingle_GivenStartBeforeEndAdjacentAndGreater(t *testing.T) {
	start1 := keys.MustNewDataStoreKey("/1/p/0/k3")
	end1 := keys.MustNewDataStoreKey("/1/p/0/k4")
	start2 := keys.MustNewDataStoreKey("/1/p/0/k1")
	end2 := keys.MustNewDataStoreKey("/1/p/0/k5")
	input := []Span{
		NewSpan(start1, end1),
		NewSpan(start2, end2),
	}

	result := MergeAscending(input)

	assert.Len(t, result, 1)
	assert.Equal(t, start2, result[0].Start())
	assert.Equal(t, end2, result[0].End())
}

func TestMergeAscending_ReturnsSingle_GivenStartPrefixesEndEqual(t *testing.T) {
	start1 := keys.MustNewDataStoreKey("/1/p/0/k1.1")
	end1 := keys.MustNewDataStoreKey("/1/p/0/k3")
	start2 := keys.MustNewDataStoreKey("/1/p/0/k1")
	end2 := keys.MustNewDataStoreKey("/1/p/0/k3")
	input := []Span{
		NewSpan(start1, end1),
		NewSpan(start2, end2),
	}

	result := MergeAscending(input)

	assert.Len(t, result, 1)
	assert.Equal(t, start2, result[0].Start())
	assert.Equal(t, end1, result[0].End())
}

func TestMergeAscending_ReturnsSingle_GivenStartPrefixesEndAdjacentAndBefore(t *testing.T) {
	start1 := keys.MustNewDataStoreKey("/1/p/0/k1.1")
	end1 := keys.MustNewDataStoreKey("/1/p/0/k3")
	start2 := keys.MustNewDataStoreKey("/1/p/0/k1")
	end2 := keys.MustNewDataStoreKey("/1/p/0/k2")
	input := []Span{
		NewSpan(start1, end1),
		NewSpan(start2, end2),
	}

	result := MergeAscending(input)

	assert.Len(t, result, 1)
	assert.Equal(t, start2, result[0].Start())
	assert.Equal(t, end1, result[0].End())
}

func TestMergeAscending_ReturnsSingle_GivenStartPrefixesEndAdjacentAndAfter(t *testing.T) {
	start1 := keys.MustNewDataStoreKey("/1/p/0/k1.1")
	end1 := keys.MustNewDataStoreKey("/1/p/0/k3")
	start2 := keys.MustNewDataStoreKey("/1/p/0/k1")
	end2 := keys.MustNewDataStoreKey("/1/p/0/k4")
	input := []Span{
		NewSpan(start1, end1),
		NewSpan(start2, end2),
	}

	result := MergeAscending(input)

	assert.Len(t, result, 1)
	assert.Equal(t, start2, result[0].Start())
	assert.Equal(t, end2, result[0].End())
}

func TestMergeAscending_ReturnsMiddleSpansMerged_GivenSpanCoveringMiddleSpans(t *testing.T) {
	start1 := keys.MustNewDataStoreKey("/1/p/0/k1")
	end1 := keys.MustNewDataStoreKey("/1/p/0/k2")
	start2 := keys.MustNewDataStoreKey("/1/p/0/k6")
	end2 := keys.MustNewDataStoreKey("/1/p/0/k7")
	start3 := keys.MustNewDataStoreKey("/1/p/0/k9")
	end3 := keys.MustNewDataStoreKey("/1/p/0/ka")
	start4 := keys.MustNewDataStoreKey("/1/p/0/kc")
	end4 := keys.MustNewDataStoreKey("/1/p/0/kd")
	start5 := keys.MustNewDataStoreKey("/1/p/0/k4")
	end5 := keys.MustNewDataStoreKey("/1/p/0/ka")
	input := []Span{
		NewSpan(start1, end1),
		NewSpan(start2, end2),
		NewSpan(start3, end3),
		NewSpan(start4, end4),
		NewSpan(start5, end5),
	}

	result := MergeAscending(input)

	assert.Len(t, result, 3)
	assert.Equal(t, start1, result[0].Start())
	assert.Equal(t, end1, result[0].End())
	// Spans 2 and 3 are within span 5
	assert.Equal(t, start5, result[1].Start())
	assert.Equal(t, end5, result[1].End())
	assert.Equal(t, start4, result[2].Start())
	assert.Equal(t, end4, result[2].End())
}

func TestMergeAscending_ReturnsSingle_GivenStartEqualEndWithin(t *testing.T) {
	start1 := keys.MustNewDataStoreKey("/1/p/0/k1")
	end1 := keys.MustNewDataStoreKey("/1/p/0/k2")
	start2 := keys.MustNewDataStoreKey("/1/p/0/k1")
	end2 := keys.MustNewDataStoreKey("/1/p/0/k1.5")
	input := []Span{
		NewSpan(start1, end1),
		NewSpan(start2, end2),
	}

	result := MergeAscending(input)

	assert.Len(t, result, 1)
	assert.Equal(t, start1, result[0].Start())
	assert.Equal(t, end1, result[0].End())
}

func TestMergeAscending_ReturnsSingle_GivenStartEqualEndWithinEndPrefix(t *testing.T) {
	start1 := keys.MustNewDataStoreKey("/1/p/0/k1")
	end1 := keys.MustNewDataStoreKey("/1/p/0/k2")
	start2 := keys.MustNewDataStoreKey("/1/p/0/k1")
	end2 := keys.MustNewDataStoreKey("/1/p/0/k2.5")
	input := []Span{
		NewSpan(start1, end1),
		NewSpan(start2, end2),
	}

	result := MergeAscending(input)

	assert.Len(t, result, 1)
	assert.Equal(t, start1, result[0].Start())
	assert.Equal(t, end1, result[0].End())
}

func TestMergeAscending_ReturnsSingle_GivenDuplicates(t *testing.T) {
	start1 := keys.MustNewDataStoreKey("/1/p/0/k1")
	end1 := keys.MustNewDataStoreKey("/1/p/0/k2")
	input := []Span{
		NewSpan(start1, end1),
		NewSpan(start1, end1),
	}

	result := MergeAscending(input)

	assert.Len(t, result, 1)
	assert.Equal(t, start1, result[0].Start())
	assert.Equal(t, end1, result[0].End())
}

func TestMergeAscending_ReturnsSingle_GivenStartWithinEndWithin(t *testing.T) {
	start1 := keys.MustNewDataStoreKey("/1/p/0/k1")
	end1 := keys.MustNewDataStoreKey("/1/p/0/k2")
	start2 := keys.MustNewDataStoreKey("/1/p/0/k1.2")
	end2 := keys.MustNewDataStoreKey("/1/p/0/k1.5")
	input := []Span{
		NewSpan(start1, end1),
		NewSpan(start2, end2),
	}

	result := MergeAscending(input)

	assert.Len(t, result, 1)
	assert.Equal(t, start1, result[0].Start())
	assert.Equal(t, end1, result[0].End())
}

func TestMergeAscending_ReturnsSingle_GivenStartWithinEndWithinEndPrefix(t *testing.T) {
	start1 := keys.MustNewDataStoreKey("/1/p/0/k1")
	end1 := keys.MustNewDataStoreKey("/1/p/0/k2")
	start2 := keys.MustNewDataStoreKey("/1/p/0/k1.2")
	end2 := keys.MustNewDataStoreKey("/1/p/0/k2.5")
	input := []Span{
		NewSpan(start1, end1),
		NewSpan(start2, end2),
	}

	result := MergeAscending(input)

	assert.Len(t, result, 1)
	assert.Equal(t, start1, result[0].Start())
	assert.Equal(t, end1, result[0].End())
}

func TestMergeAscending_ReturnsSingle_GivenStartWithinEndEqual(t *testing.T) {
	start1 := keys.MustNewDataStoreKey("/1/p/0/k1")
	end1 := keys.MustNewDataStoreKey("/1/p/0/k2")
	start2 := keys.MustNewDataStoreKey("/1/p/0/k1.2")
	end2 := keys.MustNewDataStoreKey("/1/p/0/k2")
	input := []Span{
		NewSpan(start1, end1),
		NewSpan(start2, end2),
	}

	result := MergeAscending(input)

	assert.Len(t, result, 1)
	assert.Equal(t, start1, result[0].Start())
	assert.Equal(t, end1, result[0].End())
}

func TestMergeAscending_ReturnsSingle_GivenStartWithinEndAdjacentAndBefore(t *testing.T) {
	start1 := keys.MustNewDataStoreKey("/1/p/0/k1")
	end1 := keys.MustNewDataStoreKey("/1/p/0/k3")
	start2 := keys.MustNewDataStoreKey("/1/p/0/k1.2")
	end2 := keys.MustNewDataStoreKey("/1/p/0/k2")
	input := []Span{
		NewSpan(start1, end1),
		NewSpan(start2, end2),
	}

	result := MergeAscending(input)

	assert.Len(t, result, 1)
	assert.Equal(t, start1, result[0].Start())
	assert.Equal(t, end1, result[0].End())
}

func TestMergeAscending_ReturnsSingle_GivenStartWithinEndAdjacentAndAfter(t *testing.T) {
	start1 := keys.MustNewDataStoreKey("/1/p/0/k1")
	end1 := keys.MustNewDataStoreKey("/1/p/0/k3")
	start2 := keys.MustNewDataStoreKey("/1/p/0/k1.2")
	end2 := keys.MustNewDataStoreKey("/1/p/0/k4")
	input := []Span{
		NewSpan(start1, end1),
		NewSpan(start2, end2),
	}

	result := MergeAscending(input)

	assert.Len(t, result, 1)
	assert.Equal(t, start1, result[0].Start())
	assert.Equal(t, end2, result[0].End())
}

func TestMergeAscending_ReturnsMiddleSpansMerged_GivenStartEqualEndAfterSpanCoveringMiddleSpans(
	t *testing.T,
) {
	start1 := keys.MustNewDataStoreKey("/1/p/0/k1")
	end1 := keys.MustNewDataStoreKey("/1/p/0/k2")
	start2 := keys.MustNewDataStoreKey("/1/p/0/k4")
	end2 := keys.MustNewDataStoreKey("/1/p/0/k5")
	start3 := keys.MustNewDataStoreKey("/1/p/0/k7")
	end3 := keys.MustNewDataStoreKey("/1/p/0/k8")
	start4 := keys.MustNewDataStoreKey("/1/p/0/kc")
	end4 := keys.MustNewDataStoreKey("/1/p/0/kd")
	start5 := keys.MustNewDataStoreKey("/1/p/0/k4") // equal to start2
	end5 := keys.MustNewDataStoreKey("/1/p/0/ka")
	input := []Span{
		NewSpan(start1, end1),
		NewSpan(start2, end2),
		NewSpan(start3, end3),
		NewSpan(start4, end4),
		NewSpan(start5, end5),
	}

	result := MergeAscending(input)

	assert.Len(t, result, 3)
	assert.Equal(t, start1, result[0].Start())
	assert.Equal(t, end1, result[0].End())
	// Spans 2 and 3 are within span 5
	assert.Equal(t, start5, result[1].Start())
	assert.Equal(t, end5, result[1].End())
	assert.Equal(t, start4, result[2].Start())
	assert.Equal(t, end4, result[2].End())
}

func TestMergeAscending_ReturnsMiddleSpansMerged_GivenStartWithinEndAfterSpanCoveringMiddleSpans(
	t *testing.T,
) {
	start1 := keys.MustNewDataStoreKey("/1/p/0/k1")
	end1 := keys.MustNewDataStoreKey("/1/p/0/k2")
	start2 := keys.MustNewDataStoreKey("/1/p/0/k4")
	end2 := keys.MustNewDataStoreKey("/1/p/0/k5")
	start3 := keys.MustNewDataStoreKey("/1/p/0/k7")
	end3 := keys.MustNewDataStoreKey("/1/p/0/k8")
	start4 := keys.MustNewDataStoreKey("/1/p/0/kc")
	end4 := keys.MustNewDataStoreKey("/1/p/0/kd")
	start5 := keys.MustNewDataStoreKey("/1/p/0/k4.5") // within span2
	end5 := keys.MustNewDataStoreKey("/1/p/0/ka")
	input := []Span{
		NewSpan(start1, end1),
		NewSpan(start2, end2),
		NewSpan(start3, end3),
		NewSpan(start4, end4),
		NewSpan(start5, end5),
	}

	result := MergeAscending(input)

	assert.Len(t, result, 3)
	assert.Equal(t, start1, result[0].Start())
	assert.Equal(t, end1, result[0].End())
	assert.Equal(t, start2, result[1].Start())
	assert.Equal(t, end5, result[1].End())
	assert.Equal(t, start4, result[2].Start())
	assert.Equal(t, end4, result[2].End())
}

func TestMergeAscending_ReturnsMiddleSpansMerged_GivenStartEqualToEndEndAfterSpanCoveringMiddleSpans(
	t *testing.T,
) {
	start1 := keys.MustNewDataStoreKey("/1/p/0/k1")
	end1 := keys.MustNewDataStoreKey("/1/p/0/k2")
	start2 := keys.MustNewDataStoreKey("/1/p/0/k4")
	end2 := keys.MustNewDataStoreKey("/1/p/0/k5")
	start3 := keys.MustNewDataStoreKey("/1/p/0/k7")
	end3 := keys.MustNewDataStoreKey("/1/p/0/k8")
	start4 := keys.MustNewDataStoreKey("/1/p/0/kc")
	end4 := keys.MustNewDataStoreKey("/1/p/0/kd")
	start5 := keys.MustNewDataStoreKey("/1/p/0/k5") // span2's end
	end5 := keys.MustNewDataStoreKey("/1/p/0/ka")
	input := []Span{
		NewSpan(start1, end1),
		NewSpan(start2, end2),
		NewSpan(start3, end3),
		NewSpan(start4, end4),
		NewSpan(start5, end5),
	}

	result := MergeAscending(input)

	assert.Len(t, result, 3)
	assert.Equal(t, start1, result[0].Start())
	assert.Equal(t, end1, result[0].End())
	assert.Equal(t, start2, result[1].Start())
	assert.Equal(t, end5, result[1].End())
	assert.Equal(t, start4, result[2].Start())
	assert.Equal(t, end4, result[2].End())
}

func TestMergeAscending_ReturnsMiddleSpansMerged_GivenStartAdjacentAndBeforeEndEndAfterSpanCoveringMiddleSpans(
	t *testing.T,
) {
	start1 := keys.MustNewDataStoreKey("/1/p/0/k1")
	end1 := keys.MustNewDataStoreKey("/1/p/0/k2")
	start2 := keys.MustNewDataStoreKey("/1/p/0/k4")
	end2 := keys.MustNewDataStoreKey("/1/p/0/k6")
	start3 := keys.MustNewDataStoreKey("/1/p/0/k8")
	end3 := keys.MustNewDataStoreKey("/1/p/0/k9")
	start4 := keys.MustNewDataStoreKey("/1/p/0/kd")
	end4 := keys.MustNewDataStoreKey("/1/p/0/ke")
	start5 := keys.MustNewDataStoreKey("/1/p/0/k5") // adjacent but before span2's end
	end5 := keys.MustNewDataStoreKey("/1/p/0/kb")
	input := []Span{
		NewSpan(start1, end1),
		NewSpan(start2, end2),
		NewSpan(start3, end3),
		NewSpan(start4, end4),
		NewSpan(start5, end5),
	}

	result := MergeAscending(input)

	assert.Len(t, result, 3)
	assert.Equal(t, start1, result[0].Start())
	assert.Equal(t, end1, result[0].End())
	assert.Equal(t, start2, result[1].Start())
	assert.Equal(t, end5, result[1].End())
	assert.Equal(t, start4, result[2].Start())
	assert.Equal(t, end4, result[2].End())
}

func TestMergeAscending_ReturnsMiddleSpansMerged_GivenStartAdjacentAndAfterEndEndAfterSpanCoveringMiddleSpans(
	t *testing.T,
) {
	start1 := keys.MustNewDataStoreKey("/1/p/0/k1")
	end1 := keys.MustNewDataStoreKey("/1/p/0/k2")
	start2 := keys.MustNewDataStoreKey("/1/p/0/k4")
	end2 := keys.MustNewDataStoreKey("/1/p/0/k5")
	start3 := keys.MustNewDataStoreKey("/1/p/0/k8")
	end3 := keys.MustNewDataStoreKey("/1/p/0/k9")
	start4 := keys.MustNewDataStoreKey("/1/p/0/kd")
	end4 := keys.MustNewDataStoreKey("/1/p/0/ke")
	start5 := keys.MustNewDataStoreKey("/1/p/0/k6") // adjacent and after span2's end
	end5 := keys.MustNewDataStoreKey("/1/p/0/kb")
	input := []Span{
		NewSpan(start1, end1),
		NewSpan(start2, end2),
		NewSpan(start3, end3),
		NewSpan(start4, end4),
		NewSpan(start5, end5),
	}

	result := MergeAscending(input)

	assert.Len(t, result, 3)
	assert.Equal(t, start1, result[0].Start())
	assert.Equal(t, end1, result[0].End())
	assert.Equal(t, start2, result[1].Start())
	assert.Equal(t, end5, result[1].End())
	assert.Equal(t, start4, result[2].Start())
	assert.Equal(t, end4, result[2].End())
}

func TestMergeAscending_ReturnsTwoItems_GivenSecondItemAfterFirst(t *testing.T) {
	start1 := keys.MustNewDataStoreKey("/1/p/0/k1")
	end1 := keys.MustNewDataStoreKey("/1/p/0/k2")
	start2 := keys.MustNewDataStoreKey("/1/p/0/k4")
	end2 := keys.MustNewDataStoreKey("/1/p/0/k5")
	input := []Span{
		NewSpan(start1, end1),
		NewSpan(start2, end2),
	}

	result := MergeAscending(input)

	assert.Len(t, result, 2)
	assert.Equal(t, start1, result[0].Start())
	assert.Equal(t, end1, result[0].End())
	assert.Equal(t, start2, result[1].Start())
	assert.Equal(t, end2, result[1].End())
}

func TestMergeAscending_ReturnsSingle_GivenStartAdjacentAndBeforeEndEndEqual(t *testing.T) {
	start1 := keys.MustNewDataStoreKey("/1/p/0/k3")
	end1 := keys.MustNewDataStoreKey("/1/p/0/k6")
	start2 := keys.MustNewDataStoreKey("/1/p/0/k5")
	end2 := keys.MustNewDataStoreKey("/1/p/0/k6")
	input := []Span{
		NewSpan(start1, end1),
		NewSpan(start2, end2),
	}

	result := MergeAscending(input)

	assert.Len(t, result, 1)
	assert.Equal(t, start1, result[0].Start())
	assert.Equal(t, end2, result[0].End())
}

func TestMergeAscending_ReturnsSingle_GivenStartAdjacentAndBeforeEndEndAdjacentAndAfter(
	t *testing.T,
) {
	start1 := keys.MustNewDataStoreKey("/1/p/0/k3")
	end1 := keys.MustNewDataStoreKey("/1/p/0/k6")
	start2 := keys.MustNewDataStoreKey("/1/p/0/k5")
	end2 := keys.MustNewDataStoreKey("/1/p/0/k7")
	input := []Span{
		NewSpan(start1, end1),
		NewSpan(start2, end2),
	}

	result := MergeAscending(input)

	assert.Len(t, result, 1)
	assert.Equal(t, start1, result[0].Start())
	assert.Equal(t, end2, result[0].End())
}

func TestMergeAscending_ReturnsSingle_GivenStartAdjacentAndBeforeEndEndAfter(t *testing.T) {
	start1 := keys.MustNewDataStoreKey("/1/p/0/k3")
	end1 := keys.MustNewDataStoreKey("/1/p/0/k6")
	start2 := keys.MustNewDataStoreKey("/1/p/0/k5")
	end2 := keys.MustNewDataStoreKey("/1/p/0/k8")
	input := []Span{
		NewSpan(start1, end1),
		NewSpan(start2, end2),
	}

	result := MergeAscending(input)

	assert.Len(t, result, 1)
	assert.Equal(t, start1, result[0].Start())
	assert.Equal(t, end2, result[0].End())
}

func TestMergeAscending_ReturnsSingle_GivenStartAdjacentAndAfterEndEndAfter(t *testing.T) {
	start1 := keys.MustNewDataStoreKey("/1/p/0/k3")
	end1 := keys.MustNewDataStoreKey("/1/p/0/k6")
	start2 := keys.MustNewDataStoreKey("/1/p/0/k7")
	end2 := keys.MustNewDataStoreKey("/1/p/0/k8")
	input := []Span{
		NewSpan(start1, end1),
		NewSpan(start2, end2),
	}

	result := MergeAscending(input)

	assert.Len(t, result, 1)
	assert.Equal(t, start1, result[0].Start())
	assert.Equal(t, end2, result[0].End())
}
