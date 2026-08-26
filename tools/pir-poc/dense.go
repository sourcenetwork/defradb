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
	"crypto/subtle"
	"errors"
	"io"
	"math/bits"
)

func querySize(rowCount int) int {
	return (rowCount + 7) / 8
}

func queryShares(
	target int,
	rowCount int,
	serverCount int,
	random io.Reader,
) ([][]byte, error) {
	if target < 0 || target >= rowCount {
		return nil, errors.New("target is outside the table")
	}
	if serverCount < 2 {
		return nil, errors.New("Dense XOR PIR requires at least two servers")
	}

	shareBytes := querySize(rowCount)
	finalShare := make([]byte, shareBytes)
	finalShare[target/8] = 1 << (target % 8)
	shares := make([][]byte, 0, serverCount)
	for range serverCount - 1 {
		share := make([]byte, shareBytes)
		if _, err := io.ReadFull(random, share); err != nil {
			return nil, err
		}
		xorInPlace(finalShare, share)
		shares = append(shares, share)
	}
	return append(shares, finalShare), nil
}

func answerBatch(table []byte, rowCount int, rowBytes int, queries [][]byte) ([][]byte, error) {
	if len(table) != rowCount*rowBytes {
		return nil, errors.New("table dimensions do not match its contents")
	}
	for _, query := range queries {
		if len(query) != querySize(rowCount) {
			return nil, errors.New("query share has the wrong size")
		}
	}

	answers := make([][]byte, len(queries))
	for queryIndex, query := range queries {
		answer := make([]byte, rowBytes)
		for byteIndex, queryByte := range query {
			selected := queryByte
			for selected != 0 {
				bitIndex := bits.TrailingZeros8(selected)
				row := byteIndex*8 + bitIndex
				if row < rowCount {
					start := row * rowBytes
					xorInPlace(answer, table[start:start+rowBytes])
				}
				selected &= selected - 1
			}
		}
		answers[queryIndex] = answer
	}
	return answers, nil
}

func combine(shares [][]byte) ([]byte, error) {
	if len(shares) == 0 {
		return nil, errors.New("no answer shares supplied")
	}
	answer := append([]byte(nil), shares[0]...)
	for _, share := range shares[1:] {
		if len(share) != len(answer) {
			return nil, errors.New("answer share lengths differ")
		}
		xorInPlace(answer, share)
	}
	return answer, nil
}

func xorInPlace(target []byte, share []byte) {
	subtle.XORBytes(target, target, share)
}
