// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"bytes"
	"context"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-car/v2"
	"github.com/ipld/go-car/v2/storage"
)

// savedBlock holds a CID and raw bytes collected during document save.
type savedBlock struct {
	cid  cid.Cid
	data []byte
}

// generateCAR creates a CARv1 archive from collected blocks.
func generateCAR(ctx context.Context, root cid.Cid, blocks []savedBlock) ([]byte, error) {
	var buf bytes.Buffer
	writer, err := storage.NewWritable(&buf, []cid.Cid{root}, car.WriteAsCarV1(true))
	if err != nil {
		return nil, err
	}
	for _, b := range blocks {
		if err := writer.Put(ctx, b.cid.KeyString(), b.data); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}
