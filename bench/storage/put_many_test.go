// +build test
package storage

import (
	"context"
	"fmt"
	"testing"
)

func Benchmark_Storage_Simple_WriteMany_Sync_0_1(b *testing.B) {
	valueSize := []int{
		64, 128, 256, 512, 1024,
	}

	for _, vsz := range valueSize {
		b.Run(fmt.Sprintf("ValueSize:%04d", vsz), func(b *testing.B) {
			ctx := context.Background()
			err := runStorageBenchPutMany(b, ctx, vsz, 0, 1, true)
			if err != nil {
				b.Fatal(err)
			}
		})
	}
}

func Benchmark_Storage_Simple_WriteMany_Sync_0_10(b *testing.B) {
	valueSize := []int{
		64, 128, 256, 512, 1024,
	}

	for _, vsz := range valueSize {
		b.Run(fmt.Sprintf("ValueSize:%04d", vsz), func(b *testing.B) {
			ctx := context.Background()
			err := runStorageBenchPutMany(b, ctx, vsz, 0, 10, true)
			if err != nil {
				b.Fatal(err)
			}
		})
	}
}

func Benchmark_Storage_Simple_WriteMany_Sync_0_100(b *testing.B) {
	valueSize := []int{
		64, 128, 256, 512, 1024,
	}

	for _, vsz := range valueSize {
		b.Run(fmt.Sprintf("ValueSize:%04d", vsz), func(b *testing.B) {
			ctx := context.Background()
			err := runStorageBenchPutMany(b, ctx, vsz, 0, 100, true)
			if err != nil {
				b.Fatal(err)
			}
		})
	}
}

func Benchmark_Storage_Simple_WriteMany_Sync_100_1(b *testing.B) {
	valueSize := []int{
		64, 128, 256, 512, 1024,
	}

	for _, vsz := range valueSize {
		b.Run(fmt.Sprintf("ValueSize:%04d", vsz), func(b *testing.B) {
			ctx := context.Background()
			err := runStorageBenchPutMany(b, ctx, vsz, 100, 1, true)
			if err != nil {
				b.Fatal(err)
			}
		})
	}
}

func Benchmark_Storage_Simple_WriteMany_Sync_100_10(b *testing.B) {
	valueSize := []int{
		64, 128, 256, 512, 1024,
	}

	for _, vsz := range valueSize {
		b.Run(fmt.Sprintf("ValueSize:%04d", vsz), func(b *testing.B) {
			ctx := context.Background()
			err := runStorageBenchPutMany(b, ctx, vsz, 100, 10, true)
			if err != nil {
				b.Fatal(err)
			}
		})
	}
}

func Benchmark_Storage_Simple_WriteMany_Sync_100_100(b *testing.B) {
	valueSize := []int{
		64, 128, 256, 512, 1024,
	}

	for _, vsz := range valueSize {
		b.Run(fmt.Sprintf("ValueSize:%04d", vsz), func(b *testing.B) {
			ctx := context.Background()
			err := runStorageBenchPutMany(b, ctx, vsz, 100, 100, true)
			if err != nil {
				b.Fatal(err)
			}
		})
	}
}
