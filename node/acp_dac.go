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

// documentACPConstructors is a map of document ACP types to acp implementations.
//
// It is populated by the `init` functions in the implementation-specific files - this
// allows it's population to be managed by build flags.
var documentACPConstructors = map[options.NodeDocumentACPType]func(
	context.Context,
	*options.NodeDocumentACPOptions,
) (immutable.Option[dac.DocumentACP], error){}

// NewDocumentACP returns a new ACP module with the given options.
//
// Document ACP is always enabled and can not be disabled; an unspecified type
// defaults to the local implementation.
func NewDocumentACP(
	ctx context.Context,
	opts *options.NodeDocumentACPOptions,
) (immutable.Option[dac.DocumentACP], error) {
	acpType := opts.DocumentACPType
	if acpType == "" {
		acpType = options.NodeLocalDocumentACPType
	}
	acpConstructor, ok := documentACPConstructors[acpType]
	if ok {
		return acpConstructor(ctx, opts)
	}
	return immutable.None[dac.DocumentACP](), NewErrACPTypeNotSupported(acpType)
}
