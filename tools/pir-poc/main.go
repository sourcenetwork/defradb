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
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"
	"sort"
	"time"
)

const (
	tableSeed        uint64 = 0x4d595df4d0f33173
	targetStep       uint64 = 0x9e3779b97f4a7c15
	decoyCount              = 100
	minClientSample         = 10 * time.Millisecond
	minServerSample         = 50 * time.Millisecond
	maxClientRepeats        = 1 << 20
)

type report struct {
	Schema         string           `json:"schema"`
	Language       string           `json:"language"`
	Implementation string           `json:"implementation"`
	Profile        string           `json:"profile"`
	TableSeed      uint64           `json:"table_seed"`
	Timing         string           `json:"timing"`
	Workloads      []workloadReport `json:"workloads"`
}

type workloadReport struct {
	Name                 string      `json:"name"`
	Rows                 int         `json:"rows"`
	RowBytes             int         `json:"row_bytes"`
	BatchSize            int         `json:"batch_size"`
	Samples              int         `json:"samples"`
	TableChecksumFNV1A64 string      `json:"table_checksum_fnv1a64"`
	Direct               measurement `json:"direct"`
	Decoy100             measurement `json:"decoy_100"`
	DenseXOR2            measurement `json:"dense_xor_2"`
	DenseXOR3            measurement `json:"dense_xor_3"`
}

type measurement struct {
	ClientQueryP50MS   float64 `json:"client_query_p50_ms"`
	ServerTotalP50MS   float64 `json:"server_total_p50_ms"`
	ClientFinishP50MS  float64 `json:"client_finish_p50_ms"`
	UploadBytes        int     `json:"upload_bytes"`
	DownloadBytes      int     `json:"download_bytes"`
	SourceOperandBytes int     `json:"source_operand_bytes"`
}

type workload struct {
	name      string
	rows      int
	rowBytes  int
	batchSize int
	samples   int
}

type sample struct {
	clientQuery        time.Duration
	serverTotal        time.Duration
	clientFinish       time.Duration
	uploadBytes        int
	downloadBytes      int
	sourceOperandBytes int
}

func main() {
	profile := flag.String("profile", "quick", "benchmark profile: quick or full")
	flag.Parse()

	result, err := run(*profile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(profile string) (report, error) {
	workloads, err := profileWorkloads(profile)
	if err != nil {
		return report{}, err
	}

	result := report{
		Schema:         "defradb-pir-cross-language-v1",
		Language:       "go",
		Implementation: runtime.Version(),
		Profile:        profile,
		TableSeed:      tableSeed,
		Timing:         "p50; client phases run >=10ms, server paths >=50ms; replica work is summed",
		Workloads:      make([]workloadReport, 0, len(workloads)),
	}
	for _, workload := range workloads {
		workloadResult, err := runWorkload(workload)
		if err != nil {
			return report{}, err
		}
		result.Workloads = append(result.Workloads, workloadResult)
	}
	return result, nil
}

func profileWorkloads(profile string) ([]workload, error) {
	switch profile {
	case "quick":
		return []workload{
			{name: "locator", rows: 1 << 18, rowBytes: 96, batchSize: 1, samples: 7},
			{name: "witness", rows: 1 << 15, rowBytes: 2_008, batchSize: 1, samples: 7},
			{name: "batch-16", rows: 1 << 16, rowBytes: 96, batchSize: 16, samples: 5},
		}, nil
	case "full":
		return []workload{
			{name: "locator", rows: 1 << 20, rowBytes: 96, batchSize: 1, samples: 11},
			{name: "witness", rows: 1 << 16, rowBytes: 2_008, batchSize: 1, samples: 11},
			{name: "batch-16", rows: 1 << 18, rowBytes: 96, batchSize: 16, samples: 7},
		}, nil
	default:
		return nil, fmt.Errorf("unknown profile %q; expected quick or full", profile)
	}
}

func runWorkload(workload workload) (workloadReport, error) {
	table, err := benchmarkTable(workload.rows, workload.rowBytes, tableSeed)
	if err != nil {
		return workloadReport{}, err
	}

	// One unreported pass faults pages and exercises every path before sampling.
	if _, err = measureDirect(table, workload, 0); err != nil {
		return workloadReport{}, err
	}
	if _, err = measureDecoys(table, workload, 0); err != nil {
		return workloadReport{}, err
	}
	if _, err = measureDense(table, workload, 2, 0); err != nil {
		return workloadReport{}, err
	}
	if _, err = measureDense(table, workload, 3, 0); err != nil {
		return workloadReport{}, err
	}

	direct := make([]sample, 0, workload.samples)
	decoys := make([]sample, 0, workload.samples)
	dense2 := make([]sample, 0, workload.samples)
	dense3 := make([]sample, 0, workload.samples)
	for sampleIndex := 1; sampleIndex <= workload.samples; sampleIndex++ {
		directSample, err := measureDirect(table, workload, sampleIndex)
		if err != nil {
			return workloadReport{}, err
		}
		decoySample, err := measureDecoys(table, workload, sampleIndex)
		if err != nil {
			return workloadReport{}, err
		}
		dense2Sample, err := measureDense(table, workload, 2, sampleIndex)
		if err != nil {
			return workloadReport{}, err
		}
		dense3Sample, err := measureDense(table, workload, 3, sampleIndex)
		if err != nil {
			return workloadReport{}, err
		}
		direct = append(direct, directSample)
		decoys = append(decoys, decoySample)
		dense2 = append(dense2, dense2Sample)
		dense3 = append(dense3, dense3Sample)
	}

	return workloadReport{
		Name:                 workload.name,
		Rows:                 workload.rows,
		RowBytes:             workload.rowBytes,
		BatchSize:            workload.batchSize,
		Samples:              workload.samples,
		TableChecksumFNV1A64: fmt.Sprintf("%016x", fnv1a64(table)),
		Direct:               medianMeasurement(direct),
		Decoy100:             medianMeasurement(decoys),
		DenseXOR2:            medianMeasurement(dense2),
		DenseXOR3:            medianMeasurement(dense3),
	}, nil
}

func measureDirect(table []byte, workload workload, sampleIndex int) (sample, error) {
	targets := targets(workload, sampleIndex)
	runtime.GC()
	response, serverTotal, err := measureServerRepeated(func() ([][]byte, error) {
		result := make([][]byte, len(targets))
		for index, target := range targets {
			result[index] = append([]byte(nil), row(table, workload, target)...)
		}
		return result, nil
	})
	if err != nil {
		return sample{}, err
	}

	_, clientQuery, err := measureRepeated(func() ([][8]byte, error) {
		query := make([][8]byte, len(targets))
		for index, target := range targets {
			binary.LittleEndian.PutUint64(query[index][:], uint64(target))
		}
		return query, nil
	})
	if err != nil {
		return sample{}, err
	}

	_, clientFinish, err := measureRepeated(func() (struct{}, error) {
		for index, target := range targets {
			if !bytes.Equal(response[index], row(table, workload, target)) {
				return struct{}{}, errors.New("direct answer mismatch")
			}
		}
		keepAlive(response)
		return struct{}{}, nil
	})
	if err != nil {
		return sample{}, err
	}

	return sample{
		clientQuery:        clientQuery,
		serverTotal:        serverTotal,
		clientFinish:       clientFinish,
		uploadBytes:        workload.batchSize * 8,
		downloadBytes:      workload.batchSize * workload.rowBytes,
		sourceOperandBytes: workload.batchSize * workload.rowBytes,
	}, nil
}

func measureDecoys(table []byte, workload workload, sampleIndex int) (sample, error) {
	targets := targets(workload, sampleIndex)
	candidateSets := decoyCandidates(targets, workload)
	runtime.GC()
	response, serverTotal, err := measureServerRepeated(func() ([]byte, error) {
		result := make([]byte, 0, workload.batchSize*decoyCount*workload.rowBytes)
		for _, candidates := range candidateSets {
			for _, candidate := range candidates {
				result = append(result, row(table, workload, candidate)...)
			}
		}
		return result, nil
	})
	if err != nil {
		return sample{}, err
	}

	_, clientQuery, err := measureRepeated(func() ([][]int, error) {
		return decoyCandidates(targets, workload), nil
	})
	if err != nil {
		return sample{}, err
	}

	_, clientFinish, err := measureRepeated(func() (struct{}, error) {
		for batchIndex, target := range targets {
			offset := batchIndex * decoyCount * workload.rowBytes
			if !bytes.Equal(response[offset:offset+workload.rowBytes], row(table, workload, target)) {
				return struct{}{}, errors.New("decoy answer mismatch")
			}
		}
		keepAlive(response)
		return struct{}{}, nil
	})
	if err != nil {
		return sample{}, err
	}

	return sample{
		clientQuery:        clientQuery,
		serverTotal:        serverTotal,
		clientFinish:       clientFinish,
		uploadBytes:        workload.batchSize * decoyCount * 8,
		downloadBytes:      workload.batchSize * decoyCount * workload.rowBytes,
		sourceOperandBytes: workload.batchSize * decoyCount * workload.rowBytes,
	}, nil
}

func measureDense(
	table []byte,
	workload workload,
	serverCount int,
	sampleIndex int,
) (sample, error) {
	targets := targets(workload, sampleIndex)
	perQuery, err := makeDenseQueries(targets, workload, serverCount)
	if err != nil {
		return sample{}, err
	}
	queries := make([][][]byte, serverCount)
	for server := range serverCount {
		queries[server] = make([][]byte, len(targets))
		for queryIndex := range targets {
			queries[server][queryIndex] = perQuery[queryIndex][server]
		}
	}
	keepAlive(queries)

	runtime.GC()
	var serverTotal time.Duration
	answers := make([][][]byte, 0, serverCount)
	for _, serverQueries := range queries {
		serverAnswers, elapsed, err := measureServerRepeated(func() ([][]byte, error) {
			return answerBatch(table, workload.rows, workload.rowBytes, serverQueries)
		})
		if err != nil {
			return sample{}, err
		}
		serverTotal += elapsed
		answers = append(answers, serverAnswers)
	}
	keepAlive(answers)

	_, clientQuery, err := measureRepeated(func() ([][][]byte, error) {
		return makeDenseQueries(targets, workload, serverCount)
	})
	if err != nil {
		return sample{}, err
	}

	_, clientFinish, err := measureRepeated(func() (struct{}, error) {
		for batchIndex, target := range targets {
			shares := make([][]byte, serverCount)
			for server := range serverCount {
				shares[server] = answers[server][batchIndex]
			}
			answer, err := combine(shares)
			if err != nil {
				return struct{}{}, err
			}
			if !bytes.Equal(answer, row(table, workload, target)) {
				return struct{}{}, errors.New("Dense XOR answer mismatch")
			}
			keepAlive(answer)
		}
		return struct{}{}, nil
	})
	if err != nil {
		return sample{}, err
	}

	return sample{
		clientQuery:        clientQuery,
		serverTotal:        serverTotal,
		clientFinish:       clientFinish,
		uploadBytes:        serverCount * workload.batchSize * querySize(workload.rows),
		downloadBytes:      serverCount * workload.batchSize * workload.rowBytes,
		sourceOperandBytes: serverCount * workload.batchSize * ((workload.rows + 1) / 2) * workload.rowBytes,
	}, nil
}

func makeDenseQueries(targets []int, workload workload, serverCount int) ([][][]byte, error) {
	result := make([][][]byte, len(targets))
	for index, target := range targets {
		shares, err := queryShares(target, workload.rows, serverCount, rand.Reader)
		if err != nil {
			return nil, err
		}
		result[index] = shares
	}
	return result, nil
}

func decoyCandidates(targets []int, workload workload) [][]int {
	result := make([][]int, len(targets))
	for batchIndex, target := range targets {
		result[batchIndex] = make([]int, decoyCount)
		for candidate := range decoyCount {
			if candidate == 0 {
				result[batchIndex][candidate] = target
			} else {
				result[batchIndex][candidate] = (target + candidate*104_729) % workload.rows
			}
		}
	}
	return result
}

func measureRepeated[T any](operation func() (T, error)) (T, time.Duration, error) {
	return measureRepeatedFor(minClientSample, operation)
}

func measureServerRepeated[T any](operation func() (T, error)) (T, time.Duration, error) {
	if _, err := operation(); err != nil {
		var result T
		return result, 0, err
	}
	return measureRepeatedFor(minServerSample, operation)
}

func measureRepeatedFor[T any](
	minimum time.Duration,
	operation func() (T, error),
) (T, time.Duration, error) {
	repeats := 1
	for {
		started := time.Now()
		var result T
		for range repeats {
			current, err := operation()
			if err != nil {
				return result, 0, err
			}
			result = current
		}
		elapsed := time.Since(started)
		keepAlive(result)
		if elapsed >= minimum || repeats == maxClientRepeats {
			return result, elapsed / time.Duration(repeats), nil
		}
		repeats = min(repeats*2, maxClientRepeats)
	}
}

func benchmarkTable(rowCount int, rowBytes int, seed uint64) ([]byte, error) {
	if rowCount <= 0 || rowCount&(rowCount-1) != 0 || rowBytes <= 10 {
		return nil, errors.New("row count must be a power of two and row size must exceed 10")
	}
	table := make([]byte, rowCount*rowBytes)
	state := seed
	for index := range table {
		state ^= state << 13
		state ^= state >> 7
		state ^= state << 17
		table[index] = byte(state)
	}
	return table, nil
}

func targets(workload workload, sampleIndex int) []int {
	result := make([]int, workload.batchSize)
	for batch := range workload.batchSize {
		ordinal := uint64(sampleIndex*workload.batchSize + batch + 1)
		result[batch] = int((tableSeed + ordinal*targetStep) % uint64(workload.rows))
	}
	return result
}

func row(table []byte, workload workload, index int) []byte {
	start := index * workload.rowBytes
	return table[start : start+workload.rowBytes]
}

func medianMeasurement(samples []sample) measurement {
	middle := len(samples) / 2
	median := func(selectDuration func(sample) time.Duration) float64 {
		values := make([]time.Duration, len(samples))
		for index, sample := range samples {
			values[index] = selectDuration(sample)
		}
		sort.Slice(values, func(left, right int) bool { return values[left] < values[right] })
		return float64(values[middle]) / float64(time.Millisecond)
	}
	first := samples[0]
	return measurement{
		ClientQueryP50MS:   median(func(sample sample) time.Duration { return sample.clientQuery }),
		ServerTotalP50MS:   median(func(sample sample) time.Duration { return sample.serverTotal }),
		ClientFinishP50MS:  median(func(sample sample) time.Duration { return sample.clientFinish }),
		UploadBytes:        first.uploadBytes,
		DownloadBytes:      first.downloadBytes,
		SourceOperandBytes: first.sourceOperandBytes,
	}
}

func fnv1a64(value []byte) uint64 {
	hash := uint64(0xcbf29ce484222325)
	for _, currentByte := range value {
		hash ^= uint64(currentByte)
		hash *= 0x00000100000001b3
	}
	return hash
}

// keepAlive prevents the compiler from proving that a benchmark result is unused.
func keepAlive(value any) {
	runtime.KeepAlive(value)
}
