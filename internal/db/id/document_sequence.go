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
	"math"

	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/db/sequence"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// NextDocShortID returns the next local document storage ID for the collection.
func NextDocShortID(ctx context.Context, collectionShortID uint32) (uint32, error) {
	seq, err := sequence.Get(ctx, keys.NewDocIDSequenceKey(collectionShortID))
	if err != nil {
		return 0, err
	}

	next, err := seq.Next(ctx)
	if err != nil {
		return 0, err
	}
	if next > math.MaxUint32 {
		return 0, errors.New("local document short ID sequence exhausted")
	}
	return uint32(next), nil
}
