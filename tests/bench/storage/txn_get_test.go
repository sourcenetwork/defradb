// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package storage

import (
	"context"
	"fmt"
	"testing"
)

func Benchmark_Storage_Simple_Txn_Read_Sync_1_1(b *testing.B) {
	for _, vsz := range valueSize {
		b.Run(fmt.Sprintf("ValueSize:%04d", vsz), func(b *testing.B) {
			ctx := context.Background()
			err := runStorageBenchTxnGet(b, ctx, vsz, 1, 1, true)
			if err != nil {
				b.Fatal(err)
			}
		})
	}
}

func Benchmark_Storage_Simple_Txn_Read_Sync_2_2(b *testing.B) {
	for _, vsz := range valueSize {
		b.Run(fmt.Sprintf("ValueSize:%04d", vsz), func(b *testing.B) {
			ctx := context.Background()
			err := runStorageBenchTxnGet(b, ctx, vsz, 2, 2, true)
			if err != nil {
				b.Fatal(err)
			}
		})
	}
}

func Benchmark_Storage_Simple_Txn_Read_Sync_10_10(b *testing.B) {
	for _, vsz := range valueSize {
		b.Run(fmt.Sprintf("ValueSize:%04d", vsz), func(b *testing.B) {
			ctx := context.Background()
			err := runStorageBenchTxnGet(b, ctx, vsz, 10, 10, true)
			if err != nil {
				b.Fatal(err)
			}
		})
	}
}

func Benchmark_Storage_Simple_Txn_Read_Sync_100_100(b *testing.B) {
	for _, vsz := range valueSize {
		b.Run(fmt.Sprintf("ValueSize:%04d", vsz), func(b *testing.B) {
			ctx := context.Background()
			err := runStorageBenchTxnGet(b, ctx, vsz, 100, 100, true)
			if err != nil {
				b.Fatal(err)
			}
		})
	}
}
