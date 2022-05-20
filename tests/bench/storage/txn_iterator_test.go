// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package storage

import (
	"context"
	"fmt"
	"testing"
)

func Benchmark_Storage_Simple_Txn_Iterator_Sync_1_1_1(b *testing.B) {
	for _, vsz := range valueSize {
		b.Run(fmt.Sprintf("ValueSize:%04d", vsz), func(b *testing.B) {
			ctx := context.Background()
			err := runStorageBenchTxnIterator(b, ctx, vsz, 1, 1, 1, true)
			if err != nil {
				b.Fatal(err)
			}
		})
	}
}

func Benchmark_Storage_Simple_Txn_Iterator_Sync_2_1_2(b *testing.B) {
	for _, vsz := range valueSize {
		b.Run(fmt.Sprintf("ValueSize:%04d", vsz), func(b *testing.B) {
			ctx := context.Background()
			err := runStorageBenchTxnIterator(b, ctx, vsz, 2, 1, 2, true)
			if err != nil {
				b.Fatal(err)
			}
		})
	}
}

func Benchmark_Storage_Simple_Txn_Iterator_Sync_10_1_10(b *testing.B) {
	for _, vsz := range valueSize {
		b.Run(fmt.Sprintf("ValueSize:%04d", vsz), func(b *testing.B) {
			ctx := context.Background()
			err := runStorageBenchTxnIterator(b, ctx, vsz, 10, 1, 10, true)
			if err != nil {
				b.Fatal(err)
			}
		})
	}
}

func Benchmark_Storage_Simple_Txn_Iterator_Sync_100_1_100(b *testing.B) {
	for _, vsz := range valueSize {
		b.Run(fmt.Sprintf("ValueSize:%04d", vsz), func(b *testing.B) {
			ctx := context.Background()
			err := runStorageBenchTxnIterator(b, ctx, vsz, 100, 1, 100, true)
			if err != nil {
				b.Fatal(err)
			}
		})
	}
}
