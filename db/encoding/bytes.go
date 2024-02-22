// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package encoding

import (
	"bytes"
)

const (
	// All terminators are encoded as \x00\x01 sequence.
	// In order to distinguish \x00 byte it is escaped as \x00\xff
	escape          byte = 0x00
	escapedTerm     byte = 0x01
	escaped00       byte = 0xff
	escapedFF       byte = 0x00
	escapeDesc      byte = ^escape
	escapedTermDesc byte = ^escapedTerm
	escaped00Desc   byte = ^escaped00
	escapedFFDesc   byte = ^escapedFF
)

type escapes struct {
	escape      byte
	escapedTerm byte
	escaped00   byte
	escapedFF   byte
	marker      byte
}

var (
	ascendingBytesEscapes  = escapes{escape, escapedTerm, escaped00, escapedFF, bytesMarker}
	descendingBytesEscapes = escapes{escapeDesc, escapedTermDesc, escaped00Desc, escapedFFDesc, bytesDescMarker}
)

// EncodeBytesAscending encodes the []byte value using an escape-based
// encoding. The encoded value is terminated with the sequence
// "\x00\x01" which is guaranteed to not occur elsewhere in the
// encoded value. The encoded bytes are append to the supplied buffer
// and the resulting buffer is returned.
func EncodeBytesAscending(b []byte, data []byte) []byte {
	return encodeBytesAscendingWithTerminatorAndPrefix(b, data, ascendingBytesEscapes.escapedTerm, bytesMarker)
}

// encodeBytesAscendingWithTerminatorAndPrefix encodes the []byte value using an escape-based
// encoding. The encoded value is terminated with the sequence
// "\x00\terminator". The encoded bytes are append to the supplied buffer
// and the resulting buffer is returned. The terminator allows us to pass
// different terminators for things such as JSON key encoding.
func encodeBytesAscendingWithTerminatorAndPrefix(
	b []byte, data []byte, terminator byte, prefix byte,
) []byte {
	b = append(b, prefix)
	return encodeBytesAscendingWithTerminator(b, data, terminator)
}

// encodeBytesAscendingWithTerminator encodes the []byte value using an escape-based
// encoding. The encoded value is terminated with the sequence
// "\x00\terminator". The encoded bytes are append to the supplied buffer
// and the resulting buffer is returned. The terminator allows us to pass
// different terminators for things such as JSON key encoding.
func encodeBytesAscendingWithTerminator(b []byte, data []byte, terminator byte) []byte {
	bs := encodeBytesAscendingWithoutTerminatorOrPrefix(b, data)
	return append(bs, escape, terminator)
}

// encodeBytesAscendingWithoutTerminatorOrPrefix encodes the []byte value using an escape-based
// encoding.
func encodeBytesAscendingWithoutTerminatorOrPrefix(b []byte, data []byte) []byte {
	for {
		// IndexByte is implemented by the go runtime in assembly and is
		// much faster than looping over the bytes in the slice.
		i := bytes.IndexByte(data, escape)
		if i == -1 {
			break
		}
		b = append(b, data[:i]...)
		b = append(b, escape, escaped00)
		data = data[i+1:]
	}
	return append(b, data...)
}

// EncodeBytesDescending encodes the []byte value using an
// escape-based encoding and then inverts (ones complement) the result
// so that it sorts in reverse order, from larger to smaller
// lexicographically.
func EncodeBytesDescending(b []byte, data []byte) []byte {
	n := len(b)
	b = EncodeBytesAscending(b, data)
	b[n] = bytesDescMarker
	onesComplement(b[n+1:])
	return b
}

// DecodeBytesAscending decodes a []byte value from the input buffer
// which was encoded using EncodeBytesAscending. The decoded bytes
// are appended to r. The remainder of the input buffer and the
// decoded []byte are returned.
func DecodeBytesAscending(b []byte, r []byte) ([]byte, []byte, error) {
	return decodeBytesInternal(b, r, ascendingBytesEscapes, true /* expectMarker */)
}

// DecodeBytesDescending decodes a []byte value from the input buffer
// which was encoded using EncodeBytesDescending. The decoded bytes
// are appended to r. The remainder of the input buffer and the
// decoded []byte are returned.
func DecodeBytesDescending(b []byte, r []byte) ([]byte, []byte, error) {
	// Always pass an `r` to make sure we never get back a sub-slice of `b`,
	// since we're going to modify the contents of the slice.
	if r == nil {
		r = []byte{}
	}
	b, r, err := decodeBytesInternal(b, r, descendingBytesEscapes, true /* expectMarker */)
	onesComplement(r)
	return b, r, err
}

func decodeBytesInternal(b []byte, r []byte, e escapes, expectMarker bool) ([]byte, []byte, error) {
	if expectMarker {
		if len(b) == 0 || b[0] != e.marker {
			return nil, nil, NewErrMarkersNotFound(b, e.marker)
		}
		b = b[1:]
	}

	for {
		i := bytes.IndexByte(b, e.escape)
		if i == -1 {
			return nil, nil, NewErrTerminatorNotFound(b, e.escape)
		}
		if i+1 >= len(b) {
			return nil, nil, NewErrMalformedEscape(b)
		}
		v := b[i+1]
		if v == e.escapedTerm {
			if r == nil {
				r = b[:i]
			} else {
				r = append(r, b[:i]...)
			}
			return b[i+2:], r, nil
		}

		if v != e.escaped00 {
			return nil, nil, NewErrUnknownEscapeSequence(b[i:i+2], e.escape)
		}

		r = append(r, b[:i]...)
		r = append(r, e.escapedFF)
		b = b[i+2:]
	}
}
