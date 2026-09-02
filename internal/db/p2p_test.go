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
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestDB_Connect_WithoutP2P_ReturnsError tests that an error will be returned if 
// Conneect is called with P2P disabled.
func TestDB_Connect_WithoutP2P_ReturnsError(t *testing.T) {
	db := &DB{} // P2P is nil

	err := db.Connect(context.Background(), []string{"/ip4/0.0.0.0/tcp/9171/p2p/12D3KooWBogusPeerID"})
	require.ErrorIs(t, err, ErrNoP2P)
}