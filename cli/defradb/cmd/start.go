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
	gonet "net"
	"os"
	"os/signal"
	"strings"
	"time"

	ma "github.com/multiformats/go-multiaddr"
	badgerds "github.com/sourcenetwork/defradb/datastores/badger/v3"
	"github.com/sourcenetwork/defradb/db"
	netapi "github.com/sourcenetwork/defradb/net/api"
	netpb "github.com/sourcenetwork/defradb/net/api/pb"
	netutils "github.com/sourcenetwork/defradb/net/utils"
	"github.com/sourcenetwork/defradb/node"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	badger "github.com/dgraph-io/badger/v3"
	ds "github.com/ipfs/go-datastore"
	"github.com/sourcenetwork/defradb/logging"
	"github.com/spf13/cobra"
	"github.com/textileio/go-threads/broadcast"
)

var (
	p2pAddr  string
	tcpAddr  string
	dataPath string
	peers    string

	busBufferSize = 100
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a DefraDB server ",
	Long:  `Start a new instance of DefraDB server:`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		logging.SetConfig(config.Logging.toLogConfig())
		log.Info(ctx, "Starting DefraDB process...")

		// setup signal handlers
		signalCh := make(chan os.Signal, 1)
		signal.Notify(signalCh, os.Interrupt)

		var rootstore ds.Batching

		var err error
		if config.Database.Store == "badger" {
			log.Info(ctx, "opening badger store", logging.NewKV("Path", config.Database.Badger.Path))
			rootstore, err = badgerds.NewDatastore(config.Database.Badger.Path, config.Database.Badger.Options)
		} else if config.Database.Store == "memory" {
			log.Info(ctx, "building new memory store")
			opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
			rootstore, err = badgerds.NewDatastore("", &opts)
		}

		if err != nil {
			log.FatalE(ctx, "Failed to initiate datastore", err)
		}

		var options []db.Option

		// check for p2p
		var bs *broadcast.Broadcaster
		if !config.Net.P2PDisabled {
			bs = broadcast.NewBroadcaster(busBufferSize)
			options = append(options, db.WithBroadcaster(bs))
		}

		db, err := db.NewDB(ctx, rootstore, options...)
		if err != nil {
			log.FatalE(ctx, "Failed to initiate database", err)
		}

		// init the p2p node
		var n *node.Node
		if !config.Net.P2PDisabled {
			log.Info(ctx, "Starting P2P node", logging.NewKV("tcp address", config.Net.TCPAddress))
			n, err = node.NewNode(
				ctx,
				db,
				bs,
				node.DataPath(config.Database.Badger.Path),
				node.ListenP2PAddrStrings(config.Net.P2PAddress),
				node.WithPubSub(true))
			if err != nil {
				log.ErrorE(ctx, "Failed to start p2p node:", err)
				n.Close() //nolint
				db.Close()
				os.Exit(1)
			}

			// parse peers and bootstrap
			if len(peers) != 0 {
				log.Debug(ctx, "Parsing boostrap peers", logging.NewKV("Peers", peers))
				addrs, err := netutils.ParsePeers(strings.Split(peers, ","))
				if err != nil {
					log.ErrorE(ctx, "Failed to parse boostrap peers", err)
				}
				log.Debug(ctx, "Bootstraping with peers", logging.NewKV("Addresses", addrs))
				n.Boostrap(addrs)
			}

			if err := n.Start(); err != nil {
				log.ErrorE(ctx, "Failed to start p2p listeners:", err)
				n.Close() //nolint
				db.Close()
				os.Exit(1)
			}

			MtcpAddr, err := ma.NewMultiaddr(tcpAddr)
			if err != nil {
				log.FatalE(ctx, "Error parsing multi-address,", err)
			}
			addr, err := netutils.TCPAddrFromMultiAddr(MtcpAddr)
			if err != nil {
				log.FatalE(ctx, "Failed to parse TCP address", err)
			}

			server := grpc.NewServer(grpc.KeepaliveParams(keepalive.ServerParameters{
				MaxConnectionIdle: 5 * time.Minute,
			}))
			tcplistener, err := gonet.Listen("tcp", addr)
			if err != nil {
				log.FatalE(ctx, "Failed to listen to TCP address", err, logging.NewKV("Address", addr))
			}

			netService := netapi.NewService(n.Peer)

			go func() {
				log.Info(ctx, "Started gRPC server", logging.NewKV("Address", addr))
				netpb.RegisterServiceServer(server, netService)
				if err := server.Serve(tcplistener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
					log.FatalE(ctx, "serve error", err)
				}
			}()
		}

		// run the server listener in a seperate goroutine
		go func() {
			if err := db.Listen(config.Database.Address); err != nil {
				log.ErrorE(ctx, "Failed to start API listener", err)
				if n != nil {
					n.Close() //nolint
				}
				db.Close()
				os.Exit(1)
			}
		}()

		// wait for shutdown signal
		<-signalCh
		log.Info(ctx, "Recieved interrupt; closing db")
		if n != nil {
			n.Close() //nolint
		}
		db.Close()
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
	startCmd.Flags().String("store", "badger", "Specify the data store to use (supported: badger, memory)")
	startCmd.Flags().StringVar(&peers, "peers", "", "list of peers to connect to")
	startCmd.Flags().StringVar(&p2pAddr, "p2paddr", "/ip4/0.0.0.0/tcp/9171", "listener address for the p2p network (formatted as a libp2p MultiAddr)")
	startCmd.Flags().StringVar(&tcpAddr, "tcpaddr", "/ip4/0.0.0.0/tcp/9161", "listener address for the tcp gRPC server (formatted as a libp2p MultiAddr)")
	startCmd.Flags().StringVar(&dataPath, "data", "$HOME/.defradb/data", "Data path to save DB data and other related meta-data")
	startCmd.Flags().Bool("no-p2p", false, "Turn off the peer-to-peer network synchroniation system")
}
