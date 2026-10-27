// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package node

import (
	"context"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/dac"
	"github.com/sourcenetwork/defradb/client/options"
)

// NewDocumentACP returns a new document ACP module for the given options.
//
// Document ACP is always enabled and can not be disabled; an unspecified type
// defaults to Local DAC. The Remote DAC implementation is chosen at
// build time (see acp_dac_remote.go and acp_dac_remote_js.go).
func NewDocumentACP(
	ctx context.Context,
	opts *options.NodeDocumentACPOptions,
) (immutable.Option[dac.DocumentACP], error) {
	switch opts.DocumentACPType {
	case "", options.NodeLocalDocumentACPType:
		return newLocalDocumentACP(ctx, opts)
	case options.NodeRemoteDocumentACPType:
		return newRemoteDocumentACP(ctx, opts)
	default:
		return immutable.None[dac.DocumentACP](), NewErrACPTypeNotSupported(opts.DocumentACPType)
	}
}
