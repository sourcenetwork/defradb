// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package datastore

import (
	"context"

	"github.com/sourcenetwork/corekv"
)

// Key represents a typed key for a key-value pair or a prefix within a `Keyedstore`.
type Key interface {
	// Bytes returns the serialized for of the key, as it is stored within a `Keyedstore`.
	Bytes() []byte
}

// Keyedstore is a corekv.ReaderWriter that takes typed `Key`s instead of `[]byte`
// for it's key function parameters.
type Keyedstore interface {
	// Get returns the value at the given key.
	//
	// If no item with the given key is found, nil, and an [ErrNotFound]
	// error will be returned.
	Get(ctx context.Context, key Key) ([]byte, error)

	// Has returns true if an item at the given key is found, otherwise
	// will return false.
	Has(ctx context.Context, key Key) (bool, error)

	// Iterator returns a read-only iterator using the given options.
	Iterator(ctx context.Context, opts IterOptions) (corekv.Iterator, error)

	// Set sets the value stored against the given key.
	//
	// If an item already exists at the given key it will be overwritten.
	Set(ctx context.Context, key Key, value []byte) error

	// Delete removes the value at the given key.
	//
	// If no matching key is found the behaviour is undefined:
	// https://github.com/sourcenetwork/corekv/issues/36
	Delete(ctx context.Context, key Key) error
}

// IterOptions contains the full set of available iterator options,
// it can be provided when creating an [Iterator] from a [Store].
type IterOptions struct {
	// Prefix iteration, only keys beginning with the designated prefix
	// with the given prefix will be yielded.
	//
	// Keys exactly matching the provided `Prefix` value will not be
	// yielded.
	//
	// Providing a Prefix value should cause the Start and End options
	// to be ignored, although this is currently untested:
	// https://github.com/sourcenetwork/corekv/issues/35
	Prefix Key

	// If Prefix is nil, and Start is provided, the iterator will
	// only yield items with a key lexographically greater than or
	// equal to this value.
	//
	// Providing an `End` value equal to or smaller than this value
	// will result in undefined behaviour:
	// https://github.com/sourcenetwork/corekv/issues/32
	Start Key

	// If Prefix is nil, and End is provided, the iterator will
	// only yield items with a key lexographically smaller than this
	// value.
	//
	// Providing an End value equal to or smaller than Start
	// will result in undefined behaviour:
	// https://github.com/sourcenetwork/corekv/issues/32
	End Key

	// Reverse the direction of the iteration, returning items in
	// lexographically descending order of their keys.
	Reverse bool

	// Only iterate through keys. Calling Value on the
	// iterator will return nil and no error.
	//
	// This option is currently untested:
	// https://github.com/sourcenetwork/corekv/issues/34
	//
	// It is very likely ignored for the memory store iteration:
	// https://github.com/sourcenetwork/corekv/issues/33
	KeysOnly bool
}
