// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package core

import (
	"context"
)

// ReplicatedData is a data type that allows concurrent writers to deterministically merge other
// replicated data so as to converge on the same state.
type ReplicatedData interface {
	Merge(ctx context.Context, other Delta) error
}
