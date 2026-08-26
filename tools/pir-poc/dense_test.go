// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package main

import (
	"bytes"
	mathrand "math/rand/v2"
	"testing"
)

func TestDenseXORRecoversRequestedRowAtAnyServerCount(t *testing.T) {
	table, err := benchmarkTable(32, 32, 1)
	if err != nil {
		t.Fatal(err)
	}
	random := mathrand.NewChaCha8([32]byte{7})
	for serverCount := 2; serverCount <= 6; serverCount++ {
		queries, err := queryShares(7, 32, serverCount, random)
		if err != nil {
			t.Fatal(err)
		}
		answers := make([][]byte, serverCount)
		for server := range serverCount {
			serverAnswers, err := answerBatch(table, 32, 32, [][]byte{queries[server]})
			if err != nil {
				t.Fatal(err)
			}
			answers[server] = serverAnswers[0]
		}
		answer, err := combine(answers)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(answer, table[7*32:8*32]) {
			t.Fatalf("%d-server answer does not match requested row", serverCount)
		}
	}
}

func TestMalformedInputsAreRejected(t *testing.T) {
	table, err := benchmarkTable(32, 32, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queryShares(32, 32, 2, mathrand.NewChaCha8([32]byte{})); err == nil {
		t.Fatal("out-of-range target was accepted")
	}
	if _, err := queryShares(1, 32, 1, mathrand.NewChaCha8([32]byte{})); err == nil {
		t.Fatal("one-server Dense XOR was accepted")
	}
	if _, err := answerBatch(table, 32, 32, [][]byte{{0}}); err == nil {
		t.Fatal("short query share was accepted")
	}
	if _, err := combine([][]byte{{1}, {1, 2}}); err == nil {
		t.Fatal("different answer sizes were accepted")
	}
}

func TestTargetSequenceIsStable(t *testing.T) {
	workload := workload{name: "test", rows: 1 << 10, rowBytes: 32, batchSize: 3, samples: 1}
	if got := targets(workload, 2); !equalInts(got, []int{518, 539, 560}) {
		t.Fatalf("unexpected targets: %v", got)
	}
}

func TestDeterministicCorpusMatchesTheRustPOC(t *testing.T) {
	table, err := benchmarkTable(32, 32, 1)
	if err != nil {
		t.Fatal(err)
	}
	if checksum := fnv1a64(table); checksum != 0x3cffaeb69428cab5 {
		t.Fatalf("unexpected corpus checksum: %016x", checksum)
	}
}

func equalInts(left []int, right []int) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
