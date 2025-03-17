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

package node

import (
	"context"

	"github.com/sourcenetwork/defradb/acp"

	"github.com/sourcenetwork/immutable"
)

const SourceHubACPType ACPType = "source-hub"

func init() {
	acpConstructors[SourceHubACPType] = func(ctx context.Context, options *ACPOptions) (immutable.Option[acp.ACP], error) {
		if !options.signer.HasValue() {
			return acp.NoACP, ErrSignerMissingForSourceHubACP
		}
		acpSourceHub, err := acp.NewSourceHubACP(
			options.sourceHubChainID,
			options.sourceHubGRPCAddress,
			options.sourceHubCometRPCAddress,
			options.signer.Value(),
		)
		if err != nil {
			return acp.NoACP, err
		}
		return immutable.Some(acpSourceHub), nil
	}
}
