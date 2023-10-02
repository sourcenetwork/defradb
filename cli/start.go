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
	"syscall"

	badger "github.com/dgraph-io/badger/v4"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/config"
	ds "github.com/sourcenetwork/defradb/datastore"
	badgerds "github.com/sourcenetwork/defradb/datastore/badger/v4"
	"github.com/sourcenetwork/defradb/db"
	"github.com/sourcenetwork/defradb/errors"
	httpapi "github.com/sourcenetwork/defradb/http"
	"github.com/sourcenetwork/defradb/logging"
	"github.com/sourcenetwork/defradb/net"
	netpb "github.com/sourcenetwork/defradb/net/pb"
	netutils "github.com/sourcenetwork/defradb/net/utils"
)

const badgerDatastoreName = "badger"

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
			di, err := start(cmd.Context(), cfg)
			if err != nil {
				return err
			}

			return wait(cmd.Context(), di)
		},
	}

	cmd.Flags().String(
		"peers", cfg.Net.Peers,
		"List of peers to connect to",
	)
	err := cfg.BindFlag("net.peers", cmd.Flags().Lookup("peers"))
	if err != nil {
		log.FeedbackFatalE("Could not bind net.peers", err)
	}

	cmd.Flags().Int(
		"max-txn-retries", cfg.Datastore.MaxTxnRetries,
		"Specify the maximum number of retries per transaction",
	)
	err = cfg.BindFlag("datastore.maxtxnretries", cmd.Flags().Lookup("max-txn-retries"))
	if err != nil {
		log.FeedbackFatalE("Could not bind datastore.maxtxnretries", err)
	}

	cmd.Flags().String(
		"store", cfg.Datastore.Store,
		"Specify the datastore to use (supported: badger, memory)",
	)
	err = cfg.BindFlag("datastore.store", cmd.Flags().Lookup("store"))
	if err != nil {
		log.FeedbackFatalE("Could not bind datastore.store", err)
	}

	cmd.Flags().Var(
		&cfg.Datastore.Badger.ValueLogFileSize, "valuelogfilesize",
		"Specify the datastore value log file size (in bytes). In memory size will be 2*valuelogfilesize",
	)
	err = cfg.BindFlag("datastore.badger.valuelogfilesize", cmd.Flags().Lookup("valuelogfilesize"))
	if err != nil {
		log.FeedbackFatalE("Could not bind datastore.badger.valuelogfilesize", err)
	}

	cmd.Flags().String(
		"p2paddr", cfg.Net.P2PAddress,
		"Listener address for the p2p network (formatted as a libp2p MultiAddr)",
	)
	err = cfg.BindFlag("net.p2paddress", cmd.Flags().Lookup("p2paddr"))
	if err != nil {
		log.FeedbackFatalE("Could not bind net.p2paddress", err)
	}

	cmd.Flags().String(
		"tcpaddr", cfg.Net.TCPAddress,
		"Listener address for the tcp gRPC server (formatted as a libp2p MultiAddr)",
	)
	err = cfg.BindFlag("net.tcpaddress", cmd.Flags().Lookup("tcpaddr"))
	if err != nil {
		log.FeedbackFatalE("Could not bind net.tcpaddress", err)
	}

	cmd.Flags().Bool(
		"no-p2p", cfg.Net.P2PDisabled,
		"Disable the peer-to-peer network synchronization system",
	)
	err = cfg.BindFlag("net.p2pdisabled", cmd.Flags().Lookup("no-p2p"))
	if err != nil {
		log.FeedbackFatalE("Could not bind net.p2pdisabled", err)
	}

	cmd.Flags().Bool(
		"tls", cfg.API.TLS,
		"Enable serving the API over https",
	)
	err = cfg.BindFlag("api.tls", cmd.Flags().Lookup("tls"))
	if err != nil {
		log.FeedbackFatalE("Could not bind api.tls", err)
	}

	cmd.Flags().StringArray(
		"allowed-origins", cfg.API.AllowedOrigins,
		"List of origins to allow for CORS requests",
	)
	err = cfg.BindFlag("api.allowed-origins", cmd.Flags().Lookup("allowed-origins"))
	if err != nil {
		log.FeedbackFatalE("Could not bind api.allowed-origins", err)
	}

	cmd.Flags().String(
		"pubkeypath", cfg.API.PubKeyPath,
		"Path to the public key for tls",
	)
	err = cfg.BindFlag("api.pubkeypath", cmd.Flags().Lookup("pubkeypath"))
	if err != nil {
		log.FeedbackFatalE("Could not bind api.pubkeypath", err)
	}

	cmd.Flags().String(
		"privkeypath", cfg.API.PrivKeyPath,
		"Path to the private key for tls",
	)
	err = cfg.BindFlag("api.privkeypath", cmd.Flags().Lookup("privkeypath"))
	if err != nil {
		log.FeedbackFatalE("Could not bind api.privkeypath", err)
	}

	cmd.Flags().String(
		"email", cfg.API.Email,
		"Email address used by the CA for notifications",
	)
	err = cfg.BindFlag("api.email", cmd.Flags().Lookup("email"))
	if err != nil {
		log.FeedbackFatalE("Could not bind api.email", err)
	}
	return cmd
}

type defraInstance struct {
	node   *net.Node
	db     client.DB
	server *httpapi.Server
}

func (di *defraInstance) close() {
	if di.node != nil {
		if err := di.node.Close(); err != nil {
			log.FeedbackInfo(
				"The node could not be closed successfully",
				logging.NewKV("Error", err.Error()),
			)
		}
	}
	di.db.Close()
	if err := di.server.Close(); err != nil {
		log.FeedbackInfo(
			"The server could not be closed successfully",
			logging.NewKV("Error", err.Error()),
		)
	}
}

func start(ctx context.Context, cfg *config.Config) (*defraInstance, error) {
	log.FeedbackInfo("Starting DefraDB service...")

	var rootstore ds.RootStore

	var err error
	if cfg.Datastore.Store == badgerDatastoreName {
		log.FeedbackInfo("Opening badger store", logging.NewKV("Path", cfg.Datastore.Badger.Path))
		rootstore, err = badgerds.NewDatastore(
			cfg.Datastore.Badger.Path,
			cfg.Datastore.Badger.Options,
		)
	} else if cfg.Datastore.Store == "memory" {
		log.FeedbackInfo("Building new memory store")
		opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
		rootstore, err = badgerds.NewDatastore("", &opts)
	}

	if err != nil {
		return nil, errors.Wrap("failed to open datastore", err)
	}

	options := []db.Option{
		db.WithUpdateEvents(),
		db.WithMaxRetries(cfg.Datastore.MaxTxnRetries),
	}

	db, err := db.NewDB(ctx, rootstore, options...)
	if err != nil {
		return nil, errors.Wrap("failed to create database", err)
	}

	// init the p2p node
	var n *net.Node
	if !cfg.Net.P2PDisabled {
		log.FeedbackInfo("Starting P2P node", logging.NewKV("P2P address", cfg.Net.P2PAddress))
		n, err = net.NewNode(
			ctx,
			db,
			net.WithConfig(cfg),
		)
		if err != nil {
			db.Close()
			return nil, errors.Wrap("failed to start P2P node", err)
		}

		// parse peers and bootstrap
		if len(cfg.Net.Peers) != 0 {
			log.Debug("Parsing bootstrap peers", logging.NewKV("Peers", cfg.Net.Peers))
			addrs, err := netutils.ParsePeers(strings.Split(cfg.Net.Peers, ","))
			if err != nil {
				return nil, errors.Wrap(fmt.Sprintf("failed to parse bootstrap peers %v", cfg.Net.Peers), err)
			}
			log.Debug("Bootstrapping with peers", logging.NewKV("Addresses", addrs))
			n.Boostrap(addrs)
		}

		if err := n.Start(); err != nil {
			if e := n.Close(); e != nil {
				err = errors.Wrap(fmt.Sprintf("failed to close node: %v", e.Error()), err)
			}
			db.Close()
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

		server := grpc.NewServer(
			grpc.UnaryInterceptor(
				grpc_middleware.ChainUnaryServer(
					grpc_recovery.UnaryServerInterceptor(),
				),
			),
			grpc.KeepaliveParams(
				keepalive.ServerParameters{
					MaxConnectionIdle: rpcTimeoutDuration,
				},
			),
		)
		tcplistener, err := gonet.Listen("tcp", addr)
		if err != nil {
			return nil, errors.Wrap(fmt.Sprintf("failed to listen on TCP address %v", addr), err)
		}

		go func() {
			log.FeedbackInfo("Started RPC server", logging.NewKV("Address", addr))
			netpb.RegisterCollectionServer(server, n.Peer)
			if err := server.Serve(tcplistener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
				log.FeedbackFatalE("Failed to start RPC server", err)
			}
		}()
	}

	sOpt := []func(*httpapi.Server){
		httpapi.WithAddress(cfg.API.Address),
		httpapi.WithRootDir(cfg.Rootdir),
		httpapi.WithAllowedOrigins(cfg.API.AllowedOrigins...),
	}

	if n != nil {
		sOpt = append(sOpt, httpapi.WithPeerID(n.PeerID().String()))
	}

	if cfg.API.TLS {
		sOpt = append(
			sOpt,
			httpapi.WithTLS(),
			httpapi.WithSelfSignedCert(cfg.API.PubKeyPath, cfg.API.PrivKeyPath),
			httpapi.WithCAEmail(cfg.API.Email),
		)
	}

	s := httpapi.NewServer(db, sOpt...)
	if err := s.Listen(ctx); err != nil {
		return nil, errors.Wrap(fmt.Sprintf("failed to listen on TCP address %v", s.Addr), err)
	}
	// save the address on the config in case the port number was set to random
	cfg.API.Address = s.AssignedAddr()

	// run the server in a separate goroutine
	go func() {
		log.FeedbackInfo(fmt.Sprintf("Providing HTTP API at %s.", cfg.API.AddressToURL()))
		if err := s.Run(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.FeedbackErrorE("Failed to run the HTTP server", err)
			if n != nil {
				if err := n.Close(); err != nil {
					log.FeedbackErrorE("Failed to close node", err)
				}
			}
			db.Close()
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
func wait(ctx context.Context, di *defraInstance) error {
	// setup signal handlers
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		log.FeedbackInfo("Received context cancellation; closing database...")
		di.close()
		return ctx.Err()
	case <-signalCh:
		log.FeedbackInfo("Received interrupt; closing database...")
		di.close()
		return ctx.Err()
	}
}
