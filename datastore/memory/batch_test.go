// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package memory

import (
	"context"
	"testing"

	ds "github.com/ipfs/go-datastore"
	"github.com/stretchr/testify/assert"
)

// NewBasicBatch, Put, Delete, Commit

func TestNewBatch(t *testing.T) {
	s := NewStore()
	b := NewBasicBatch(s)
	assert.NotNil(t, b)
}

func TestBatchOperations(t *testing.T) {
	ctx := context.Background()

	s := newLoadedStore()
	b, err := s.Batch(ctx)
	if err != nil {
		t.Error(err)
	}

	err = b.Delete(ctx, testKey1)
	if err != nil {
		t.Error(err)
	}

	err = b.Put(ctx, testKey3, testValue3)
	if err != nil {
		t.Error(err)
	}

	err = b.Commit(ctx)
	if err != nil {
		t.Error(err)
	}

	_, err = s.Get(ctx, testKey1)
	assert.ErrorIs(t, err, ds.ErrNotFound)

	resp, err := s.Get(ctx, testKey3)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, testValue3, resp)
}
