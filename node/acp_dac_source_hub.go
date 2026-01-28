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

	"github.com/sourcenetwork/immutable"
)

const SourceHubDocumentACPType DocumentACPType = "source-hub"

func init() {
	documentACPConstructors[SourceHubDocumentACPType] = func(
		ctx context.Context,
		options *DocumentACPOptions,
	) (immutable.Option[dac.DocumentACP], error) {
		if !options.signer.HasValue() {
			return dac.NoDocumentACP, ErrSignerMissingForSourceHubACP
		}
		acpSourceHub, err := dac.NewSourceHubACP(
			options.sourceHubChainID,
			options.sourceHubGRPCAddress,
			options.sourceHubCometRPCAddress,
			options.signer.Value(),
		)
		if err != nil {
			return dac.NoDocumentACP, err
		}
		return immutable.Some(acpSourceHub), nil
	}
}
