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
	Compare(Span) int
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

// Compare returns -1 if the provided span is less, 0 if it is equal, and 1 if its greater
func (s span) Compare(s2 Span) int {
	panic("not implemented") // TODO: Implement
}

// Spans is a collection of individual spans
type Spans []Span

// KeyValue is a KV store response containing the resulting ds.Key and byte array value
type KeyValue struct {
	Key   Key
	Value []byte
}
