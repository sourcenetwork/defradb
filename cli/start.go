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
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/http"
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/keyring"
	"github.com/sourcenetwork/defradb/net"
	netutils "github.com/sourcenetwork/defradb/net/utils"
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
			if err := createConfig(rootdir, cmd.Root().PersistentFlags()); err != nil {
				return err
			}
			return setContextConfig(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := mustGetContextConfig(cmd)

			dbOpts := []db.Option{
				db.WithUpdateEvents(),
				db.WithMaxRetries(cfg.GetInt("datastore.MaxTxnRetries")),
				// TODO-ACP: Infuture when we add support for the --no-acp flag when admin signatures are in,
				// we can allow starting of db without acp. Currently that can only be done programmatically.
				// https://github.com/sourcenetwork/defradb/issues/2271
				db.WithACPInMemory(),
			}

			netOpts := []net.NodeOpt{
				net.WithListenAddresses(cfg.GetStringSlice("net.p2pAddresses")...),
				net.WithEnablePubSub(cfg.GetBool("net.pubSubEnabled")),
				net.WithEnableRelay(cfg.GetBool("net.relayEnabled")),
			}

			serverOpts := []http.ServerOpt{
				http.WithAddress(cfg.GetString("api.address")),
				http.WithAllowedOrigins(cfg.GetStringSlice("api.allowed-origins")...),
				http.WithTLSCertPath(cfg.GetString("api.pubKeyPath")),
				http.WithTLSKeyPath(cfg.GetString("api.privKeyPath")),
			}

			storeOpts := []node.StoreOpt{
				node.WithPath(cfg.GetString("datastore.badger.path")),
				node.WithInMemory(cfg.GetString("datastore.store") == configStoreMemory),
			}

			var peers []peer.AddrInfo
			if val := cfg.GetStringSlice("net.peers"); len(val) > 0 {
				addrs, err := netutils.ParsePeers(val)
				if err != nil {
					return errors.Wrap(fmt.Sprintf("failed to parse bootstrap peers %s", val), err)
				}
				peers = addrs
			}

			if cfg.GetString("datastore.store") != configStoreMemory {
				rootDir := mustGetContextRootDir(cmd)
				// TODO-ACP: Infuture when we add support for the --no-acp flag when admin signatures are in,
				// we can allow starting of db without acp. Currently that can only be done programmatically.
				// https://github.com/sourcenetwork/defradb/issues/2271
				dbOpts = append(dbOpts, db.WithACP(rootDir))
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
				netOpts = append(netOpts, net.WithPrivateKey(peerKey))
				// load the optional encryption key
				encryptionKey, err := kr.Get(encryptionKeyName)
				if err != nil && !errors.Is(err, keyring.ErrNotFound) {
					return err
				}
				storeOpts = append(storeOpts, node.WithEncryptionKey(encryptionKey))
			}

			opts := []node.NodeOpt{
				node.WithPeers(peers...),
				node.WithStoreOpts(storeOpts...),
				node.WithDatabaseOpts(dbOpts...),
				node.WithNetOpts(netOpts...),
				node.WithServerOpts(serverOpts...),
				node.WithDisableP2P(cfg.GetBool("net.p2pDisabled")),
			}

			n, err := node.NewNode(cmd.Context(), opts...)
			if err != nil {
				return err
			}

			defer func() {
				if err := n.Close(cmd.Context()); err != nil {
					log.ErrorContextE(cmd.Context(), "Stopping DefraDB", err)
				}
			}()

			log.InfoContext(cmd.Context(), "Starting DefraDB")
			if err := n.Start(cmd.Context()); err != nil {
				return err
			}

			signalCh := make(chan os.Signal, 1)
			signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

			select {
			case <-cmd.Context().Done():
				log.InfoContext(cmd.Context(), "Received context cancellation; shutting down...")
			case <-signalCh:
				log.InfoContext(cmd.Context(), "Received interrupt; shutting down...")
			}

			return nil
		},
	}

	return cmd
}
