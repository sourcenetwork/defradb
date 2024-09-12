// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cli

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sourcenetwork/immutable"
	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/datastore/badger/v4"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/http"
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/keyring"
	"github.com/sourcenetwork/defradb/net"
	"github.com/sourcenetwork/defradb/node"
)

func MakeStartCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "start",
		Short: "Start a DefraDB node",
		Long:  "Start a DefraDB node.",
		// Load the root config if it exists, otherwise create it.
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if err := setContextRootDir(cmd); err != nil {
				return err
			}
			rootdir := mustGetContextRootDir(cmd)
			if err := createConfig(rootdir, cmd.Flags()); err != nil {
				return err
			}
			return setContextConfig(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := mustGetContextConfig(cmd)

			storeOpts := []node.StoreOpt{
				node.WithStorePath(cfg.GetString("datastore.badger.path")),
				node.WithBadgerInMemory(cfg.GetString("datastore.store") == configStoreMemory),
			}
			nodeOpts := []node.Option{
				node.WithDisableP2P(cfg.GetBool("net.p2pDisabled")),
				node.WithSourceHubChainID(cfg.GetString("acp.sourceHub.ChainID")),
				node.WithSourceHubGRPCAddress(cfg.GetString("acp.sourceHub.GRPCAddress")),
				node.WithSourceHubCometRPCAddress(cfg.GetString("acp.sourceHub.CometRPCAddress")),
				// db options
				db.WithMaxRetries(cfg.GetInt("datastore.MaxTxnRetries")),
				// net node options
				net.WithListenAddresses(cfg.GetStringSlice("net.p2pAddresses")...),
				net.WithEnablePubSub(cfg.GetBool("net.pubSubEnabled")),
				net.WithEnableRelay(cfg.GetBool("net.relayEnabled")),
				net.WithBootstrapPeers(cfg.GetStringSlice("net.peers")...),
				// http server options
				http.WithAddress(cfg.GetString("api.address")),
				http.WithAllowedOrigins(cfg.GetStringSlice("api.allowed-origins")...),
				http.WithTLSCertPath(cfg.GetString("api.pubKeyPath")),
				http.WithTLSKeyPath(cfg.GetString("api.privKeyPath")),
				node.WithLensRuntime(node.LensRuntimeType(cfg.GetString("lens.runtime"))),
			}

			if cfg.GetString("datastore.store") != configStoreMemory {
				rootDir := mustGetContextRootDir(cmd)
				// TODO-ACP: Infuture when we add support for the --no-acp flag when admin signatures are in,
				// we can allow starting of db without acp. Currently that can only be done programmatically.
				// https://github.com/sourcenetwork/defradb/issues/2271
				nodeOpts = append(nodeOpts, node.WithACPPath(rootDir))
			}

			if !cfg.GetBool("keyring.disabled") {
				kr, err := openKeyring(cmd)
				if err != nil {
					return NewErrKeyringHelp(err)
				}

				// load the required peer key
				peerKey, err := kr.Get(peerKeyName)
				if err != nil {
					return NewErrKeyringHelp(err)
				}
				nodeOpts = append(nodeOpts, net.WithPrivateKey(peerKey))

				// load the optional encryption key
				encryptionKey, err := kr.Get(encryptionKeyName)
				if err != nil && !errors.Is(err, keyring.ErrNotFound) {
					return err
				}
				storeOpts = append(storeOpts, node.WithBadgerEncryptionKey(encryptionKey))

				sourceHubKeyName := cfg.GetString("acp.sourceHub.KeyName")
				if sourceHubKeyName != "" {
					signer, err := keyring.NewTxSignerFromKeyringKey(kr, sourceHubKeyName)
					if err != nil {
						return err
					}
					nodeOpts = append(nodeOpts, node.WithTxnSigner(immutable.Some[node.TxSigner](signer)))
				}
			}

			acpType := cfg.GetString("acp.type")
			if acpType != "" {
				nodeOpts = append(nodeOpts, node.WithACPType(node.ACPType(acpType)))
			}

			signalCh := make(chan os.Signal, 1)
			signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

			for _, o := range storeOpts {
				nodeOpts = append(nodeOpts, any(o))
			}

		START:
			n, err := node.NewNode(cmd.Context(), nodeOpts...)
			if err != nil {
				return err
			}
			log.InfoContext(cmd.Context(), "Starting DefraDB")
			if err := n.Start(cmd.Context()); err != nil {
				return err
			}
			purgeSub, err := n.DB.Events().Subscribe(event.PurgeName)
			if err != nil {
				return err
			}

		SELECT:
			select {
			case <-purgeSub.Message():
				if !cfg.GetBool("development") {
					goto SELECT
				}
				log.InfoContext(cmd.Context(), "Received purge event; restarting...")
				if err := n.Close(cmd.Context()); err != nil {
					return err
				}
				rootstore, err := node.NewStore(cmd.Context(), storeOpts...)
				if err != nil {
					return err
				}
				if b, ok := rootstore.(*badger.Datastore); ok {
					// only badger persists data so drop all values manually
					if err := b.DB.DropAll(); err != nil {
						return err
					}
				}
				if err := rootstore.Close(); err != nil {
					return err
				}
				goto START

			case <-cmd.Context().Done():
				log.InfoContext(cmd.Context(), "Received context cancellation; shutting down...")

			case <-signalCh:
				log.InfoContext(cmd.Context(), "Received interrupt; shutting down...")
			}

			return n.Close(cmd.Context())
		},
	}
	// set default flag values from config
	cfg := defaultConfig()
	cmd.PersistentFlags().StringArray(
		"peers",
		cfg.GetStringSlice(configFlags["peers"]),
		"List of peers to connect to",
	)
	cmd.PersistentFlags().Int(
		"max-txn-retries",
		cfg.GetInt(configFlags["max-txn-retries"]),
		"Specify the maximum number of retries per transaction",
	)
	cmd.PersistentFlags().String(
		"store",
		cfg.GetString(configFlags["store"]),
		"Specify the datastore to use (supported: badger, memory)",
	)
	cmd.PersistentFlags().Int(
		"valuelogfilesize",
		cfg.GetInt(configFlags["valuelogfilesize"]),
		"Specify the datastore value log file size (in bytes). In memory size will be 2*valuelogfilesize",
	)
	cmd.PersistentFlags().StringSlice(
		"p2paddr",
		cfg.GetStringSlice(configFlags["p2paddr"]),
		"Listen addresses for the p2p network (formatted as a libp2p MultiAddr)",
	)
	cmd.PersistentFlags().Bool(
		"no-p2p",
		cfg.GetBool(configFlags["no-p2p"]),
		"Disable the peer-to-peer network synchronization system",
	)
	cmd.PersistentFlags().StringArray(
		"allowed-origins",
		cfg.GetStringSlice(configFlags["allowed-origins"]),
		"List of origins to allow for CORS requests",
	)
	cmd.PersistentFlags().String(
		"pubkeypath",
		cfg.GetString(configFlags["pubkeypath"]),
		"Path to the public key for tls",
	)
	cmd.PersistentFlags().String(
		"privkeypath",
		cfg.GetString(configFlags["privkeypath"]),
		"Path to the private key for tls",
	)
	cmd.PersistentFlags().Bool(
		"development",
		cfg.GetBool(configFlags["development"]),
		"Enables a set of features that make development easier but should not be enabled in production",
	)
	return cmd
}
