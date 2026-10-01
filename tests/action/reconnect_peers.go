// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package action

import (
	"context"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/state"
	"github.com/sourcenetwork/immutable"
)

// ReconnectPeers reconnects all the peers for all the nodes.
//
// It is useful to run if restarting pub-sub peer nodes.
// Replicators do not require this action and will automatically
// reconnect on their own.
type ReconnectPeers struct {
	stateful
}

var _ Action = (*ReconnectPeers)(nil)
var _ Stateful = (*ReconnectPeers)(nil)

func (a *ReconnectPeers) Execute() {
	nodeIDs, nodes := getNodesWithIDs(immutable.None[int](), a.s.Nodes)
	for sourceIndex, sourceNode := range nodes {
		sourceNodeID := nodeIDs[sourceIndex]
		// Inject every source node's identity into the context while refreshing so the [Connect] & [PeerInfo]
		// call doesn't fail due to lack of authorization(s) if NAC is enabled.
		nodeIdentity := NodeIdentity(sourceNodeID)
		sourceOpts := options.PeerInfo()
		sourceIdent := getIdentityForRequestSpecificToNode(a.s, nodeIdentity, sourceNodeID)
		if sourceIdent.HasValue() {
			sourceOpts.SetIdentity(sourceIdent.Value())
		}

		for targetIndex := range sourceNode.P2P.Connections {
			targetNode := nodes[targetIndex]
			targetNodeID := nodeIDs[targetIndex]
			// Inject target node's identity into the context to bypass NAC for the gated [PeerInfo] operation,
			// otherwise due to lack of authorization(s) we might not be able to see the peer addresses at all.
			targetOpts := options.PeerInfo()
			targetIdent := getIdentityForRequestSpecificToNode(a.s, NodeIdentity(targetNodeID), targetNodeID)
			if targetIdent.HasValue() {
				targetOpts.SetIdentity(targetIdent.Value())
			}
			sourceAddresses, err := sourceNode.PeerInfo(a.s.Ctx, sourceOpts)
			require.NoError(a.s.T, err)
			targetAddresses, err := targetNode.PeerInfo(a.s.Ctx, targetOpts)
			require.NoError(a.s.T, err)

			log.InfoContext(a.s.Ctx, "Connect peers",
				corelog.Any("Source", sourceAddresses),
				corelog.Any("Target", targetAddresses),
			)

			opt := options.WithIdentity(options.Connect(),
				getIdentityForRequestSpecificToNode(a.s, nodeIdentity, sourceNodeID))
			err = connectWithRetry(a.s.Ctx, sourceNode, targetAddresses, opt)
			require.NoError(a.s.T, err)
		}
	}
}

// connectWithRetry attempts to connect to target addresses with retry logic
// to handle transient connection failures.
func connectWithRetry(
	ctx context.Context,
	node *state.NodeState,
	targetAddresses []string,
	opt options.Enumerable[options.ConnectOptions],
) error {
	const maxRetries = 5
	const retryDelay = 50 * time.Millisecond

	var lastErr error
	for attempt := range maxRetries {
		lastErr = node.Connect(ctx, targetAddresses, opt)
		if lastErr == nil {
			return nil
		}
		if attempt < maxRetries-1 {
			time.Sleep(retryDelay)
		}
	}
	return lastErr
}
