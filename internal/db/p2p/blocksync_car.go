// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package p2p

import (
	"context"
	"io"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	car "github.com/ipld/go-car/v2"
	carstorage "github.com/ipld/go-car/v2/storage"

	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/datastore"
)

// writeCAR streams the given blocks to w as a CARv1 payload with the provided roots.
//
// CARv1 is a pure streaming format (no trailing index), so the blocks are written one after
// another without buffering the whole payload.
func writeCAR(ctx context.Context, w io.Writer, roots []cid.Cid, blks []blocks.Block) error {
	cw, err := carstorage.NewWritable(w, roots, car.WriteAsCarV1(true))
	if err != nil {
		return NewErrWriteCAR(err)
	}
	for _, b := range blks {
		if err := cw.Put(ctx, b.Cid().KeyString(), b.RawData()); err != nil {
			return NewErrWriteCAR(err)
		}
	}
	if err := cw.Finalize(); err != nil {
		return NewErrWriteCAR(err)
	}
	return nil
}

// ingestCAR reads a CARv1 stream from r and stores each block into blockDst, routing any block
// whose CID is in encCIDs into encDst (the encryption store) instead.
//
// The reader is streamed block-by-block, so an arbitrarily large CAR can be ingested without
// buffering the whole payload. The CAR roots are returned.
func ingestCAR(
	ctx context.Context,
	blockDst datastore.Blockstore,
	encDst datastore.Blockstore,
	encCIDs map[cid.Cid]struct{},
	r io.Reader,
) ([]cid.Cid, error) {
	br, err := car.NewBlockReader(r)
	if err != nil {
		return nil, NewErrReadCAR(err)
	}
	for {
		b, err := br.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, NewErrReadCAR(err)
		}

		dst := blockDst
		if _, ok := encCIDs[b.Cid()]; ok {
			dst = encDst
		}
		if err := dst.Put(ctx, b); err != nil {
			return nil, NewErrIngestCARBlock(err, b.Cid().String())
		}
	}
	return br.Roots, nil
}
