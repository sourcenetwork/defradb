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

	"github.com/sourcenetwork/defradb/acp/dac"
	"github.com/sourcenetwork/defradb/client/options"

	"github.com/sourcenetwork/immutable"
)

func init() {
	documentACPConstructors[options.NodeSourceHubDocumentACPType] = func(
		ctx context.Context,
		opts *options.NodeDocumentACPOptions,
	) (immutable.Option[dac.DocumentACP], error) {
		if !opts.Signer.HasValue() {
			return dac.NoDocumentACP, ErrSignerMissingForSourceHubACP
		}
		acpSourceHub, err := dac.NewSourceHubACP(
			opts.SourceHubChainID,
			opts.SourceHubGRPCAddress,
			opts.SourceHubCometRPCAddress,
			opts.Signer.Value(),
		)
		if err != nil {
			return dac.NoDocumentACP, err
		}
		return immutable.Some(acpSourceHub), nil
	}
}
