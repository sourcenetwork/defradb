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
	"testing"

	"github.com/sourcenetwork/corelog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/net"
	netutils "github.com/sourcenetwork/defradb/net/utils"
	"github.com/sourcenetwork/defradb/node/db"
	testutils "github.com/sourcenetwork/defradb/tests/integration"
)

var (
	log = corelog.NewLogger("test.net")
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

	// The identity for all requests.
	// TODO-ACP: https://github.com/sourcenetwork/defradb/issues/2366 - Improve in ACP <> P2P implementation
	Identity string

	// Configuration parameters for each peer
	NodeConfig [][]net.NodeOpt

	// List of peers for each net.
	// Only peers with lower index than the node can be used in the list of peers.
	NodePeers map[int][]int

	// List of replicators for each net.
	// Only peers with lower index than the node can be used in the list of peers.
	NodeReplicators map[int][]int

	SeedDocuments        []string
	DocumentsToReplicate []*client.Document

	// node/docID/values
	Updates          map[int]map[int][]string
	Results          map[int]map[int]map[string]any
	ReplicatorResult map[int]map[string]map[string]any
}

func setupDefraNode(
	t *testing.T,
	opts []net.NodeOpt,
	peers []string,
	seeds []string,
) (*net.Node, []client.DocID, error) {
	ctx := context.Background()

	log.InfoContext(ctx, "Building new memory store")
	db, err := testutils.NewBadgerMemoryDB(ctx, db.WithUpdateEvents())
	if err != nil {
		return nil, nil, err
	}

	if err := seedSchema(ctx, db); err != nil {
		return nil, nil, err
	}

	// seed the database with a set of documents
	docIDs := []client.DocID{}
	for _, document := range seeds {
		docID, err := seedDocument(ctx, db, document)
		require.NoError(t, err)
		docIDs = append(docIDs, docID)
	}

	// init the P2P node
	var n *net.Node
	n, err = net.NewNode(ctx, db, opts...)
	if err != nil {
		return nil, nil, errors.Wrap("failed to start P2P node", err)
	}

	// parse peers and bootstrap
	if len(peers) != 0 {
		log.InfoContext(ctx, "Parsing bootstrap peers", corelog.Any("Peers", peers))
		addrs, err := netutils.ParsePeers(peers)
		if err != nil {
			return nil, nil, errors.Wrap(fmt.Sprintf("failed to parse bootstrap peers %v", peers), err)
		}
		log.InfoContext(ctx, "Bootstrapping with peers", corelog.Any("Addresses", addrs))
		n.Bootstrap(addrs)
	}

	log.InfoContext(ctx, "Starting P2P node", corelog.Any("P2P addresses", n.PeerInfo().Addrs))
	if err := n.Start(); err != nil {
		n.Close()
		return nil, nil, errors.Wrap("unable to start P2P listeners", err)
	}

	return n, docIDs, nil
}

func seedSchema(ctx context.Context, db client.DB) error {
	_, err := db.AddSchema(ctx, userCollectionGQLSchema)
	return err
}

func seedDocument(
	ctx context.Context,
	db client.DB,
	document string,
) (client.DocID, error) {
	col, err := db.GetCollectionByName(ctx, userCollection)
	if err != nil {
		return client.DocID{}, err
	}

	doc, err := client.NewDocFromJSON([]byte(document), col.Definition())
	if err != nil {
		return client.DocID{}, err
	}

	err = col.Save(ctx, doc)
	if err != nil {
		return client.DocID{}, err
	}

	return doc.ID(), nil
}

func saveDocument(
	ctx context.Context,
	db client.DB,
	document *client.Document,
) error {
	col, err := db.GetCollectionByName(ctx, userCollection)
	if err != nil {
		return err
	}

	return col.Save(ctx, document)
}

func updateDocument(
	ctx context.Context,
	db client.DB,
	docID client.DocID,
	update string,
) error {
	col, err := db.GetCollectionByName(ctx, userCollection)
	if err != nil {
		return err
	}

	doc, err := getDocument(ctx, db, docID)
	if err != nil {
		return err
	}

	if err := doc.SetWithJSON([]byte(update)); err != nil {
		return err
	}

	return col.Save(ctx, doc)
}

func getDocument(
	ctx context.Context,
	db client.DB,
	docID client.DocID,
) (*client.Document, error) {
	col, err := db.GetCollectionByName(ctx, userCollection)
	if err != nil {
		return nil, err
	}

	doc, err := col.Get(ctx, docID, false)
	if err != nil {
		return nil, err
	}
	return doc, err
}

func executeTestCase(t *testing.T, test P2PTestCase) {
	ctx := context.Background()

	docIDs := []client.DocID{}
	nodes := []*net.Node{}

	for i, cfg := range test.NodeConfig {
		log.InfoContext(ctx, fmt.Sprintf("Setting up node %d", i))
		var peerAddresses []string
		if peers, ok := test.NodePeers[i]; ok {
			for _, p := range peers {
				if p >= len(nodes) {
					log.InfoContext(ctx, "cannot set a peer that hasn't been started. Skipping to next peer")
					continue
				}
				peerInfo := nodes[p].PeerInfo()
				peerAddresses = append(
					peerAddresses,
					fmt.Sprintf("%s/p2p/%s", peerInfo.Addrs[0], peerInfo.ID),
				)
			}
		}
		n, d, err := setupDefraNode(
			t,
			cfg,
			peerAddresses,
			test.SeedDocuments,
		)
		require.NoError(t, err)

		if i == 0 {
			docIDs = d
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
				log.InfoContext(ctx, fmt.Sprintf("Waiting for node %d to connect with peer %d", i, j))
				err := n.WaitForPubSubEvent(p.PeerID())
				require.NoError(t, err)
				log.InfoContext(ctx, fmt.Sprintf("Node %d connected to peer %d", i, j))
			}
		}
	}

	// update and sync peers
	for n, updateMap := range test.Updates {
		if n >= len(nodes) {
			log.InfoContext(ctx, "cannot update a node that hasn't been started. Skipping to next node")
			continue
		}

		for d, updates := range updateMap {
			for _, update := range updates {
				log.InfoContext(ctx, fmt.Sprintf("Updating node %d with update %d", n, d))
				err := updateDocument(
					ctx,
					nodes[n].DB,
					docIDs[d],
					update,
				)
				require.NoError(t, err)

				// wait for peers to sync
				for n2, p := range nodes {
					if n2 == n {
						continue
					}
					log.InfoContext(ctx, fmt.Sprintf("Waiting for node %d to sync with peer %d", n2, n))
					err := p.WaitForPushLogByPeerEvent(nodes[n].PeerInfo().ID)
					require.NoError(t, err)
					log.InfoContext(ctx, fmt.Sprintf("Node %d synced", n2))
				}
			}
		}

		// check that peers actually received the update
		for n2, resultsMap := range test.Results {
			if n2 == n {
				continue
			}
			if n2 >= len(nodes) {
				log.InfoContext(ctx, "cannot check results of a node that hasn't been started. Skipping to next node")
				continue
			}

			for d, results := range resultsMap {
				for field, result := range results {
					doc, err := getDocument(
						ctx,
						nodes[n2].DB,
						docIDs[d],
					)
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
					err := n.Peer.SetReplicator(ctx, client.Replicator{
						Info: nodes[r].PeerInfo(),
					})
					require.NoError(t, err)
				}
			}
		}
	}

	if len(test.DocumentsToReplicate) > 0 {
		for n, reps := range test.NodeReplicators {
			for _, doc := range test.DocumentsToReplicate {
				err := saveDocument(
					ctx,
					nodes[n].DB,
					doc,
				)
				require.NoError(t, err)
			}
			for _, rep := range reps {
				log.InfoContext(ctx, fmt.Sprintf("Waiting for node %d to sync with peer %d", rep, n))
				err := nodes[rep].WaitForPushLogByPeerEvent(nodes[n].PeerID())
				require.NoError(t, err)
				log.InfoContext(ctx, fmt.Sprintf("Node %d synced", rep))

				for docID, results := range test.ReplicatorResult[rep] {
					for field, result := range results {
						d, err := client.NewDocIDFromString(docID)
						require.NoError(t, err)

						doc, err := getDocument(
							ctx,
							nodes[rep].DB,
							d,
						)
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
		n.Close()
		n.DB.Close()
	}
}
