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
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/config"
)

func TestNewMergedOptionsSimple(t *testing.T) {
	opt, err := NewMergedOptions()
	require.NoError(t, err)
	require.NotNil(t, opt)
}

func TestNewMergedOptionsWithNilOption(t *testing.T) {
	opt, err := NewMergedOptions(nil)
	require.NoError(t, err)
	require.NotNil(t, opt)
}

func TestNewConnManagerSimple(t *testing.T) {
	conMngr, err := NewConnManager(1, 10, time.Second)
	require.NoError(t, err)
	err = conMngr.Close()
	require.NoError(t, err)
}

func TestNewConnManagerWithError(t *testing.T) {
	_, err := NewConnManager(1, 10, -time.Second)
	require.Contains(t, err.Error(), "grace period must be non-negative")
}

func TestWithConfigWithP2PAddressError(t *testing.T) {
	cfg := config.Config{
		Net: &config.NetConfig{
			P2PAddress: "/willerror/0.0.0.0/tcp/9999",
		},
	}
	err := WithConfig(&cfg)(&Options{})
	require.Contains(t, err.Error(), "failed to parse multiaddr")
}

func TestWithPrivateKey(t *testing.T) {
	key, _, err := crypto.GenerateKeyPair(crypto.Ed25519, 0)
	require.NoError(t, err)

	opt, err := NewMergedOptions(WithPrivateKey(key))
	require.NoError(t, err)
	require.NotNil(t, opt)
	require.Equal(t, key, opt.PrivateKey)
}

func TestWithPubSub(t *testing.T) {
	opt, err := NewMergedOptions(WithPubSub(true))
	require.NoError(t, err)
	require.NotNil(t, opt)
	require.True(t, opt.EnablePubSub)
}

func TestWithEnableRelay(t *testing.T) {
	opt, err := NewMergedOptions(WithEnableRelay(true))
	require.NoError(t, err)
	require.NotNil(t, opt)
	require.True(t, opt.EnableRelay)
}

func TestWithListenP2PAddrStringsWithError(t *testing.T) {
	addr := "/willerror/0.0.0.0/tcp/9999"
	_, err := NewMergedOptions(WithListenP2PAddrStrings(addr))
	require.Contains(t, err.Error(), "failed to parse multiaddr")
}

func TestWithListenP2PAddrStrings(t *testing.T) {
	addr := "/ip4/0.0.0.0/tcp/9999"
	opt, err := NewMergedOptions(WithListenP2PAddrStrings(addr))
	require.NoError(t, err)
	require.NotNil(t, opt)
	require.Equal(t, addr, opt.ListenAddrs[0].String())
}

func TestWithListenAddrs(t *testing.T) {
	addr := "/ip4/0.0.0.0/tcp/9999"
	a, err := ma.NewMultiaddr(addr)
	require.NoError(t, err)

	opt, err := NewMergedOptions(WithListenAddrs(a))
	require.NoError(t, err)
	require.NotNil(t, opt)
	require.Equal(t, addr, opt.ListenAddrs[0].String())
}
