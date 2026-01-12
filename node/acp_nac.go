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

	"github.com/sourcenetwork/defradb/client/options"
	acpDB "github.com/sourcenetwork/defradb/internal/db/acp"
)

// NewNodeACP creates a new node ACP with the given options.
func NewNodeACP(ctx context.Context, opts ...*options.NodeACPOptions) (acpDB.NACInfo, error) {
	var opt *options.NodeACPOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	if opt == nil {
		opt = options.NodeACP()
	}

	return acpDB.NewNACInfo(ctx, opt.Path, opt.IsEnabled)
}
