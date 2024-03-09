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
	"path/filepath"
	"syscall"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/db"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/http"
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
				// It would be ideal to not have the key path tied to the datastore.
				// Running with memory store mode will always generate a random key.
				// Adding support for an ephemeral mode and moving the key to the
				// config would solve both of these issues.
				rootdir := mustGetContextRootDir(cmd)
				key, err := loadOrGeneratePrivateKey(filepath.Join(rootdir, "data", "key"))
				if err != nil {
					return err
				}
				netOpts = append(netOpts, net.WithPrivateKey(key))
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
					log.FeedbackErrorE(cmd.Context(), "Stopping DefraDB", err)
				}
			}()

			log.FeedbackInfo(cmd.Context(), "Starting DefraDB")
			if err := n.Start(cmd.Context()); err != nil {
				return err
			}

			signalCh := make(chan os.Signal, 1)
			signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

			select {
			case <-cmd.Context().Done():
				log.FeedbackInfo(cmd.Context(), "Received context cancellation; shutting down...")
			case <-signalCh:
				log.FeedbackInfo(cmd.Context(), "Received interrupt; shutting down...")
			}

			return nil
		},
	}

	return cmd
}
