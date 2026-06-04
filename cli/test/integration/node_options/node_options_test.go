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
)

// Both tests set each CLI-configurable node option and verify the returned node-options JSON.
// Two sets with different values confirm that results are not coincidentally equal to defaults.
//
// Set A: in-memory storage, P2P disabled, no keyring.
// Set B: disk storage, P2P enabled, keyring enabled, development mode, NAC enabled.
//
// Two fields cannot be differentiated due to CLI/platform constraints:
//   - DB.LensRuntime: always "wazero" (DEFRA_LENS_RUNTIME env var is required on Windows)
//   - Store.Store: always "" (start.go derives BadgerInMemory from the store flag but never
//     calls opts.Store().SetType(), so the type field is never set in the options)

// testIdentityKeyHex is a valid secp256k1 private key used across CLI command examples.
const testIdentityKeyHex = "e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac"

// TestGetNodeOptions_SetA starts a node with set A values and verifies that all
// configurable options are reflected in `client node-options`.
func TestGetNodeOptions_SetA(t *testing.T) {
	// defaultChunkSize is what node.Start() applies whenever BadgerInMemory is true.
	const defaultChunkSize = float64(1048575)

	// P2P.EnablePubSub: make it false for set A (default is true from config).
	// Viper reads DEFRA_NET_PUBSUBENABLED via AutomaticEnv.
	t.Setenv("DEFRA_NET_PUBSUBENABLED", "false")

	test := &integration.Test{
		Actions: []action.Action{
			action.StartWithArgs([]string{
				"--no-p2p",
				"--no-signing",
				"--max-txn-retries=7",
				"--p2paddr=/ip4/127.0.0.1/tcp/9171",
				"--peers=/ip4/127.0.0.2/tcp/9172",
				"--allowed-origins=http://a.example.com",
				// Overrides the --document-acp-type=local set by StartWithArgs base args.
				"--document-acp-type=none",
				"--replicator-retry-intervals=5,10",
			}),
			&action.AssertNodeOptions{
				Expected: []action.NodeOptionExpected{
					// Node-level
					{Path: []string{"DisableP2P"}, Value: true},
					{Path: []string{"EnableDevelopment"}, Value: false},
					// Store — in-memory; BadgerEncryptionKey null because keyring is disabled
					{Path: []string{"Store", "Store"}, Value: ""},
					{Path: []string{"Store", "BadgerInMemory"}, Value: true},
					{Path: []string{"Store", "BadgerEncryptionKey"}, Value: nil},
					// DB — ChunkSize is overridden to defaultChunkSize when BadgerInMemory is true
					{Path: []string{"DB", "EnableSigning"}, Value: false},
					{Path: []string{"DB", "Identity"}, Value: nil},
					{Path: []string{"DB", "MaxTxnRetries"}, Value: float64(7)},
					{Path: []string{"DB", "LensRuntime"}, Value: "wazero"},
					{Path: []string{"DB", "ChunkSize"}, Value: defaultChunkSize},
					{Path: []string{"DB", "RetryIntervals"}, Value: []any{float64(5000000000), float64(10000000000)}},
					{Path: []string{"DB", "SearchableEncryptionKey"}, Value: nil},
					// P2P — stored but not used because --no-p2p; pubsub disabled via env var
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
	// P2P.EnableRelay: make it true for set B (default is false from config).
	// Viper reads DEFRA_NET_RELAYENABLED via AutomaticEnv.
	t.Setenv("DEFRA_NET_RELAYENABLED", "true")
	// The file keyring backend requires a secret (cfg key "keyring.secret").
	// Viper maps it via AutomaticEnv + prefix DEFRA → DEFRA_KEYRING_SECRET.
	t.Setenv("DEFRA_KEYRING_SECRET", "test-secret-for-node-options-setb")

	test := &integration.Test{
		// Disk-mode + P2P startup is slower than in-memory; allow more time.
		Timeout: 10 * time.Second,
		Actions: []action.Action{
			// StartDiskWithArgs uses Badger on disk.
			// --no-keyring=false overrides the --no-keyring in the base args, enabling the
			// keyring so that P2P, encryption, and identity keys are created and stored.
			// This makes PrivateKey, BadgerEncryptionKey, SearchableEncryptionKey, and
			// Identity all "<redacted>" (vs nil in set A).
			// --node-acp-enable + --identity enables Node ACP (vs false in set A).
			// P2P is not disabled: the subsystem starts on a dedicated port.
			// --development sets EnableDevelopment=true (vs false in set A).
			action.StartDiskWithArgs([]string{
				"--no-keyring=false",
				"--development",
				"--max-txn-retries=14",
				"--p2paddr=/ip4/127.0.0.1/tcp/9173",
				// No --peers: bootstrap peers need a full multiaddr with peer ID to be valid
				// when P2P is running. Set A stores a peer (P2P disabled) but Set B cannot.
				// Empty vs non-empty still differentiates BootstrapPeers between the two sets.
				"--allowed-origins=http://b.example.com",
				"--replicator-retry-intervals=15,30",
				"--node-acp-enable",
				"--identity=" + testIdentityKeyHex,
			}),
			&action.AssertNodeOptions{
				Expected: []action.NodeOptionExpected{
					// Node-level
					{Path: []string{"DisableP2P"}, Value: false},
					{Path: []string{"EnableDevelopment"}, Value: true},
					// Store — disk mode; keyring creates an encryption key → redacted
					{Path: []string{"Store", "Store"}, Value: ""},
					{Path: []string{"Store", "BadgerInMemory"}, Value: false},
					{Path: []string{"Store", "BadgerEncryptionKey"}, Value: "<redacted>"},
					// DB — ChunkSize nil because BadgerInMemory is false
					{Path: []string{"DB", "EnableSigning"}, Value: true},
					{Path: []string{"DB", "Identity"}, Value: "<redacted>"},
					{Path: []string{"DB", "MaxTxnRetries"}, Value: float64(14)},
					{Path: []string{"DB", "LensRuntime"}, Value: "wazero"},
					{Path: []string{"DB", "ChunkSize"}, Value: nil},
					{Path: []string{"DB", "RetryIntervals"}, Value: []any{float64(15000000000), float64(30000000000)}},
					{Path: []string{"DB", "SearchableEncryptionKey"}, Value: "<redacted>"},
					// P2P — pubsub default true; relay enabled via env var; keyring creates private key
					{Path: []string{"P2P", "EnablePubSub"}, Value: true},
					{Path: []string{"P2P", "EnableRelay"}, Value: true},
					{Path: []string{"P2P", "ListenAddresses"}, Value: []any{"/ip4/127.0.0.1/tcp/9173"}},
					{Path: []string{"P2P", "BootstrapPeers"}, Value: []any{}},
					{Path: []string{"P2P", "PrivateKey"}, Value: "<redacted>"},
					// HTTP
					{Path: []string{"HTTP", "AllowedOrigins"}, Value: []any{"http://b.example.com"}},
					// Document ACP — disk mode sets path to rootdir
					{Path: []string{"DocumentACP", "DocumentACPType"}, Value: "local"},
					{Path: []string{"DocumentACP", "Path"}, UseRootDir: true},
					// Node ACP — enabled via --node-acp-enable; disk mode sets path to rootdir
					{Path: []string{"NodeACP", "IsEnabled"}, Value: true},
					{Path: []string{"NodeACP", "Path"}, UseRootDir: true},
				},
			},
		},
	}

	test.Execute(t)
}
