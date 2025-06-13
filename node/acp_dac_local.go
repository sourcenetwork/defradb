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

	"github.com/sourcenetwork/immutable"
)

const LocalDocumentACPType DocumentACPType = "local"

func init() {
	constructor := func(ctx context.Context, options *DocumentACPOptions) (immutable.Option[dac.DocumentACP], error) {
		localDocumentACP := dac.NewLocalDocumentACP()
		localDocumentACP.Init(ctx, options.path)
		return immutable.Some(localDocumentACP), nil
	}
	documentACPConstructors[LocalDocumentACPType] = constructor
	documentACPConstructors[DefaultDocumentACPType] = constructor
}
