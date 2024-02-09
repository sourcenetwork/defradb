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
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/db"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/http"
	"github.com/sourcenetwork/defradb/net"
	netutils "github.com/sourcenetwork/defradb/net/utils"
	"github.com/sourcenetwork/defradb/node"
)

func MakeStartCommand(cfg *config.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "start",
		Short: "Start a DefraDB node",
		Long:  "Start a DefraDB node.",
		// Load the root config if it exists, otherwise create it.
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if err := loadConfig(cfg); err != nil {
				return err
			}
			if !cfg.ConfigFileExists() {
				return createConfig(cfg)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			dbOpts := []db.Option{
				db.WithUpdateEvents(),
				db.WithMaxRetries(cfg.Datastore.MaxTxnRetries),
			}

			netOpts := []net.NodeOpt{
				net.WithListenAddresses(cfg.Net.P2PAddresses...),
				net.WithEnablePubSub(cfg.Net.PubSubEnabled),
				net.WithEnableRelay(cfg.Net.RelayEnabled),
			}

			serverOpts := []http.ServerOpt{
				http.WithAddress(cfg.API.Address),
				http.WithAllowedOrigins(cfg.API.AllowedOrigins...),
				http.WithTLSCertPath(cfg.API.PubKeyPath),
				http.WithTLSKeyPath(cfg.API.PrivKeyPath),
			}

			storeOpts := []node.StoreOpt{
				node.WithPath(cfg.Datastore.Badger.Path),
				node.WithInMemory(cfg.Datastore.Store == config.DatastoreMemory),
			}

			var peers []peer.AddrInfo
			if cfg.Net.Peers != "" {
				addrs, err := netutils.ParsePeers(strings.Split(cfg.Net.Peers, ","))
				if err != nil {
					return errors.Wrap(fmt.Sprintf("failed to parse bootstrap peers %v", cfg.Net.Peers), err)
				}
				peers = addrs
			}

			if cfg.Datastore.Store == "badger" {
				// It would be ideal to not have the key path tied to the datastore.
				// Running with memory store mode will always generate a random key.
				// Adding support for an ephemeral mode and moving the key to the
				// config would solve both of these issues.
				key, err := loadOrGeneratePrivateKey(filepath.Join(cfg.Rootdir, "data", "key"))
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
				node.WithDisableP2P(cfg.Net.P2PDisabled),
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

	cmd.Flags().String(
		"peers", cfg.Net.Peers,
		"List of peers to connect to",
	)
	err := cfg.BindFlag("net.peers", cmd.Flags().Lookup("peers"))
	if err != nil {
		log.FeedbackFatalE(context.Background(), "Could not bind net.peers", err)
	}

	cmd.Flags().Int(
		"max-txn-retries", cfg.Datastore.MaxTxnRetries,
		"Specify the maximum number of retries per transaction",
	)
	err = cfg.BindFlag("datastore.maxtxnretries", cmd.Flags().Lookup("max-txn-retries"))
	if err != nil {
		log.FeedbackFatalE(context.Background(), "Could not bind datastore.maxtxnretries", err)
	}

	cmd.Flags().String(
		"store", cfg.Datastore.Store,
		"Specify the datastore to use (supported: badger, memory)",
	)
	err = cfg.BindFlag("datastore.store", cmd.Flags().Lookup("store"))
	if err != nil {
		log.FeedbackFatalE(context.Background(), "Could not bind datastore.store", err)
	}

	cmd.Flags().Var(
		&cfg.Datastore.Badger.ValueLogFileSize, "valuelogfilesize",
		"Specify the datastore value log file size (in bytes). In memory size will be 2*valuelogfilesize",
	)
	err = cfg.BindFlag("datastore.badger.valuelogfilesize", cmd.Flags().Lookup("valuelogfilesize"))
	if err != nil {
		log.FeedbackFatalE(context.Background(), "Could not bind datastore.badger.valuelogfilesize", err)
	}

	cmd.Flags().StringSlice(
		"p2paddr", cfg.Net.P2PAddresses,
		"Listen addresses for the p2p network (formatted as a libp2p MultiAddr)",
	)
	err = cfg.BindFlag("net.p2paddresses", cmd.Flags().Lookup("p2paddr"))
	if err != nil {
		log.FeedbackFatalE(context.Background(), "Could not bind net.p2paddress", err)
	}

	cmd.Flags().Bool(
		"no-p2p", cfg.Net.P2PDisabled,
		"Disable the peer-to-peer network synchronization system",
	)
	err = cfg.BindFlag("net.p2pdisabled", cmd.Flags().Lookup("no-p2p"))
	if err != nil {
		log.FeedbackFatalE(context.Background(), "Could not bind net.p2pdisabled", err)
	}

	cmd.Flags().Bool(
		"tls", cfg.API.TLS,
		"Enable serving the API over https",
	)
	err = cfg.BindFlag("api.tls", cmd.Flags().Lookup("tls"))
	if err != nil {
		log.FeedbackFatalE(context.Background(), "Could not bind api.tls", err)
	}

	cmd.Flags().StringArray(
		"allowed-origins", cfg.API.AllowedOrigins,
		"List of origins to allow for CORS requests",
	)
	err = cfg.BindFlag("api.allowed-origins", cmd.Flags().Lookup("allowed-origins"))
	if err != nil {
		log.FeedbackFatalE(context.Background(), "Could not bind api.allowed-origins", err)
	}

	cmd.Flags().String(
		"pubkeypath", cfg.API.PubKeyPath,
		"Path to the public key for tls",
	)
	err = cfg.BindFlag("api.pubkeypath", cmd.Flags().Lookup("pubkeypath"))
	if err != nil {
		log.FeedbackFatalE(context.Background(), "Could not bind api.pubkeypath", err)
	}

	cmd.Flags().String(
		"privkeypath", cfg.API.PrivKeyPath,
		"Path to the private key for tls",
	)
	err = cfg.BindFlag("api.privkeypath", cmd.Flags().Lookup("privkeypath"))
	if err != nil {
		log.FeedbackFatalE(context.Background(), "Could not bind api.privkeypath", err)
	}

	cmd.Flags().String(
		"email", cfg.API.Email,
		"Email address used by the CA for notifications",
	)
	err = cfg.BindFlag("api.email", cmd.Flags().Lookup("email"))
	if err != nil {
		log.FeedbackFatalE(context.Background(), "Could not bind api.email", err)
	}
	return cmd
}
