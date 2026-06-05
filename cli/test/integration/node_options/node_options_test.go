// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package node_options

import (
	"testing"
	"time"

	"github.com/sourcenetwork/defradb/cli/test/action"
	"github.com/sourcenetwork/defradb/cli/test/integration"
	"github.com/sourcenetwork/defradb/node"
)

// Both tests set each CLI-configurable node option and verify the returned node-options JSON.
// Two sets with different values confirm that results are not coincidentally equal to defaults.
//
// Note: DB.LensRuntime is always "wazero" here so as to allow the tests to run on Windows.
// Note also: TLS paths are not tested because we would need genuine TLS certs to test using them.

// testIdentityKeyHex is a valid secp256k1 private key used across CLI command examples.
const testIdentityKeyHex = "e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac"

// TestGetNodeOptions_SetA starts a node with set A values and verifies that all
// configurable options are reflected in `client node-options`.
func TestGetNodeOptions_SetA(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			action.StartWithArgs([]string{
				"--no-p2p",
				"--no-signing",
				"--pubsub=false",
				"--max-txn-retries=7",
				"--p2paddr=/ip4/127.0.0.1/tcp/9171",
				"--peers=/ip4/127.0.0.2/tcp/9172",
				"--allowed-origins=http://a.example.com",
				"--document-acp-type=none",
				"--replicator-retry-intervals=5,10",
				"--valuelogfilesize=1048576",
			}),
			&action.AssertNodeOptions{
				Expected: []action.NodeOptionExpected{
					// Node-level
					{Path: []string{"DisableP2P"}, Value: true},
					{Path: []string{"EnableDevelopment"}, Value: false},
					// Store
					// BadgerEncryptionKey is null because keyring is disabled
					{Path: []string{"Store", "Store"}, Value: ""},
					{Path: []string{"Store", "BadgerInMemory"}, Value: true},
					{Path: []string{"Store", "BadgerEncryptionKey"}, Value: nil},
					{Path: []string{"Store", "BadgerFileSize"}, Value: float64(1048576)},
					// DB
					{Path: []string{"DB", "EnableSigning"}, Value: false},
					{Path: []string{"DB", "Identity"}, Value: nil},
					{Path: []string{"DB", "MaxTxnRetries"}, Value: float64(7)},
					{Path: []string{"DB", "LensRuntime"}, Value: "wazero"},
					// ChunkSize is overridden to defaultChunkSize when BadgerInMemory is true
					{Path: []string{"DB", "ChunkSize"}, Value: float64(node.GetDefaultChunkSize())},
					{Path: []string{"DB", "RetryIntervals"}, Value: []any{float64(5000000000), float64(10000000000)}},
					{Path: []string{"DB", "SearchableEncryptionKey"}, Value: nil},
					// P2P
					// The following are stored but not used because p2p is disabled
					{Path: []string{"P2P", "EnablePubSub"}, Value: false},
					{Path: []string{"P2P", "EnableRelay"}, Value: false},
					{Path: []string{"P2P", "ListenAddresses"}, Value: []any{"/ip4/127.0.0.1/tcp/9171"}},
					{Path: []string{"P2P", "BootstrapPeers"}, Value: []any{"/ip4/127.0.0.2/tcp/9172"}},
					{Path: []string{"P2P", "PrivateKey"}, Value: nil},
					// HTTP
					{Path: []string{"HTTP", "AllowedOrigins"}, Value: []any{"http://a.example.com"}},
					// Document ACP
					{Path: []string{"DocumentACP", "DocumentACPType"}, Value: "none"},
					{Path: []string{"DocumentACP", "Path"}, Value: ""},
					// Node ACP
					{Path: []string{"NodeACP", "IsEnabled"}, Value: false},
					{Path: []string{"NodeACP", "Path"}, Value: ""},
				},
			},
		},
	}

	test.Execute(t)
}

// TestGetNodeOptions_SetB starts a node with set B values (different from set A on every
// field that the CLI can control) and verifies the returned options.
func TestGetNodeOptions_SetB(t *testing.T) {
	// The file keyring backend requires a secret
	t.Setenv("DEFRA_KEYRING_SECRET", "test-secret-for-node-options-setb")

	test := &integration.Test{
		// Disk-mode + P2P startup is slower than in-memory; allow more time.
		Timeout: 10 * time.Second,
		Actions: []action.Action{
			action.StartDiskWithArgs([]string{
				"--no-keyring=false",
				"--development",
				"--relay",
				"--max-txn-retries=14",
				"--p2paddr=/ip4/127.0.0.1/tcp/9173",
				// No --peers: bootstrap peers need a full multiaddr with peer ID to be valid
				// when P2P is running. Set A stores a peer (P2P disabled) but Set B cannot.
				// Empty vs non-empty still differentiates BootstrapPeers between the two sets.
				"--allowed-origins=http://b.example.com",
				"--replicator-retry-intervals=15,30",
				"--node-acp-enable",
				"--identity=" + testIdentityKeyHex,
				"--valuelogfilesize=2097152",
			}),
			&action.AssertNodeOptions{
				Expected: []action.NodeOptionExpected{
					// Node-level
					{Path: []string{"DisableP2P"}, Value: false},
					{Path: []string{"EnableDevelopment"}, Value: true},
					// Store
					{Path: []string{"Store", "Store"}, Value: ""},
					{Path: []string{"Store", "BadgerInMemory"}, Value: false},
					{Path: []string{"Store", "BadgerEncryptionKey"}, Value: "<redacted>"},
					{Path: []string{"Store", "BadgerFileSize"}, Value: float64(2097152)},
					// DB
					{Path: []string{"DB", "EnableSigning"}, Value: true},
					{Path: []string{"DB", "Identity"}, Value: "<redacted>"},
					{Path: []string{"DB", "MaxTxnRetries"}, Value: float64(14)},
					{Path: []string{"DB", "LensRuntime"}, Value: "wazero"},
					{Path: []string{"DB", "ChunkSize"}, Value: nil},
					{Path: []string{"DB", "RetryIntervals"}, Value: []any{float64(15000000000), float64(30000000000)}},
					{Path: []string{"DB", "SearchableEncryptionKey"}, Value: "<redacted>"},
					// P2P
					{Path: []string{"P2P", "EnablePubSub"}, Value: true},
					{Path: []string{"P2P", "EnableRelay"}, Value: true},
					{Path: []string{"P2P", "ListenAddresses"}, Value: []any{"/ip4/127.0.0.1/tcp/9173"}},
					{Path: []string{"P2P", "BootstrapPeers"}, Value: []any{}},
					{Path: []string{"P2P", "PrivateKey"}, Value: "<redacted>"},
					// HTTP
					{Path: []string{"HTTP", "AllowedOrigins"}, Value: []any{"http://b.example.com"}},
					// Document ACP
					{Path: []string{"DocumentACP", "DocumentACPType"}, Value: "local"},
					{Path: []string{"DocumentACP", "Path"}, UseRootDir: true},
					// Node ACP
					{Path: []string{"NodeACP", "IsEnabled"}, Value: true},
					{Path: []string{"NodeACP", "Path"}, UseRootDir: true},
				},
			},
		},
	}

	test.Execute(t)
}
