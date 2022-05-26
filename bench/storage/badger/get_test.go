package badger

import (
	"context"
	"fmt"
	"testing"
)

var (
	valueSize = []int{
		256, //128, 256, 512, 1024,
	}
)

func Benchmark_Badger_Simple_Read_Sync_1_1(b *testing.B) {
	for _, vsz := range valueSize {
		b.Run(fmt.Sprintf("ValueSize:%04d", vsz), func(b *testing.B) {
			ctx := context.Background()
			err := runBadgerBenchGet(b, ctx, vsz, 1, 1, true)
			if err != nil {
				b.Fatal(err)
			}
		})
	}
}

func Benchmark_Badger_Simple_Read_Sync_10_1(b *testing.B) {
	for _, vsz := range valueSize {
		b.Run(fmt.Sprintf("ValueSize:%04d", vsz), func(b *testing.B) {
			ctx := context.Background()
			err := runBadgerBenchGet(b, ctx, vsz, 10, 1, true)
			if err != nil {
				b.Fatal(err)
			}
		})
	}
}

func Benchmark_Badger_Simple_Read_Sync_100_1(b *testing.B) {
	for _, vsz := range valueSize {
		b.Run(fmt.Sprintf("ValueSize:%04d", vsz), func(b *testing.B) {
			ctx := context.Background()
			err := runBadgerBenchGet(b, ctx, vsz, 100, 1, true)
			if err != nil {
				b.Fatal(err)
			}
		})
	}
}

func Benchmark_Badger_Simple_Read_Sync_1000_1(b *testing.B) {
	for _, vsz := range valueSize {
		b.Run(fmt.Sprintf("ValueSize:%04d", vsz), func(b *testing.B) {
			ctx := context.Background()
			err := runBadgerBenchGet(b, ctx, vsz, 1000, 1, true)
			if err != nil {
				b.Fatal(err)
			}
		})
	}
}

func Benchmark_Badger_Simple_Read_Sync_100000_1(b *testing.B) {
	for _, vsz := range valueSize {
		b.Run(fmt.Sprintf("ValueSize:%04d", vsz), func(b *testing.B) {
			ctx := context.Background()
			err := runBadgerBenchGet(b, ctx, vsz, 100000, 1, true)
			if err != nil {
				b.Fatal(err)
			}
		})
	}
}

func Benchmark_Badger_Simple_Read_Sync_10_10(b *testing.B) {
	for _, vsz := range valueSize {
		b.Run(fmt.Sprintf("ValueSize:%04d", vsz), func(b *testing.B) {
			ctx := context.Background()
			err := runBadgerBenchGet(b, ctx, vsz, 10, 10, true)
			if err != nil {
				b.Fatal(err)
			}
		})
	}
}

func Benchmark_Badger_Simple_Read_Sync_100_100(b *testing.B) {
	for _, vsz := range valueSize {
		b.Run(fmt.Sprintf("ValueSize:%04d", vsz), func(b *testing.B) {
			ctx := context.Background()
			err := runBadgerBenchGet(b, ctx, vsz, 100, 100, true)
			if err != nil {
				b.Fatal(err)
			}
		})
	}
}

func Benchmark_Badger_Simple_Read_Sync_1000_1000(b *testing.B) {
	for _, vsz := range valueSize {
		b.Run(fmt.Sprintf("ValueSize:%04d", vsz), func(b *testing.B) {
			ctx := context.Background()
			err := runBadgerBenchGet(b, ctx, vsz, 1000, 1000, true)
			if err != nil {
				b.Fatal(err)
			}
		})
	}
}

func Benchmark_Badger_Simple_Read_Sync_100000_1000(b *testing.B) {
	for _, vsz := range valueSize {
		b.Run(fmt.Sprintf("ValueSize:%04d", vsz), func(b *testing.B) {
			ctx := context.Background()
			err := runBadgerBenchGet(b, ctx, vsz, 100000, 1000, true)
			if err != nil {
				b.Fatal(err)
			}
		})
	}
}
func Benchmark_Badger_Simple_Read_Sync_100000_100000(b *testing.B) {
	for _, vsz := range valueSize {
		b.Run(fmt.Sprintf("ValueSize:%04d", vsz), func(b *testing.B) {
			ctx := context.Background()
			err := runBadgerBenchGet(b, ctx, vsz, 100000, 100000, true)
			if err != nil {
				b.Fatal(err)
			}
		})
	}
}

func Benchmark_Badger_Simple_Read2_Sync_10000(b *testing.B) {
	for _, vsz := range valueSize {
		b.Run(fmt.Sprintf("ValueSize:%04d", vsz), func(b *testing.B) {
			ctx := context.Background()
			err := runBadgerBenchGet2(b, ctx, vsz, 10000, false)
			if err != nil {
				b.Fatal(err)
			}
		})
	}
}
