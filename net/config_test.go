// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package net

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithListenAddresses(t *testing.T) {
	opts := &Options{}
	addresses := []string{"/ip4/127.0.0.1/tcp/6666", "/ip4/0.0.0.0/tcp/6666"}
	WithListenAddresses(addresses...)(opts)
	assert.Equal(t, addresses, opts.ListenAddresses)
}

func TestWithEnableRelay(t *testing.T) {
	opts := &Options{}
	WithEnableRelay(true)(opts)
	assert.Equal(t, true, opts.EnableRelay)
}

func TestWithEnablePubSub(t *testing.T) {
	opts := &Options{}
	WithEnablePubSub(true)(opts)
	assert.Equal(t, true, opts.EnablePubSub)
}
