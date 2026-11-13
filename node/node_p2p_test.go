// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build !js

package node

import (
	"testing"

	"github.com/libp2p/go-libp2p/core/network"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client/options"
)

func TestCIDRsToMultiaddrs(t *testing.T) {
	got, err := cidrsToMultiaddrs([]string{
		"10.244.0.0/16",
		"10.244.1.75/32",
		"10.244.1.75/16", // host bits are masked off
		"fd00::/8",
		"2001:db8::1/128",
	})
	require.NoError(t, err)

	want := []string{
		"/ip4/10.244.0.0/ipcidr/16",
		"/ip4/10.244.1.75/ipcidr/32",
		"/ip4/10.244.0.0/ipcidr/16",
		"/ip6/fd00::/ipcidr/8",
		"/ip6/2001:db8::1/ipcidr/128",
	}
	require.Len(t, got, len(want))
	for i := range want {
		assert.Equal(t, want[i], got[i].String())
	}
}

func TestCIDRsToMultiaddrs_Invalid(t *testing.T) {
	for _, cidr := range []string{"", "10.244.0.0", "10.244.0.0/33", "not-a-cidr", "fd00::/129"} {
		_, err := cidrsToMultiaddrs([]string{"10.0.0.0/8", cidr})
		require.Error(t, err, "cidr %q", cidr)
		assert.ErrorContains(t, err, errInvalidAllowlistCIDR)
		assert.ErrorContains(t, err, cidr)
	}
}

func TestP2PResourceManager_UnsetKeepsLibp2pDefaults(t *testing.T) {
	rm, err := p2pResourceManager(options.NodeP2POptions{})
	require.NoError(t, err)
	assert.Nil(t, rm)
}

func TestP2PResourceManager_InvalidAllowlistFailsStartup(t *testing.T) {
	_, err := p2pResourceManager(options.NodeP2POptions{AllowlistCIDRs: []string{"10.244.0.0"}})
	assert.ErrorContains(t, err, errInvalidAllowlistCIDR)
}

// TestP2PResourceManager_ConnsPerIP opens inbound connections through the resource manager the
// node builds, which is the check the TCP listener runs on accept. It uses the source-NAT address
// of the incident rather than loopback, because libp2p exempts 127.0.0.0/8 and ::1 from the
// per-IP limit by default, so a loopback dial can never hit the cap.
func TestP2PResourceManager_ConnsPerIP(t *testing.T) {
	const natAddr = "/ip4/10.244.1.75/tcp/4001"
	const natAddr6 = "/ip6/fd00:10:244::1/tcp/4001"

	tests := []struct {
		name string
		opts options.NodeP2POptions
		addr string
		want []bool
	}{
		{
			name: "default cap is libp2p's 8",
			opts: options.NodeP2POptions{MaxStreamsPerPeer: 1},
			addr: natAddr,
			want: []bool{true, true, true, true, true, true, true, true, false},
		},
		{
			name: "third connection from same IPv4 is refused",
			opts: options.NodeP2POptions{MaxConnsPerIP: 2},
			addr: natAddr,
			want: []bool{true, true, false},
		},
		{
			name: "third connection from same IPv6 /56 is refused",
			opts: options.NodeP2POptions{MaxConnsPerIP: 2},
			addr: natAddr6,
			want: []bool{true, true, false},
		},
		{
			name: "allowlisted CIDR bypasses the cap",
			opts: options.NodeP2POptions{MaxConnsPerIP: 2, AllowlistCIDRs: []string{"10.244.0.0/16"}},
			addr: natAddr,
			want: []bool{true, true, true, true, true, true, true, true, true, true},
		},
		{
			name: "allowlist does not cover other networks",
			opts: options.NodeP2POptions{MaxConnsPerIP: 2, AllowlistCIDRs: []string{"10.245.0.0/16"}},
			addr: natAddr,
			want: []bool{true, true, false},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rm, err := p2pResourceManager(tc.opts)
			require.NoError(t, err)
			require.NotNil(t, rm)
			t.Cleanup(func() { _ = rm.Close() })

			endpoint := ma.StringCast(tc.addr)
			got := make([]bool, len(tc.want))
			for i := range got {
				scope, err := rm.OpenConnection(network.DirInbound, true, endpoint)
				got[i] = err == nil
				if err == nil {
					t.Cleanup(scope.Done)
				}
			}
			assert.Equal(t, tc.want, got)
		})
	}
}
