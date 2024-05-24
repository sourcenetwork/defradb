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
	"github.com/spf13/pflag"

	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/http"
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/keyring"
	"github.com/sourcenetwork/defradb/net"
	netutils "github.com/sourcenetwork/defradb/net/utils"
	"github.com/sourcenetwork/defradb/node"
)

// startFlags is a set of persistent flags that are bound to config values.
var startFlags = pflag.NewFlagSet("start", pflag.ContinueOnError)

func init() {
	startFlags.StringArray(
		"peers",
		[]string{},
		"List of peers to connect to",
	)
	startFlags.Int(
		"max-txn-retries",
		5,
		"Specify the maximum number of retries per transaction",
	)
	startFlags.String(
		"store",
		"badger",
		"Specify the datastore to use (supported: badger, memory)",
	)
	startFlags.Int(
		"valuelogfilesize",
		1<<30,
		"Specify the datastore value log file size (in bytes). In memory size will be 2*valuelogfilesize",
	)
	startFlags.StringSlice(
		"p2paddr",
		[]string{"/ip4/127.0.0.1/tcp/9171"},
		"Listen addresses for the p2p network (formatted as a libp2p MultiAddr)",
	)
	startFlags.Bool(
		"no-p2p",
		false,
		"Disable the peer-to-peer network synchronization system",
	)
	startFlags.StringArray(
		"allowed-origins",
		[]string{},
		"List of origins to allow for CORS requests",
	)
	startFlags.String(
		"pubkeypath",
		"",
		"Path to the public key for tls",
	)
	startFlags.String(
		"privkeypath",
		"",
		"Path to the private key for tls",
	)
}

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
			if err := createConfig(rootdir); err != nil {
				return err
			}
			return setContextConfig(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := mustGetContextConfig(cmd)

			var peers []peer.AddrInfo
			if val := cfg.GetStringSlice("net.peers"); len(val) > 0 {
				addrs, err := netutils.ParsePeers(val)
				if err != nil {
					return errors.Wrap(fmt.Sprintf("failed to parse bootstrap peers %s", val), err)
				}
				peers = addrs
			}

			opts := []node.Option{
				node.WithPath(cfg.GetString("datastore.badger.path")),
				node.WithInMemory(cfg.GetString("datastore.store") == configStoreMemory),
				node.WithDisableP2P(cfg.GetBool("net.p2pDisabled")),
				node.WithACPType(node.LocalACPType),
				node.WithPeers(peers...),
				// db options
				db.WithUpdateEvents(),
				db.WithMaxRetries(cfg.GetInt("datastore.MaxTxnRetries")),
				// net node options
				net.WithListenAddresses(cfg.GetStringSlice("net.p2pAddresses")...),
				net.WithEnablePubSub(cfg.GetBool("net.pubSubEnabled")),
				net.WithEnableRelay(cfg.GetBool("net.relayEnabled")),
				// http server options
				http.WithAddress(cfg.GetString("api.address")),
				http.WithAllowedOrigins(cfg.GetStringSlice("api.allowed-origins")...),
				http.WithTLSCertPath(cfg.GetString("api.pubKeyPath")),
				http.WithTLSKeyPath(cfg.GetString("api.privKeyPath")),
			}

			if cfg.GetString("datastore.store") != configStoreMemory {
				rootDir := mustGetContextRootDir(cmd)
				// TODO-ACP: Infuture when we add support for the --no-acp flag when admin signatures are in,
				// we can allow starting of db without acp. Currently that can only be done programmatically.
				// https://github.com/sourcenetwork/defradb/issues/2271
				opts = append(opts, node.WithACPPath(rootDir))
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
				opts = append(opts, net.WithPrivateKey(peerKey))
				// load the optional encryption key
				encryptionKey, err := kr.Get(encryptionKeyName)
				if err != nil && !errors.Is(err, keyring.ErrNotFound) {
					return err
				}
				opts = append(opts, node.WithEncryptionKey(encryptionKey))
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

	cmd.PersistentFlags().AddFlagSet(startFlags)

	return cmd
}
