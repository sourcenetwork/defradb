// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalByteSize(t *testing.T) {
	var bs ByteSize

	b := []byte("10")
	err := bs.UnmarshalText(b)
	assert.NoError(t, err)
	assert.Equal(t, 10*B, bs)

	b = []byte("10B")
	err = bs.UnmarshalText(b)
	assert.NoError(t, err)
	assert.Equal(t, 10*B, bs)

	b = []byte("10 B")
	err = bs.UnmarshalText(b)
	assert.NoError(t, err)
	assert.Equal(t, 10*B, bs)

	kb := []byte("10KB")
	err = bs.UnmarshalText(kb)
	assert.NoError(t, err)
	assert.Equal(t, 10*KiB, bs)

	kb = []byte("10KiB")
	err = bs.UnmarshalText(kb)
	assert.NoError(t, err)
	assert.Equal(t, 10*KiB, bs)

	kb = []byte("10 kb")
	err = bs.UnmarshalText(kb)
	assert.NoError(t, err)
	assert.Equal(t, 10*KiB, bs)

	mb := []byte("10MB")
	err = bs.UnmarshalText(mb)
	assert.NoError(t, err)
	assert.Equal(t, 10*MiB, bs)

	mb = []byte("10MiB")
	err = bs.UnmarshalText(mb)
	assert.NoError(t, err)
	assert.Equal(t, 10*MiB, bs)

	gb := []byte("10GB")
	err = bs.UnmarshalText(gb)
	assert.NoError(t, err)
	assert.Equal(t, 10*GiB, bs)

	gb = []byte("10GiB")
	err = bs.UnmarshalText(gb)
	assert.NoError(t, err)
	assert.Equal(t, 10*GiB, bs)

	tb := []byte("10TB")
	err = bs.UnmarshalText(tb)
	assert.NoError(t, err)
	assert.Equal(t, 10*TiB, bs)

	tb = []byte("10TiB")
	err = bs.UnmarshalText(tb)
	assert.NoError(t, err)
	assert.Equal(t, 10*TiB, bs)

	pb := []byte("10PB")
	err = bs.UnmarshalText(pb)
	assert.NoError(t, err)
	assert.Equal(t, 10*PiB, bs)

	pb = []byte("10PiB")
	err = bs.UnmarshalText(pb)
	assert.NoError(t, err)
	assert.Equal(t, 10*PiB, bs)

	eb := []byte("१")
	err = bs.UnmarshalText(eb)
	assert.ErrorIs(t, err, ErrUnableToParseByteSize)
}

func TestByteSizeType(t *testing.T) {
	var bs ByteSize
	assert.Equal(t, "ByteSize", bs.Type())
}

func TestByteSizeToString(t *testing.T) {
	b := 999 * B
	assert.Equal(t, "999", b.String())

	mb := 10 * MiB
	assert.Equal(t, "10MiB", mb.String())
}
