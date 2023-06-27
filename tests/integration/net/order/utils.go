// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package order

import (
	"context"
	"fmt"
	"strings"
	"testing"

	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/config"
	coreDB "github.com/sourcenetwork/defradb/db"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/logging"
	"github.com/sourcenetwork/defradb/net"
	net_pb "github.com/sourcenetwork/defradb/net/pb"
	netutils "github.com/sourcenetwork/defradb/net/utils"
	testutils "github.com/sourcenetwork/defradb/tests/integration"
)

var (
	log = logging.MustNewLogger("test.net")
)

const (
	userCollectionGQLSchema = `
		type Users {
			Name: String
			Email: String
			Age: Int 
			Height: Float
			Verified: Boolean
		}
	`

	userCollection = "Users"
)

type P2PTestCase struct {
	Query string
	// Configuration parameters for each peer
	NodeConfig []*config.Config

	// List of peers for each net.
	// Only peers with lower index than the node can be used in the list of peers.
	NodePeers map[int][]int

	// List of replicators for each net.
	// Only peers with lower index than the node can be used in the list of peers.
	NodeReplicators map[int][]int

	SeedDocuments        []string
	DocumentsToReplicate []*client.Document

	// node/dockey/values
	Updates          map[int]map[int][]string
	Results          map[int]map[int]map[string]any
	ReplicatorResult map[int]map[string]map[string]any
}

func setupDefraNode(t *testing.T, cfg *config.Config, seeds []string) (*net.Node, []client.DocKey, error) {
	ctx := context.Background()

	log.Info(ctx, "Building new memory store")
	db, err := testutils.NewBadgerMemoryDB(ctx, coreDB.WithUpdateEvents())
	if err != nil {
		return nil, nil, err
	}

	if err := seedSchema(ctx, db); err != nil {
		return nil, nil, err
	}

	// seed the database with a set of documents
	dockeys := []client.DocKey{}
	for _, document := range seeds {
		dockey, err := seedDocument(ctx, db, document)
		require.NoError(t, err)
		dockeys = append(dockeys, dockey)
	}

	// init the p2p node
	var n *net.Node
	log.Info(ctx, "Starting P2P node", logging.NewKV("P2P address", cfg.Net.P2PAddress))
	n, err = net.NewNode(
		ctx,
		db,
		net.WithConfig(cfg),
	)
	if err != nil {
		return nil, nil, errors.Wrap("failed to start P2P node", err)
	}

	// parse peers and bootstrap
	if len(cfg.Net.Peers) != 0 {
		log.Info(ctx, "Parsing bootstrap peers", logging.NewKV("Peers", cfg.Net.Peers))
		addrs, err := netutils.ParsePeers(strings.Split(cfg.Net.Peers, ","))
		if err != nil {
			return nil, nil, errors.Wrap(fmt.Sprintf("failed to parse bootstrap peers %v", cfg.Net.Peers), err)
		}
		log.Info(ctx, "Bootstrapping with peers", logging.NewKV("Addresses", addrs))
		n.Boostrap(addrs)
	}

	if err := n.Start(); err != nil {
		closeErr := n.Close()
		if closeErr != nil {
			return nil, nil, errors.Wrap(fmt.Sprintf("unable to start P2P listeners: %v: problem closing node", err), closeErr)
		}
		return nil, nil, errors.Wrap("unable to start P2P listeners", err)
	}

	cfg.Net.P2PAddress = n.ListenAddrs()[0].String()

	return n, dockeys, nil
}

func seedSchema(ctx context.Context, db client.DB) error {
	_, err := db.AddSchema(ctx, userCollectionGQLSchema)
	return err
}

func seedDocument(ctx context.Context, db client.DB, document string) (client.DocKey, error) {
	col, err := db.GetCollectionByName(ctx, userCollection)
	if err != nil {
		return client.DocKey{}, err
	}

	doc, err := client.NewDocFromJSON([]byte(document))
	if err != nil {
		return client.DocKey{}, err
	}

	err = col.Save(ctx, doc)
	if err != nil {
		return client.DocKey{}, err
	}

	return doc.Key(), nil
}

func saveDocument(ctx context.Context, db client.DB, document *client.Document) error {
	col, err := db.GetCollectionByName(ctx, userCollection)
	if err != nil {
		return err
	}

	return col.Save(ctx, document)
}

func updateDocument(ctx context.Context, db client.DB, dockey client.DocKey, update string) error {
	col, err := db.GetCollectionByName(ctx, userCollection)
	if err != nil {
		return err
	}

	doc, err := getDocument(ctx, db, dockey)
	if err != nil {
		return err
	}

	if err := doc.SetWithJSON([]byte(update)); err != nil {
		return err
	}

	return col.Save(ctx, doc)
}

func getDocument(ctx context.Context, db client.DB, dockey client.DocKey) (*client.Document, error) {
	col, err := db.GetCollectionByName(ctx, userCollection)
	if err != nil {
		return nil, err
	}

	doc, err := col.Get(ctx, dockey, false)
	if err != nil {
		return nil, err
	}
	return doc, err
}

func executeTestCase(t *testing.T, test P2PTestCase) {
	ctx := context.Background()

	dockeys := []client.DocKey{}
	nodes := []*net.Node{}

	for i, cfg := range test.NodeConfig {
		log.Info(ctx, fmt.Sprintf("Setting up node %d", i))
		cfg.Datastore.Badger.Path = t.TempDir()
		if peers, ok := test.NodePeers[i]; ok {
			peerAddresses := []string{}
			for _, p := range peers {
				if p >= len(nodes) {
					log.Info(ctx, "cannot set a peer that hasn't been started. Skipping to next peer")
					continue
				}
				peerAddresses = append(
					peerAddresses,
					fmt.Sprintf("%s/p2p/%s", test.NodeConfig[p].Net.P2PAddress, nodes[p].PeerID()),
				)
			}
			cfg.Net.Peers = strings.Join(peerAddresses, ",")
		}
		n, d, err := setupDefraNode(t, cfg, test.SeedDocuments)
		require.NoError(t, err)

		if i == 0 {
			dockeys = d
		}
		nodes = append(nodes, n)
	}

	//////////////////////////////////////////////////////////////
	//////////////////////////////////////////////////////////////
	// PubSub related test logic

	// wait for peers to connect to each other
	if len(test.NodePeers) > 0 {
		for i, n := range nodes {
			for j, p := range nodes {
				if i == j {
					continue
				}
				log.Info(ctx, fmt.Sprintf("Waiting for node %d to connect with peer %d", i, j))
				err := n.WaitForPubSubEvent(p.PeerID())
				require.NoError(t, err)
				log.Info(ctx, fmt.Sprintf("Node %d connected to peer %d", i, j))
			}
		}
	}

	// update and sync peers
	for n, updateMap := range test.Updates {
		if n >= len(nodes) {
			log.Info(ctx, "cannot update a node that hasn't been started. Skipping to next node")
			continue
		}

		for d, updates := range updateMap {
			for _, update := range updates {
				log.Info(ctx, fmt.Sprintf("Updating node %d with update %d", n, d))
				err := updateDocument(ctx, nodes[n].DB, dockeys[d], update)
				require.NoError(t, err)

				// wait for peers to sync
				for n2, p := range nodes {
					if n2 == n {
						continue
					}
					log.Info(ctx, fmt.Sprintf("Waiting for node %d to sync with peer %d", n2, n))
					err := p.WaitForPushLogByPeerEvent(nodes[n].PeerID())
					require.NoError(t, err)
					log.Info(ctx, fmt.Sprintf("Node %d synced", n2))
				}
			}
		}

		// check that peers actually received the update
		for n2, resultsMap := range test.Results {
			if n2 == n {
				continue
			}
			if n2 >= len(nodes) {
				log.Info(ctx, "cannot check results of a node that hasn't been started. Skipping to next node")
				continue
			}

			for d, results := range resultsMap {
				for field, result := range results {
					doc, err := getDocument(ctx, nodes[n2].DB, dockeys[d])
					require.NoError(t, err)

					val, err := doc.Get(field)
					require.NoError(t, err)

					assert.Equal(t, result, val)
				}
			}
		}
	}

	//////////////////////////////////////////////////////////////
	//////////////////////////////////////////////////////////////
	// Replicator related test logic

	if len(test.NodeReplicators) > 0 {
		for i, n := range nodes {
			if reps, ok := test.NodeReplicators[i]; ok {
				for _, r := range reps {
					addr, err := ma.NewMultiaddr(
						fmt.Sprintf("%s/p2p/%s", test.NodeConfig[r].Net.P2PAddress, nodes[r].PeerID()),
					)
					require.NoError(t, err)
					_, err = n.Peer.SetReplicator(
						ctx,
						&net_pb.SetReplicatorRequest{
							Addr: addr.Bytes(),
						},
					)
					require.NoError(t, err)
				}
			}
		}
	}

	if len(test.DocumentsToReplicate) > 0 {
		for n, reps := range test.NodeReplicators {
			for _, doc := range test.DocumentsToReplicate {
				err := saveDocument(ctx, nodes[n].DB, doc)
				require.NoError(t, err)
			}
			for _, rep := range reps {
				log.Info(ctx, fmt.Sprintf("Waiting for node %d to sync with peer %d", rep, n))
				err := nodes[rep].WaitForPushLogByPeerEvent(nodes[n].PeerID())
				require.NoError(t, err)
				log.Info(ctx, fmt.Sprintf("Node %d synced", rep))

				for dockey, results := range test.ReplicatorResult[rep] {
					for field, result := range results {
						d, err := client.NewDocKeyFromString(dockey)
						require.NoError(t, err)

						doc, err := getDocument(ctx, nodes[rep].DB, d)
						require.NoError(t, err)

						val, err := doc.Get(field)
						require.NoError(t, err)

						assert.Equal(t, result, val)
					}
				}
			}
		}
	}

	// clean up
	for _, n := range nodes {
		if err := n.Close(); err != nil {
			log.Info(ctx, "node not closing as expected", logging.NewKV("Error", err.Error()))
		}
		n.DB.Close(ctx)
	}
}

const randomMultiaddr = "/ip4/0.0.0.0/tcp/0"

func randomNetworkingConfig() *config.Config {
	cfg := config.DefaultConfig()
	cfg.Net.P2PAddress = randomMultiaddr
	cfg.Net.RPCAddress = "0.0.0.0:0"
	cfg.Net.TCPAddress = randomMultiaddr
	return cfg
}
