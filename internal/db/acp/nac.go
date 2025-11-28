// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package acp

import (
	"context"

	"github.com/sourcenetwork/defradb/acp/nac"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/immutable"
)

// NodeACPDesc contains node acp specific state information.
type NodeACPDesc struct {
	// Status represents the current state of Node ACP.
	Status client.NACStatus

	// Policy contains the policy information of the current node acp setup.
	//
	// When node access control is in a cleaned state, there will be no policy information.
	//
	// Note: The policy information must be validated at the step that enables node access the
	// very first time to ensure that the registered policy with the node acp system is valid.
	// For example, ensure that it adheres to the resource interface rules for node access control.
	Policy immutable.Option[client.PolicyDescription]
}

// NewNodeACPDesc returns a new [NodeACPDesc] that represents a clean node acp state.
func NewNodeACPDesc() NodeACPDesc {
	return NodeACPDesc{
		Status: client.NACNotConfigured,
		Policy: immutable.None[client.PolicyDescription](),
	}
}

// NACInfo contains the current node acp state information, along with the node acp instance.
type NACInfo struct {
	// NodeACP is the acp system, that is always initialized and started ([Start()] called).
	// The reason for having the system always available is to accommodate edge cases where we
	// need node access control internally even when the admin user might have had disabled it.
	// For example, when node acp was enabled once, but the admin user disabled it temporarily, then
	// to know if the identity that is re-enabling is authorized or not, we need the access control.
	NodeACP *nac.NodeACP

	// NodeACPDesc contains the current node acp specific state and other information.
	NodeACPDesc NodeACPDesc

	// EnabledInConfig is true if specified flag to start node access control for the first time.
	//
	// Note: If node access control is temporarily disabled or is already started, then this
	// config value takes no effect, and is ignored.
	EnabledInConfig bool
}

// NewNACInfo returns a newly contructed [NACInfo] with a clean [NodeACPDesc] state.
func NewNACInfo(ctx context.Context, path string, enabled bool) (NACInfo, error) {
	nodeACP, err := nac.NewNodeACP(path)
	if err != nil {
		return NACInfo{}, err
	}
	// We keep NAC started to have access control ability even when node acp is disabled
	// temporarily as we want to only allow authorized user(s) to re-enable node acp.
	if err := nodeACP.Start(ctx); err != nil {
		return NACInfo{}, err
	}

	nacInfo := NACInfo{
		NodeACP:         &nodeACP,
		NodeACPDesc:     NewNodeACPDesc(),
		EnabledInConfig: enabled,
	}
	return nacInfo, nil
}
