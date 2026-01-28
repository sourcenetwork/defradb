// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

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
