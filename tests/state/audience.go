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

//go:build !js

package state

import (
	"strings"

	"github.com/sourcenetwork/immutable"
)

// hostedClient is satisfied by any client wrapper that fronts a node with an
// HTTP host (the http, cli and external wrappers). Matched structurally
// rather than by importing those packages' concrete types, to avoid an
// import cycle (tests/clients/external imports tests/state in its tests).
type hostedClient interface {
	Host() string
}

func GetNodeAudience(s *State, nodeIndex int) immutable.Option[string] {
	if nodeIndex >= len(s.Nodes) {
		return immutable.None[string]()
	}
	if client, ok := s.Nodes[nodeIndex].Client.(hostedClient); ok {
		return immutable.Some(strings.TrimPrefix(client.Host(), "http://"))
	}

	return immutable.None[string]()
}
