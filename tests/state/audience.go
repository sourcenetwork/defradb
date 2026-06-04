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

	"github.com/sourcenetwork/defradb/tests/clients/cli"
	"github.com/sourcenetwork/defradb/tests/clients/http"
)

func GetNodeAudience(s *State, nodeIndex int) immutable.Option[string] {
	if nodeIndex >= len(s.Nodes) {
		return immutable.None[string]()
	}
	switch client := s.Nodes[nodeIndex].Client.(type) {
	case *http.Wrapper:
		return immutable.Some(strings.TrimPrefix(client.Host(), "http://"))
	case *cli.Wrapper:
		return immutable.Some(strings.TrimPrefix(client.Host(), "http://"))
	}

	return immutable.None[string]()
}
