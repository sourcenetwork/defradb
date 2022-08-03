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

var (
	valueSize = []int{
		64, 128, 256, 512, 1024,
	}
)

func Benchmark_Storage_Simple_Read_Sync_1_1(b *testing.B) {
	for _, vsz := range valueSize {
		b.Run(fmt.Sprintf("ValueSize:%04d", vsz), func(b *testing.B) {
			ctx := context.Background()
			err := runStorageBenchGet(b, ctx, vsz, 1, 1, true)
			if err != nil {
				b.Fatal(err)
			}
		})
	}
}

func Benchmark_Storage_Simple_Read_Sync_1_10(b *testing.B) {
	for _, vsz := range valueSize {
		b.Run(fmt.Sprintf("ValueSize:%04d", vsz), func(b *testing.B) {
			ctx := context.Background()
			err := runStorageBenchGet(b, ctx, vsz, 1, 10, true)
			if err != nil {
				b.Fatal(err)
			}
		})
	}
}

func Benchmark_Storage_Simple_Read_Sync_1_100(b *testing.B) {
	for _, vsz := range valueSize {
		b.Run(fmt.Sprintf("ValueSize:%04d", vsz), func(b *testing.B) {
			ctx := context.Background()
			err := runStorageBenchGet(b, ctx, vsz, 1, 100, true)
			if err != nil {
				b.Fatal(err)
			}
		})
	}
}

func Benchmark_Storage_Simple_Read_Sync_100_1(b *testing.B) {
	for _, vsz := range valueSize {
		b.Run(fmt.Sprintf("ValueSize:%04d", vsz), func(b *testing.B) {
			ctx := context.Background()
			err := runStorageBenchGet(b, ctx, vsz, 100, 1, true)
			if err != nil {
				b.Fatal(err)
			}
		})
	}
}

func Benchmark_Storage_Simple_Read_Sync_100_10(b *testing.B) {
	for _, vsz := range valueSize {
		b.Run(fmt.Sprintf("ValueSize:%04d", vsz), func(b *testing.B) {
			ctx := context.Background()
			err := runStorageBenchGet(b, ctx, vsz, 100, 10, true)
			if err != nil {
				b.Fatal(err)
			}
		})
	}
}

func Benchmark_Storage_Simple_Read_Sync_100_100(b *testing.B) {
	for _, vsz := range valueSize {
		b.Run(fmt.Sprintf("ValueSize:%04d", vsz), func(b *testing.B) {
			ctx := context.Background()
			err := runStorageBenchGet(b, ctx, vsz, 100, 100, true)
			if err != nil {
				b.Fatal(err)
			}
		})
	}
}
