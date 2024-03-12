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
	"github.com/sourcenetwork/defradb/errors"
)

const (
	errInsufficientBytesToDecode = "insufficient bytes to decode buffer into a target type"
	errCanNotDecodeFieldValue    = "can not decode field value"
	errMarkersNotFound           = "did not find any of required markers in buffer"
	errTerminatorNotFound        = "did not find required terminator in buffer"
	errMalformedEscape           = "malformed escape in buffer"
	errUnknownEscapeSequence     = "unknown escape sequence"
	errInvalidUvarintLength      = "invalid length for uvarint"
	errVarintOverflow            = "varint overflows a 64-bit integer"
)

var (
	ErrInsufficientBytesToDecode = errors.New(errInsufficientBytesToDecode)
	ErrCanNotDecodeFieldValue    = errors.New(errCanNotDecodeFieldValue)
	ErrMarkersNotFound           = errors.New(errMarkersNotFound)
	ErrTerminatorNotFound        = errors.New(errTerminatorNotFound)
	ErrMalformedEscape           = errors.New(errMalformedEscape)
	ErrUnknownEscapeSequence     = errors.New(errUnknownEscapeSequence)
	ErrInvalidUvarintLength      = errors.New(errInvalidUvarintLength)
	ErrVarintOverflow            = errors.New(errVarintOverflow)
)

// NewErrInsufficientBytesToDecode returns a new error indicating that the provided
// bytes are not sufficient to decode into a target type.
func NewErrInsufficientBytesToDecode(b []byte, decodeTarget string) error {
	return errors.New(errInsufficientBytesToDecode,
		errors.NewKV("Buffer", b), errors.NewKV("Decode Target", decodeTarget))
}

// NewErrCanNotDecodeFieldValue returns a new error indicating that the encoded
// bytes could not be decoded.
func NewErrCanNotDecodeFieldValue(b []byte, innerErr ...error) error {
	kvs := []errors.KV{errors.NewKV("Buffer", b)}
	if len(innerErr) > 0 {
		kvs = append(kvs, errors.NewKV("InnerErr", innerErr[0]))
	}
	return errors.New(errCanNotDecodeFieldValue, kvs...)
}

// NewErrMarkersNotFound returns a new error indicating that the required
// marker was not found in the buffer.
func NewErrMarkersNotFound(b []byte, markers ...byte) error {
	return errors.New(errMarkersNotFound, errors.NewKV("Markers", markers), errors.NewKV("Buffer", b))
}

// NewErrTerminatorNotFound returns a new error indicating that the required
// terminator was not found in the buffer.
func NewErrTerminatorNotFound(b []byte, terminator byte) error {
	return errors.New(errTerminatorNotFound, errors.NewKV("Terminator", terminator), errors.NewKV("Buffer", b))
}

// NewErrMalformedEscape returns a new error indicating that the buffer
// contains a malformed escape sequence.
func NewErrMalformedEscape(b []byte) error {
	return errors.New(errMalformedEscape, errors.NewKV("Buffer", b))
}

// NewErrUnknownEscapeSequence returns a new error indicating that the buffer
// contains an unknown escape sequence.
func NewErrUnknownEscapeSequence(b []byte, escape byte) error {
	return errors.New(errUnknownEscapeSequence, errors.NewKV("Escape", escape), errors.NewKV("Buffer", b))
}

// NewErrInvalidUvarintLength returns a new error indicating that the buffer
// contains an invalid length for a uvarint.
func NewErrInvalidUvarintLength(b []byte, length int) error {
	return errors.New(errInvalidUvarintLength, errors.NewKV("Buffer", b), errors.NewKV("Length", length))
}

// NewErrVarintOverflow returns a new error indicating that the buffer
// contains a varint that overflows a 64-bit integer.
func NewErrVarintOverflow(b []byte, value uint64) error {
	return errors.New(errVarintOverflow, errors.NewKV("Buffer", b), errors.NewKV("Value", value))
}
