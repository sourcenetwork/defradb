// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package nac

import (
	"context"

	"github.com/sourcenetwork/defradb/acp"
	"github.com/sourcenetwork/defradb/acp/local"
	acpTypes "github.com/sourcenetwork/defradb/acp/types"
)

const localStoreName = "local_node_acp"

var _ acp.ACPSystemClient = (*NodeACP)(nil)

// NodeACP represents a node acp (local to the node) implementation that makes no remote calls.
type NodeACP struct {
	*local.LocalACP
}

func NewNodeACP(pathToStore string) (NodeACP, error) {
	localACP, err := local.NewLocalACP(pathToStore, localStoreName)
	if err != nil {
		return NodeACP{}, err
	}

	return NodeACP{LocalACP: localACP}, nil
}

func (a *NodeACP) ValidateResourceInterface(
	ctx context.Context,
	policyID string,
	resourceName string,
) error {
	return acp.ValidateResourceInterface(
		ctx,
		policyID,
		resourceName,
		acpTypes.NodeACP,
		a,
	)
}
