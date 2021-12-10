// Copyright 2020 Source Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package core

import "strings"

// Span is a range of keys from [Start, End)
type Span interface {
	// Start returns the starting key of the Span
	Start() Key
	// End returns the ending key of the Span
	End() Key
	// Contains returns true of the Span contains the provided Span's range
	Contains(Span) bool
	// Equal returns true if the provided Span is equal to the current
	Equal(Span) bool
	// Compare returns -1 if the provided span is less, 0 if it is equal, and 1 if its greater
	Compare(Span) SpanComparisonResult
}

type span struct {
	start Key
	end   Key
}

func NewSpan(start, end Key) Span {
	return span{
		start: start,
		end:   end,
	}
}

// Start returns the starting key of the Span
func (s span) Start() Key {
	return s.start
}

// End returns the ending key of the Span
func (s span) End() Key {
	return s.end
}

// Contains returns true of the Span contains the provided Span's range
func (s span) Contains(s2 Span) bool {
	panic("not implemented") // TODO: Implement
}

// Equal returns true if the provided Span is equal to the current
func (s span) Equal(s2 Span) bool {
	panic("not implemented") // TODO: Implement
}

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
func (this span) Compare(other Span) SpanComparisonResult {
	if this == other {
		return Equal
	}

	thisStart := this.start.String()
	thisEnd := this.end.String()
	otherStart := other.Start().String()
	otherEnd := other.End().String()

	if thisStart < otherStart {
		if thisEnd == otherStart || isAdjacent(this.end, other.Start()) {
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

	if thisStart == otherEnd || isAdjacent(this.start, other.End()) {
		return StartEqualToEndEndAfter
	}

	return After
}

func isAdjacent(this Key, other Key) bool {
	return len(this.String()) == len(other.String()) && (this.PrefixEnd().String() == other.String() || this.String() == other.PrefixEnd().String())
}

// Spans is a collection of individual spans
type Spans []Span

// KeyValue is a KV store response containing the resulting ds.Key and byte array value
type KeyValue struct {
	Key   Key
	Value []byte
}

// Merges an unordered, potentially overlapping and/or duplicated collection of Spans into
// a unique set in ascending order, where overlapping spans are merged into a single span.
// Will handle spans with keys of different lengths, where one might be a prefix of another.
// Adjacent spans will also be merged.
func (spans Spans) MergeAscending() Spans {
	if len(spans) <= 1 {
		return spans
	}

	uniqueSpans := Spans{}

	for _, span := range spans {
		uniqueSpanFound := false

		i := 0
		for i < len(uniqueSpans) {
			uniqueSpan := uniqueSpans[i]
			switch span.Compare(uniqueSpan) {
			case Before:
				// Shift all remaining unique spans one place to the right
				newArray := make(Spans, len(uniqueSpans)+1)
				for j := len(uniqueSpans); j > i; j-- {
					newArray[j] = uniqueSpans[i]
				}

				// Then we insert
				newArray[i] = NewSpan(span.Start(), span.End())

				// Move the values prior to the new one across
				for j := 0; j < i; j++ {
					newArray[j] = uniqueSpans[j]
				}
				uniqueSpans = newArray
				uniqueSpanFound = true
				// Exit the unique-span loop, this span has been handled
				i = len(uniqueSpans)
			case StartBeforeEndEqualToStart, StartBeforeEndWithin, StartBeforeEndEqual:
				uniqueSpans[i] = NewSpan(span.Start(), uniqueSpan.End())
				uniqueSpanFound = true
				i++
			case StartBeforeEndAfter:
				uniqueSpans = uniqueSpans.removeBefore(i, span.End().String())
				uniqueSpans[i] = NewSpan(span.Start(), span.End())
				uniqueSpanFound = true
				// Exit the unique-span loop, this span has been handled
				i = len(uniqueSpans)
			case StartEqualEndWithin, Equal, StartWithinEndWithin, StartWithinEndEqual:
				uniqueSpanFound = true
				// Do nothing, span is contained within an existing unique-span
				i = len(uniqueSpans)
			case StartEqualEndAfter, StartWithinEndAfter, StartEqualToEndEndAfter:
				uniqueSpans = uniqueSpans.removeBefore(i, span.End().String())
				uniqueSpans[i] = NewSpan(uniqueSpan.Start(), span.End())
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
// than the given value.  The returned collection will be a different instance to the given
// and the given collection will not be mutated.
func (spans Spans) removeBefore(startIndex int, end string) Spans {
	indexOfLastMatchingItem := -1
	for i := startIndex; i < len(spans); i++ {
		if spans[i].End().String() <= end {
			indexOfLastMatchingItem = i
		}
	}

	if indexOfLastMatchingItem == -1 {
		return spans
	}

	numberOfItemsToRemove := indexOfLastMatchingItem - startIndex
	result := make(Spans, len(spans)-numberOfItemsToRemove)
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
