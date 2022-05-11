// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cmd

import (
	"context"
	"errors"
	"fmt"
	gonet "net"
	"os"
	"os/signal"
	"strings"

	ma "github.com/multiformats/go-multiaddr"
	httpapi "github.com/sourcenetwork/defradb/api/http"
	"github.com/sourcenetwork/defradb/config"
	badgerds "github.com/sourcenetwork/defradb/datastore/badger/v3"
	"github.com/sourcenetwork/defradb/db"
	netapi "github.com/sourcenetwork/defradb/net/api"
	netpb "github.com/sourcenetwork/defradb/net/api/pb"
	netutils "github.com/sourcenetwork/defradb/net/utils"
	"github.com/sourcenetwork/defradb/node"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	badger "github.com/dgraph-io/badger/v3"
	ds "github.com/ipfs/go-datastore"
	"github.com/sourcenetwork/defradb/api/http"
	"github.com/sourcenetwork/defradb/logging"
	"github.com/spf13/cobra"
	"github.com/textileio/go-threads/broadcast"
)

var (
	busBufferSize = 100
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a DefraDB server ",
	Long:  `Start a new instance of DefraDB server:`,
	// Load the root config if it exists, otherwise create it.
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		rootDir, exists, err := config.GetRootDir(rootDir)
		if err != nil {
			log.FatalE(ctx, "Failed to get root dir", err)
		}
		if !exists {
			err = config.CreateRootDirWithDefaultConfig(rootDir)
			if err != nil {
				log.FatalE(ctx, "Failed to create root dir", err)
			}
		}
		err = cfg.Load(rootDir)
		if err != nil {
			log.FatalE(ctx, "Failed to load config", err)
		}
		loggingConfig, err := cfg.GetLoggingConfig()
		if err != nil {
			log.FatalE(ctx, "Failed to load logging config", err)
		}
		logging.SetConfig(loggingConfig)
		log.Info(ctx, "Configuration loaded from DefraDB root directory", logging.NewKV("rootdir", rootDir))
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		log.Info(ctx, "Starting DefraDB service...")

		// setup signal handlers
		signalCh := make(chan os.Signal, 1)
		signal.Notify(signalCh, os.Interrupt)

		var rootstore ds.Batching

		var err error
		if cfg.Datastore.Store == "badger" {
			log.Info(
				ctx,
				"Opening badger store",
				logging.NewKV("Path", cfg.Datastore.Badger.Path),
			)
			rootstore, err = badgerds.NewDatastore(
				cfg.Datastore.Badger.Path,
				cfg.Datastore.Badger.Options,
			)
		} else if cfg.Datastore.Store == "memory" {
			log.Info(ctx, "Building new memory store")
			opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
			rootstore, err = badgerds.NewDatastore("", &opts)
		}

		if err != nil {
			log.FatalE(ctx, "Failed to initiate datastore", err)
		}

		var options []db.Option

		// check for p2p
		var bs *broadcast.Broadcaster
		if !cfg.Net.P2PDisabled {
			bs = broadcast.NewBroadcaster(busBufferSize)
			options = append(options, db.WithBroadcaster(bs))
		}

		db, err := db.NewDB(ctx, rootstore, options...)
		if err != nil {
			log.FatalE(ctx, "Failed to initiate database", err)
		}

		// init the p2p node
		var n *node.Node
		if !cfg.Net.P2PDisabled {
			log.Info(ctx, "Starting P2P node", logging.NewKV("P2P address", cfg.Net.P2PAddress))
			nodeConfig, err := cfg.NodeConfig()
			if err != nil {
				log.FatalE(ctx, "Failed to get node config", err)
			}
			n, err = node.NewNode(
				ctx,
				db,
				bs,
				nodeConfig,
			)
			if err != nil {
				log.ErrorE(ctx, "Failed to start p2p node", err)
				n.Close() //nolint
				db.Close(ctx)
				os.Exit(1)
			}

			// parse peers and bootstrap
			if len(cfg.Net.Peers) != 0 {
				log.Debug(ctx, "Parsing bootstrap peers", logging.NewKV("Peers", cfg.Net.Peers))
				addrs, err := netutils.ParsePeers(strings.Split(cfg.Net.Peers, ","))
				if err != nil {
					log.ErrorE(ctx, "Failed to parse bootstrap peers", err)
				}
				log.Debug(ctx, "Bootstraping with peers", logging.NewKV("Addresses", addrs))
				n.Boostrap(addrs)
			}

			if err := n.Start(); err != nil {
				log.ErrorE(ctx, "Failed to start p2p listeners", err)
				n.Close() //nolint
				db.Close(ctx)
				os.Exit(1)
			}

			MtcpAddr, err := ma.NewMultiaddr(cfg.Net.TCPAddress)
			if err != nil {
				log.FatalE(ctx, "Error parsing multi-address,", err)
			}
			addr, err := netutils.TCPAddrFromMultiAddr(MtcpAddr)
			if err != nil {
				log.FatalE(ctx, "Failed to parse TCP address", err)
			}

			rpcTimeoutDuration, err := cfg.Net.RPCTimeoutDuration()
			if err != nil {
				log.FatalE(ctx, "Failed to parse RPC timeout duration", err)
			}

			server := grpc.NewServer(grpc.KeepaliveParams(keepalive.ServerParameters{
				MaxConnectionIdle: rpcTimeoutDuration,
			}))
			tcplistener, err := gonet.Listen("tcp", addr)
			if err != nil {
				log.FatalE(
					ctx,
					"Failed to listen to TCP address",
					err,
					logging.NewKV("Address", addr),
				)
			}

			netService := netapi.NewService(n.Peer)

			go func() {
				log.Info(ctx, "Started gRPC server", logging.NewKV("Address", addr))
				netpb.RegisterServiceServer(server, netService)
				if err := server.Serve(tcplistener); err != nil &&
					!errors.Is(err, grpc.ErrServerStopped) {
					log.FatalE(ctx, "serve error", err)
				}
			}()
		}

		// run the server listener in a separate goroutine
		go func() {
			log.Info(
				ctx,
				fmt.Sprintf(
					"Providing HTTP API at %s%s. Use the GraphQL query endpoint at %s%s/graphql ",
					cfg.API.AddressToURL(),
					httpapi.RootPath,
					cfg.API.AddressToURL(),
					httpapi.RootPath,
				),
			)
			s := http.NewServer(db)
			if err := s.Listen(cfg.API.Address); err != nil {
				log.ErrorE(ctx, "Failed to start HTTP API listener", err)
				if n != nil {
					n.Close() //nolint
				}
				db.Close(ctx)
				os.Exit(1)
			}
		}()

		// wait for shutdown signal
		<-signalCh
		log.Info(ctx, "Received interrupt; closing db")
		if n != nil {
			n.Close() //nolint
		}
		db.Close(ctx)
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	startCmd.Flags().StringVar(
		&cfg.Net.Peers,
		"peers",
		defaultCfg.Net.Peers,
		"list of peers to connect to",
	)
	startCmd.Flags().StringVar(
		&cfg.Datastore.Store,
		"store",
		defaultCfg.Datastore.Store,
		"Specify the data store to use (supported: badger, memory)",
	)
	startCmd.Flags().StringVar(
		&cfg.Net.P2PAddress,
		"p2paddr",
		defaultCfg.Net.P2PAddress,
		"listener address for the p2p network (formatted as a libp2p MultiAddr)",
	)
	startCmd.Flags().StringVar(
		&cfg.Net.TCPAddress,
		"tcpaddr",
		defaultCfg.Net.TCPAddress,
		"listener address for the tcp gRPC server (formatted as a libp2p MultiAddr)",
	)
	startCmd.Flags().BoolVar(
		&cfg.Net.P2PDisabled,
		"no-p2p",
		defaultCfg.Net.P2PDisabled,
		"Turn off the peer-to-peer network synchronization system",
	)
}
