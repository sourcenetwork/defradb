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

	"github.com/sourcenetwork/defradb/acp/dac"
	"github.com/sourcenetwork/defradb/client/options"

	"github.com/sourcenetwork/immutable"
)

func init() {
	constructor := func(
		ctx context.Context,
		opts *options.NodeDocumentACPOptions,
	) (immutable.Option[dac.DocumentACP], error) {
		localDocumentACP, err := dac.NewLocalDocumentACP(opts.Path)
		if err != nil {
			return dac.NoDocumentACP, err
		}

		return immutable.Some(localDocumentACP), nil
	}
	documentACPConstructors[options.NodeLocalDocumentACPType] = constructor
	documentACPConstructors[options.NodeDefaultDocumentACPType] = constructor
}
