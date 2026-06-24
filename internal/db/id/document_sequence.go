// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package id

import (
	"context"

	"github.com/sourcenetwork/defradb/internal/db/sequence"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// NextDocShortID returns the next node-unique document storage ID.
func NextDocShortID(ctx context.Context) (uint64, error) {
	seq, err := sequence.Get(ctx, keys.NewDocIDSequenceKey())
	if err != nil {
		return 0, err
	}

	next, err := seq.Next(ctx)
	if err != nil {
		return 0, err
	}
	return next, nil
}
