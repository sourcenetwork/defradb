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
	gonet "net"
	"net/http"
	"os"
	"os/signal"
	"strings"

	badger "github.com/dgraph-io/badger/v3"
	ds "github.com/ipfs/go-datastore"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	httpapi "github.com/sourcenetwork/defradb/api/http"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/config"
	badgerds "github.com/sourcenetwork/defradb/datastore/badger/v3"
	"github.com/sourcenetwork/defradb/db"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/logging"
	netapi "github.com/sourcenetwork/defradb/net/api"
	netpb "github.com/sourcenetwork/defradb/net/api/pb"
	netutils "github.com/sourcenetwork/defradb/net/utils"
	"github.com/sourcenetwork/defradb/node"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a DefraDB node",
	Long:  "Start a new instance of DefraDB node.",
	// Load the root config if it exists, otherwise create it.
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		rootDir, exists, err := config.GetRootDir(rootDirParam)
		if err != nil {
			return errors.Wrap("failed to get root dir", err)
		}
		if !exists {
			err = config.CreateRootDirWithDefaultConfig(rootDir)
			if err != nil {
				return errors.Wrap("failed to create root dir", err)
			}
		}
		err = cfg.Load(rootDir)
		if err != nil {
			return errors.Wrap("failed to load config", err)
		}

		// parse loglevel overrides
		if err := parseAndConfigLog(cmd.Context(), cfg.Log, cmd); err != nil {
			return err
		}
		log.FeedbackInfo(cmd.Context(), fmt.Sprintf("Configuration loaded from DefraDB directory %v", rootDir))
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		di, err := start(cmd.Context())
		if err != nil {
			return err
		}

		wait(cmd.Context(), di)

		return nil
	},
}

func init() {
	startCmd.Flags().String(
		"peers", cfg.Net.Peers,
		"List of peers to connect to",
	)
	err := viper.BindPFlag("net.peers", startCmd.Flags().Lookup("peers"))
	if err != nil {
		log.FeedbackFatalE(context.Background(), "Could not bind net.peers", err)
	}

	startCmd.Flags().String(
		"store", cfg.Datastore.Store,
		"Specify the datastore to use (supported: badger, memory)",
	)
	err = viper.BindPFlag("datastore.store", startCmd.Flags().Lookup("store"))
	if err != nil {
		log.FeedbackFatalE(context.Background(), "Could not bind datastore.store", err)
	}

	startCmd.Flags().Var(
		&cfg.Datastore.Badger.ValueLogFileSize, "valuelogfilesize",
		"Specify the datastore value log file size (in bytes). In memory size will be 2*valuelogfilesize",
	)
	err = viper.BindPFlag("datastore.badger.valuelogfilesize", startCmd.Flags().Lookup("valuelogfilesize"))
	if err != nil {
		log.FeedbackFatalE(context.Background(), "Could not bind datastore.badger.valuelogfilesize", err)
	}

	startCmd.Flags().String(
		"p2paddr", cfg.Net.P2PAddress,
		"Listener address for the p2p network (formatted as a libp2p MultiAddr)",
	)
	err = viper.BindPFlag("net.p2paddress", startCmd.Flags().Lookup("p2paddr"))
	if err != nil {
		log.FeedbackFatalE(context.Background(), "Could not bind net.p2paddress", err)
	}

	startCmd.Flags().String(
		"tcpaddr", cfg.Net.TCPAddress,
		"Listener address for the tcp gRPC server (formatted as a libp2p MultiAddr)",
	)
	err = viper.BindPFlag("net.tcpaddress", startCmd.Flags().Lookup("tcpaddr"))
	if err != nil {
		log.FeedbackFatalE(context.Background(), "Could not bind net.tcpaddress", err)
	}

	startCmd.Flags().Bool(
		"no-p2p", cfg.Net.P2PDisabled,
		"Disable the peer-to-peer network synchronization system",
	)
	err = viper.BindPFlag("net.p2pdisabled", startCmd.Flags().Lookup("no-p2p"))
	if err != nil {
		log.FeedbackFatalE(context.Background(), "Could not bind net.p2pdisabled", err)
	}

	startCmd.Flags().Bool(
		"tls", cfg.API.TLS,
		"Enable serving the API over https",
	)
	err = viper.BindPFlag("api.tls", startCmd.Flags().Lookup("tls"))
	if err != nil {
		log.FeedbackFatalE(context.Background(), "Could not bind api.tls", err)
	}

	startCmd.Flags().String(
		"pubkeypath", cfg.API.PubKeyPath,
		"Path to the public key for tls",
	)
	err = viper.BindPFlag("api.pubkeypath", startCmd.Flags().Lookup("pubkeypath"))
	if err != nil {
		log.FeedbackFatalE(context.Background(), "Could not bind api.pubkeypath", err)
	}

	startCmd.Flags().String(
		"privkeypath", cfg.API.PrivKeyPath,
		"Path to the private key for tls",
	)
	err = viper.BindPFlag("api.privkeypath", startCmd.Flags().Lookup("privkeypath"))
	if err != nil {
		log.FeedbackFatalE(context.Background(), "Could not bind api.privkeypath", err)
	}

	startCmd.Flags().String(
		"email", cfg.API.Email,
		"Email address used by the CA for notifications",
	)
	err = viper.BindPFlag("api.email", startCmd.Flags().Lookup("email"))
	if err != nil {
		log.FeedbackFatalE(context.Background(), "Could not bind api.email", err)
	}

	rootCmd.AddCommand(startCmd)
}

type defraInstance struct {
	node   *node.Node
	db     client.DB
	server *httpapi.Server
}

func (di *defraInstance) close(ctx context.Context) {
	if di.node != nil {
		if err := di.node.Close(); err != nil {
			log.FeedbackInfo(
				ctx,
				"The node could not be closed successfully",
				logging.NewKV("Error", err.Error()),
			)
		}
	}
	di.db.Close(ctx)
	if err := di.server.Close(); err != nil {
		log.FeedbackInfo(
			ctx,
			"The server could not be closed successfully",
			logging.NewKV("Error", err.Error()),
		)
	}
}

func start(ctx context.Context) (*defraInstance, error) {
	log.FeedbackInfo(ctx, "Starting DefraDB service...")

	var rootstore ds.Batching

	var err error
	if cfg.Datastore.Store == badgerDatastoreName {
		log.FeedbackInfo(
			ctx,
			"Opening badger store",
			logging.NewKV("Path", cfg.Datastore.Badger.Path),
		)
		rootstore, err = badgerds.NewDatastore(
			cfg.Datastore.Badger.Path,
			cfg.Datastore.Badger.Options,
		)
	} else if cfg.Datastore.Store == "memory" {
		log.FeedbackInfo(ctx, "Building new memory store")
		opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
		rootstore, err = badgerds.NewDatastore("", &opts)
	}

	if err != nil {
		return nil, errors.Wrap("failed to open datastore", err)
	}

	var options []db.Option

	// check for p2p
	if !cfg.Net.P2PDisabled {
		options = append(options, db.WithUpdateEvents())
	}

	db, err := db.NewDB(ctx, rootstore, options...)
	if err != nil {
		return nil, errors.Wrap("failed to create database", err)
	}

	// init the p2p node
	var n *node.Node
	if !cfg.Net.P2PDisabled {
		log.FeedbackInfo(ctx, "Starting P2P node", logging.NewKV("P2P address", cfg.Net.P2PAddress))
		n, err = node.NewNode(
			ctx,
			db,
			cfg.NodeConfig(),
		)
		if err != nil {
			db.Close(ctx)
			return nil, errors.Wrap("failed to start P2P node", err)
		}

		// parse peers and bootstrap
		if len(cfg.Net.Peers) != 0 {
			log.Debug(ctx, "Parsing bootstrap peers", logging.NewKV("Peers", cfg.Net.Peers))
			addrs, err := netutils.ParsePeers(strings.Split(cfg.Net.Peers, ","))
			if err != nil {
				return nil, errors.Wrap(fmt.Sprintf("failed to parse bootstrap peers %v", cfg.Net.Peers), err)
			}
			log.Debug(ctx, "Bootstrapping with peers", logging.NewKV("Addresses", addrs))
			n.Boostrap(addrs)
		}

		if err := n.Start(); err != nil {
			if e := n.Close(); e != nil {
				err = errors.Wrap(fmt.Sprintf("failed to close node: %v", e.Error()), err)
			}
			db.Close(ctx)
			return nil, errors.Wrap("failed to start P2P listeners", err)
		}

		MtcpAddr, err := ma.NewMultiaddr(cfg.Net.TCPAddress)
		if err != nil {
			return nil, errors.Wrap("failed to parse multiaddress", err)
		}
		addr, err := netutils.TCPAddrFromMultiAddr(MtcpAddr)
		if err != nil {
			return nil, errors.Wrap("failed to parse TCP address", err)
		}

		rpcTimeoutDuration, err := cfg.Net.RPCTimeoutDuration()
		if err != nil {
			return nil, errors.Wrap("failed to parse RPC timeout duration", err)
		}

		server := grpc.NewServer(grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: rpcTimeoutDuration,
		}))
		tcplistener, err := gonet.Listen("tcp", addr)
		if err != nil {
			return nil, errors.Wrap(fmt.Sprintf("failed to listen on TCP address %v", addr), err)
		}

		netService := netapi.NewService(n.Peer)

		go func() {
			log.FeedbackInfo(ctx, "Started RPC server", logging.NewKV("Address", addr))
			netpb.RegisterServiceServer(server, netService)
			if err := server.Serve(tcplistener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
				log.FeedbackFatalE(ctx, "Failed to start RPC server", err)
			}
		}()
	}

	rootDir, _, err := config.GetRootDir(rootDirParam)
	if err != nil {
		return nil, errors.Wrap("failed to get root dir", err)
	}

	sOpt := []func(*httpapi.Server){
		httpapi.WithAddress(cfg.API.Address),
		httpapi.WithRootDir(rootDir),
	}

	if n != nil {
		sOpt = append(sOpt, httpapi.WithPeerID(n.PeerID().String()))
	}

	if cfg.API.TLS {
		sOpt = append(
			sOpt,
			httpapi.WithSelfSignedCert(cfg.API.PubKeyPath, cfg.API.PrivKeyPath),
			httpapi.WithCAEmail(cfg.API.Email),
		)
	}

	s := httpapi.NewServer(db, sOpt...)
	if err := s.Listen(ctx); err != nil {
		return nil, errors.Wrap(fmt.Sprintf("failed to listen on TCP address %v", s.Addr), err)
	}

	// run the server in a separate goroutine
	go func() {
		log.FeedbackInfo(
			ctx,
			fmt.Sprintf(
				"Providing HTTP API at %s%s. Use the GraphQL query endpoint at %s%s/graphql ",
				cfg.API.AddressToURL(),
				httpapi.RootPath,
				cfg.API.AddressToURL(),
				httpapi.RootPath,
			),
		)
		if err := s.Run(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.FeedbackErrorE(ctx, "Failed to run the HTTP server", err)
			if n != nil {
				if err := n.Close(); err != nil {
					log.FeedbackErrorE(ctx, "Failed to close node", err)
				}
			}
			db.Close(ctx)
			os.Exit(1)
		}
	}()

	return &defraInstance{
		node:   n,
		db:     db,
		server: s,
	}, nil
}

// wait waits for an interrupt signal to close the program.
func wait(ctx context.Context, di *defraInstance) {
	// setup signal handlers
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	<-signalCh
	log.FeedbackInfo(ctx, "Received interrupt; closing database...")
	di.close(ctx)
	os.Exit(0)
}
