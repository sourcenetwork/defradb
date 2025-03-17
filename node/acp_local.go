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

	"github.com/sourcenetwork/defradb/acp"
	"github.com/sourcenetwork/immutable"
)

const LocalACPType ACPType = "local"

func init() {
	constructor := func(ctx context.Context, options *ACPOptions) (immutable.Option[acp.ACP], error) {
		acpLocal := acp.NewLocalACP()
		acpLocal.Init(ctx, options.path)
		return immutable.Some[acp.ACP](acpLocal), nil
	}
	acpConstructors[LocalACPType] = constructor
	acpConstructors[DefaultACPType] = constructor
}
