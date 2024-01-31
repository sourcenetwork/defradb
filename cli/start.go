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
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	badger "github.com/sourcenetwork/badger/v4"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/sourcenetwork/defradb/client"
	ds "github.com/sourcenetwork/defradb/datastore"
	badgerds "github.com/sourcenetwork/defradb/datastore/badger/v4"
	"github.com/sourcenetwork/defradb/db"
	"github.com/sourcenetwork/defradb/errors"
	httpapi "github.com/sourcenetwork/defradb/http"
	"github.com/sourcenetwork/defradb/logging"
	"github.com/sourcenetwork/defradb/net"
	netutils "github.com/sourcenetwork/defradb/net/utils"
)

const badgerDatastoreName = "badger"

func MakeStartCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "start",
		Short: "Start a DefraDB node",
		Long:  "Start a DefraDB node.",
		// Load the root config if it exists, otherwise create it.
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return setConfigContext(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := mustGetConfigContext(cmd)

			di, err := start(cmd.Context(), cfg)
			if err != nil {
				return err
			}

			return wait(cmd.Context(), di)
		},
	}

	return cmd
}

type defraInstance struct {
	node   *net.Node
	db     client.DB
	server *httpapi.Server
}

func (di *defraInstance) close(ctx context.Context) {
	if di.node != nil {
		di.node.Close()
	} else {
		di.db.Close()
	}
	if err := di.server.Close(); err != nil {
		log.FeedbackInfo(
			ctx,
			"The server could not be closed successfully",
			logging.NewKV("Error", err.Error()),
		)
	}
}

func start(ctx context.Context, cfg *viper.Viper) (*defraInstance, error) {
	log.FeedbackInfo(ctx, "Starting DefraDB service...")

	var rootstore ds.RootStore

	var err error
	if cfg.Datastore.Store == badgerDatastoreName {
		log.FeedbackInfo(ctx, "Opening badger store", logging.NewKV("Path", cfg.Datastore.Badger.Path))
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

	options := []db.Option{
		db.WithUpdateEvents(),
		db.WithMaxRetries(cfg.Datastore.MaxTxnRetries),
	}

	db, err := db.NewDB(ctx, rootstore, options...)
	if err != nil {
		return nil, errors.Wrap("failed to create database", err)
	}

	// init the p2p node
	var node *net.Node
	if !cfg.GetBool("net.p2pdisabled") {
		nodeOpts := []net.NodeOpt{
			net.WithListenAddresses(cfg.GetStringSlice("net.p2paddresses")...),
			net.WithEnablePubSub(cfg.GetBool("net.pubsubenabled")),
			net.WithEnableRelay(cfg.GetBool("net.relayenabled")),
		}
		if cfg.GetString("datastore.store") == badgerDatastoreName {
			// It would be ideal to not have the key path tied to the datastore.
			// Running with memory store mode will always generate a random key.
			// Adding support for an ephemeral mode and moving the key to the
			// config would solve both of these issues.
			key, err := loadOrGeneratePrivateKey(filepath.Join(cfg.Rootdir, "data", "key"))
			if err != nil {
				return nil, err
			}
			nodeOpts = append(nodeOpts, net.WithPrivateKey(key))
		}
		log.FeedbackInfo(ctx, "Starting P2P node", logging.NewKV("P2P addresses", cfg.Net.P2PAddresses))
		node, err = net.NewNode(ctx, db, nodeOpts...)
		if err != nil {
			db.Close()
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
			node.Bootstrap(addrs)
		}

		if err := node.Start(); err != nil {
			node.Close()
			return nil, errors.Wrap("failed to start P2P listeners", err)
		}
	}

	sOpt := []func(*httpapi.Server){
		httpapi.WithAddress(cfg.API.Address),
		httpapi.WithRootDir(cfg.Rootdir),
		httpapi.WithAllowedOrigins(cfg.API.AllowedOrigins...),
	}

	if cfg.API.TLS {
		sOpt = append(
			sOpt,
			httpapi.WithTLS(),
			httpapi.WithSelfSignedCert(cfg.API.PubKeyPath, cfg.API.PrivKeyPath),
			httpapi.WithCAEmail(cfg.API.Email),
		)
	}

	var server *httpapi.Server
	if node != nil {
		server, err = httpapi.NewServer(node, sOpt...)
	} else {
		server, err = httpapi.NewServer(db, sOpt...)
	}
	if err != nil {
		return nil, errors.Wrap("failed to create http server", err)
	}
	if err := server.Listen(ctx); err != nil {
		return nil, errors.Wrap(fmt.Sprintf("failed to listen on TCP address %v", server.Addr), err)
	}
	// save the address on the config in case the port number was set to random
	cfg.API.Address = server.AssignedAddr()

	// run the server in a separate goroutine
	go func() {
		log.FeedbackInfo(ctx, fmt.Sprintf("Providing HTTP API at %s.", cfg.API.AddressToURL()))
		if err := server.Run(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.FeedbackErrorE(ctx, "Failed to run the HTTP server", err)
			if node != nil {
				node.Close()
			} else {
				db.Close()
			}
			os.Exit(1)
		}
	}()

	return &defraInstance{
		node:   node,
		db:     db,
		server: server,
	}, nil
}

// wait waits for an interrupt signal to close the program.
func wait(ctx context.Context, di *defraInstance) error {
	// setup signal handlers
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		log.FeedbackInfo(ctx, "Received context cancellation; closing database...")
		di.close(ctx)
		return ctx.Err()
	case <-signalCh:
		log.FeedbackInfo(ctx, "Received interrupt; closing database...")
		di.close(ctx)
		return ctx.Err()
	}
}
