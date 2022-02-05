// Copyright 2020 Source Inc.
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
	"os"
	"os/signal"
	"strings"

	badgerds "github.com/sourcenetwork/defradb/datastores/badger/v3"
	"github.com/sourcenetwork/defradb/db"
	netutils "github.com/sourcenetwork/defradb/net/utils"
	"github.com/sourcenetwork/defradb/node"

	badger "github.com/dgraph-io/badger/v3"
	ds "github.com/ipfs/go-datastore"
	"github.com/spf13/cobra"
)

var (
	p2pAddr  string
	dataPath string
	peers    string
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a DefraDB server ",
	Long:  `Start a new instance of DefraDB server:`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Starting DefraDB process...")
		ctx := context.Background()

		// setup signal handlers
		signalCh := make(chan os.Signal, 1)
		signal.Notify(signalCh, os.Interrupt)

		var rootstore ds.Batching
		var options interface{}
		var err error
		if config.Database.Store == "badger" {
			log.Info("opening badger store: ", config.Database.Badger.Path)
			rootstore, err = badgerds.NewDatastore(config.Database.Badger.Path, config.Database.Badger.Options)
			options = config.Database.Badger
		} else if config.Database.Store == "memory" {
			log.Info("building new memory store")
			opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
			rootstore, err = badgerds.NewDatastore("", &opts)
			options = config.Database.Memory
		}

		if err != nil {
			log.Error("Failed to initiate datastore:", err)
			os.Exit(1)
		}

		db, err := db.NewDB(rootstore, options)
		if err != nil {
			log.Error("Failed to initiate database:", err)
			os.Exit(1)
		}
		if err := db.Start(ctx); err != nil {
			log.Error("Failed to start the database: ", err)
			db.Close()
			os.Exit(1)
		}

		// init the p2p node
		n, err := node.NewNode(
			ctx,
			db,
			node.DataPath(config.Database.Badger.Path),
			node.ListenAddrStrings(config.Net.P2PAddress),
			node.WithPubSub(true))
		if err != nil {
			log.Error("Failed to start p2p node:", err)
			db.Close()
			os.Exit(1)
		}

		db.SetPeer(n.Peer)

		// parse peers and bootstrap
		if len(peers) != 0 {
			log.Debug("Parsing boostrap peers: ", peers)
			addrs, err := netutils.ParsePeers(strings.Split(peers, ","))
			if err != nil {
				log.Warn("Failed to parse boostrap peers: ", err)
			}
			log.Debug("Bootstraping with peers: ", addrs)
			n.Boostrap(addrs)
		}

		// run the server listener in a seperate goroutine
		serveErr := make(chan error)
		go func() {
			if err := n.Start(); err != nil {
				log.Error("Failed to start p2p listener:", err)
				db.Close()
				os.Exit(1)
			}
			if err := db.Listen(config.Database.Address); err != nil {
				log.Error("Failed to start API listener:", err)
				db.Close()
				os.Exit(1)
			}
		}()

		select {
		case <-signalCh:
			log.Info("Recieved interrupt; closing db")
			db.Close()
			os.Exit(0)
		case <-serveErr:
			log.Error("API Server failed: ", err)
			db.Close()
			os.Exit(1)
		}

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
	startCmd.Flags().StringVar(&dataPath, "data", "$HOME/.defradb/data", "Data path to save DB data and other related meta-data")
}
