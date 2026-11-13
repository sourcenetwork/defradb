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

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
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

// TestP2PResourceManager_AllowlistBeyondDefaultCap is the k8s setup: only the pod network is
// allowlisted and max_conns_per_ip is left at libp2p's default of 8. Distinct peers behind the
// shared NAT address fill its 8 connections and keep being admitted past them, while a public
// address that is not allowlisted is still refused at its 9th connection at the same time.
func TestP2PResourceManager_AllowlistBeyondDefaultCap(t *testing.T) {
	const (
		defaultCap   = 8
		natAddr      = "/ip4/10.244.1.75/tcp/4001" // in-cluster SNAT address, allowlisted
		publicAddr   = "/ip4/203.0.113.7/tcp/4001" // public host, not allowlisted
		clusterConns = 3 * defaultCap
	)

	rm, err := p2pResourceManager(options.NodeP2POptions{AllowlistCIDRs: []string{"10.244.0.0/16"}})
	require.NoError(t, err)
	require.NotNil(t, rm)
	t.Cleanup(func() { _ = rm.Close() })

	// open accepts an inbound connection and completes the handshake as a fresh peer, as each
	// in-cluster pod would. SetPeer moves the connection out of the transient scope, which only
	// holds connections still mid-handshake (and caps allowlisted ones at 16 when autoscaled).
	open := func(addr string) bool {
		scope, err := rm.OpenConnection(network.DirInbound, true, ma.StringCast(addr))
		if err != nil {
			return false
		}
		t.Cleanup(scope.Done)
		priv, _, err := crypto.GenerateEd25519Key(nil)
		require.NoError(t, err)
		id, err := peer.IDFromPrivateKey(priv)
		require.NoError(t, err)
		return scope.SetPeer(id) == nil
	}

	// The allowlisted address fills the default cap first...
	for i := range defaultCap {
		require.True(t, open(natAddr), "cluster conn %d of the first %d", i+1, defaultCap)
	}
	// ...and is still admitted well past it while those 8 stay open.
	for i := defaultCap; i < clusterConns; i++ {
		require.True(t, open(natAddr), "cluster conn %d, beyond the default cap of %d", i+1, defaultCap)
	}

	// Meanwhile a non-allowlisted address keeps libp2p's default cap.
	for i := range defaultCap {
		require.True(t, open(publicAddr), "public conn %d", i+1)
	}
	assert.False(t, open(publicAddr), "public conn %d must be refused", defaultCap+1)

	// And the allowlisted address is unaffected by the public one being at its cap.
	assert.True(t, open(natAddr), "cluster conn %d after public cap reached", clusterConns+1)
}

// TestP2PResourceManager_AllowlistExemptFromConnRateLimit churns connections (open, then close)
// so the concurrent cap never binds and only libp2p's new-connection rate limiter can refuse
// them. That limiter allows a burst of 16 per IPv4 address and runs before the allowlist check,
// so allowlisted addresses must be exempted from it explicitly, e.g. for a reconnect storm after
// a restart, when every in-cluster peer redials through the same NAT address at once.
func TestP2PResourceManager_AllowlistExemptFromConnRateLimit(t *testing.T) {
	const (
		natAddr    = "/ip4/10.244.1.75/tcp/4001"
		publicAddr = "/ip4/203.0.113.7/tcp/4001"
		burst      = 2 * libp2pDefaultMaxConnsPerIP
	)

	rm, err := p2pResourceManager(options.NodeP2POptions{AllowlistCIDRs: []string{"10.244.0.0/16"}})
	require.NoError(t, err)
	t.Cleanup(func() { _ = rm.Close() })

	churn := func(addr string) bool {
		scope, err := rm.OpenConnection(network.DirInbound, true, ma.StringCast(addr))
		if err != nil {
			return false
		}
		scope.Done()
		return true
	}

	for i := range 5 * burst {
		require.True(t, churn(natAddr), "allowlisted reconnect %d", i+1)
	}

	for i := range burst {
		require.True(t, churn(publicAddr), "public reconnect %d", i+1)
	}
	assert.False(t, churn(publicAddr), "public reconnect %d must be rate limited", burst+1)
}
