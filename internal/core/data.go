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
	"strings"

	"github.com/sourcenetwork/defradb/internal/keys"
)

// Span is a range of keys from [Start, End).
type Span struct {
	// Start represents the starting key of the Span.
	Start keys.DataStoreKey

	// End represents the ending key of the Span.
	End keys.DataStoreKey
}

// NewSpan creates a new Span from the provided start and end keys.
func NewSpan(start, end keys.DataStoreKey) Span {
	return Span{
		Start: start,
		End:   end,
	}
}

// SpanComparisonResult is the result of comparing two spans.
type SpanComparisonResult uint

const (
	Before SpanComparisonResult = iota
	StartBeforeEndEqualToStart
	StartBeforeEndWithin
	StartBeforeEndEqual
	StartBeforeEndAfter
	StartEqualEndWithin
	Equal
	StartEqualEndAfter
	StartWithinEndWithin
	StartWithinEndAfter
	StartWithinEndEqual
	StartEqualToEndEndAfter
	After
)

// Compares two spans returning how the compare to each other.
// If the end of one span is adjacent to the other (with no gap possible)
// then those ends are considered equal.
func (this Span) Compare(other Span) SpanComparisonResult {
	if this == other {
		return Equal
	}

	thisStart := this.Start.ToString()
	thisEnd := this.End.ToString()
	otherStart := other.Start.ToString()
	otherEnd := other.End.ToString()

	if thisStart < otherStart {
		if thisEnd == otherStart || isAdjacent(this.End, other.Start) {
			return StartBeforeEndEqualToStart
		}

		if thisEnd < otherStart {
			return Before
		}

		if thisEnd < otherEnd || strings.HasPrefix(thisEnd, otherEnd) {
			return StartBeforeEndWithin
		}

		if thisEnd == otherEnd {
			return StartBeforeEndEqual
		}

		if thisEnd > otherEnd {
			return StartBeforeEndAfter
		}
	}

	if thisStart == otherStart {
		if thisEnd < otherEnd || strings.HasPrefix(thisEnd, otherEnd) {
			return StartEqualEndWithin
		}

		if thisEnd == otherEnd {
			return Equal
		}

		if thisEnd > otherEnd {
			return StartEqualEndAfter
		}
	}

	if thisStart < otherEnd {
		if thisEnd < otherEnd || strings.HasPrefix(thisEnd, otherEnd) {
			return StartWithinEndWithin
		}

		if thisEnd == otherEnd {
			return StartWithinEndEqual
		}

		if thisEnd > otherEnd {
			return StartWithinEndAfter
		}
	}

	if thisStart == otherEnd || isAdjacent(this.Start, other.End) {
		return StartEqualToEndEndAfter
	}

	return After
}

func isAdjacent(this keys.DataStoreKey, other keys.DataStoreKey) bool {
	return len(this.ToString()) == len(other.ToString()) &&
		(this.PrefixEnd().ToString() == other.ToString() ||
			this.ToString() == other.PrefixEnd().ToString())
}

// Merges an unordered, potentially overlapping and/or duplicated collection of Spans into
// a unique set in ascending order, where overlapping spans are merged into a single span.
// Will handle spans with keys of different lengths, where one might be a prefix of another.
// Adjacent spans will also be merged.
func MergeAscending(spans []Span) []Span {
	if len(spans) <= 1 {
		return spans
	}

	uniqueSpans := []Span{}

	for _, span := range spans {
		uniqueSpanFound := false

		i := 0
		for i < len(uniqueSpans) {
			uniqueSpan := uniqueSpans[i]
			switch span.Compare(uniqueSpan) {
			case Before:
				// Shift all remaining unique spans one place to the right
				newArray := make([]Span, len(uniqueSpans)+1)
				for j := len(uniqueSpans); j > i; j-- {
					newArray[j] = uniqueSpans[i]
				}

				// Then we insert
				newArray[i] = NewSpan(span.Start, span.End)

				// Move the values prior to the new one across
				for j := 0; j < i; j++ {
					newArray[j] = uniqueSpans[j]
				}
				uniqueSpans = newArray
				uniqueSpanFound = true
				// Exit the unique-span loop, this span has been handled
				i = len(uniqueSpans)
			case StartBeforeEndEqualToStart, StartBeforeEndWithin, StartBeforeEndEqual:
				uniqueSpans[i] = NewSpan(span.Start, uniqueSpan.End)
				uniqueSpanFound = true
				i++
			case StartBeforeEndAfter:
				uniqueSpans = removeBefore(uniqueSpans, i, span.End.ToString())
				uniqueSpans[i] = NewSpan(span.Start, span.End)
				uniqueSpanFound = true
				// Exit the unique-span loop, this span has been handled
				i = len(uniqueSpans)
			case StartEqualEndWithin, Equal, StartWithinEndWithin, StartWithinEndEqual:
				uniqueSpanFound = true
				// Do nothing, span is contained within an existing unique-span
				i = len(uniqueSpans)
			case StartEqualEndAfter, StartWithinEndAfter, StartEqualToEndEndAfter:
				uniqueSpans = removeBefore(uniqueSpans, i, span.End.ToString())
				uniqueSpans[i] = NewSpan(uniqueSpan.Start, span.End)
				uniqueSpanFound = true
				// Exit the unique-span loop, this span has been handled
				i = len(uniqueSpans)
			case After:
				i++
			}
		}

		if !uniqueSpanFound {
			uniqueSpans = append(uniqueSpans, span)
		}
	}

	return uniqueSpans
}

// Removes any items from the collection (given index onwards) who's end key is smaller
// than the given value. The returned collection will be a different instance.
func removeBefore(spans []Span, startIndex int, end string) []Span {
	indexOfLastMatchingItem := -1
	for i := startIndex; i < len(spans); i++ {
		if spans[i].End.ToString() <= end {
			indexOfLastMatchingItem = i
		}
	}

	if indexOfLastMatchingItem == -1 {
		return spans
	}

	numberOfItemsToRemove := indexOfLastMatchingItem - startIndex
	result := make([]Span, len(spans)-numberOfItemsToRemove)
	// Add the items preceding the removed items
	for i := 0; i < startIndex; i++ {
		result[i] = spans[i]
	}

	j := startIndex + numberOfItemsToRemove
	// Add the items following the removed items
	for i := indexOfLastMatchingItem + 1; i < len(spans); i++ {
		result[j] = spans[i]
	}

	return result
}
