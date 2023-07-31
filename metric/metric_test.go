// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package metric

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestMetricSyncHistogram(t *testing.T) {
	meter := NewMeter()
	meter.Register("HistogramOnly")
	workDuration, err := meter.GetSyncHistogram(
		"workDuration",
		"ms",
	)
	if err != nil {
		t.Error(err)
	}

	ctx := context.Background()

	// Note: Bucket bounds = [0 5 10 25 ...]
	elapsedTime := 2 * time.Nanosecond
	// Goes in second bucket.
	workDuration.Record(ctx, elapsedTime.Nanoseconds())

	elapsedTime = 4 * time.Nanosecond
	// Goes in second bucket.
	workDuration.Record(ctx, elapsedTime.Nanoseconds())

	elapsedTime = 6 * time.Nanosecond
	// Goes in third bucket.
	workDuration.Record(ctx, elapsedTime.Nanoseconds())

	data, err := meter.Dump(ctx)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 1, len(data.ScopeMetrics))
	assert.Equal(t, "HistogramOnly", data.ScopeMetrics[0].Scope.Name)
	assert.Equal(t, 1, len(data.ScopeMetrics[0].Metrics))
	assert.Equal(t, "workDuration", data.ScopeMetrics[0].Metrics[0].Name)

	firstMetricData := data.ScopeMetrics[0].Metrics[0].Data
	histData, isHistData := firstMetricData.(metricdata.Histogram[int64])
	if !isHistData {
		t.Error(err)
	}

	assert.Equal(t, 1, len(histData.DataPoints))
	assert.Equal(t, uint64(3), histData.DataPoints[0].Count)
	assert.Equal(t, int64(12), histData.DataPoints[0].Sum) // 2 + 4 + 6
	assert.Equal(
		t,
		[]float64{0, 5, 10, 25, 50, 75, 100, 250, 500, 750, 1000, 2500, 5000, 7500, 10000},
		histData.DataPoints[0].Bounds,
	)
	assert.Equal(
		t,
		[]uint64{0, 2, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		histData.DataPoints[0].BucketCounts,
	)

	if meter.Close(ctx) != nil {
		t.Error(err)
	}
}

func TestMetricSyncCounter(t *testing.T) {
	meter := NewMeter()
	meter.Register("CounterOnly")
	stuffCounter, err := meter.GetSyncCounter(
		"countStuff",
		"1",
	)
	if err != nil {
		t.Error(err)
	}

	ctx := context.Background()
	stuffCounter.Add(ctx, 12)
	stuffCounter.Add(ctx, 1)

	data, err := meter.Dump(ctx)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 1, len(data.ScopeMetrics))
	assert.Equal(t, "CounterOnly", data.ScopeMetrics[0].Scope.Name)
	assert.Equal(t, 1, len(data.ScopeMetrics[0].Metrics))
	assert.Equal(t, "countStuff", data.ScopeMetrics[0].Metrics[0].Name)

	firstMetricData := data.ScopeMetrics[0].Metrics[0].Data
	sumData, isSum := firstMetricData.(metricdata.Sum[int64])
	if !isSum {
		t.Error(err)
	}
	assert.Equal(t, "CumulativeTemporality", sumData.Temporality.String())
	assert.Equal(t, 1, len(sumData.DataPoints))
	assert.Equal(t, int64(13), sumData.DataPoints[0].Value) // 12 + 1

	if meter.Close(ctx) != nil {
		t.Error(err)
	}
}

func TestMetricWithCounterAndHistogramIntrumentOnOneMeter(t *testing.T) {
	meter := NewMeter()

	meter.Register("CounterAndHistogram")

	stuffCounter, err := meter.GetSyncCounter(
		"countStuff",
		"1",
	)
	if err != nil {
		t.Error(err)
	}

	workDuration, err := meter.GetSyncHistogram(
		"workDuration",
		"ms",
	)
	if err != nil {
		t.Error(err)
	}

	ctx := context.Background()

	elapsedTime := 2 * time.Nanosecond
	workDuration.Record(ctx, elapsedTime.Nanoseconds())

	stuffCounter.Add(ctx, 12)

	elapsedTime = 4 * time.Nanosecond
	workDuration.Record(ctx, elapsedTime.Nanoseconds())

	elapsedTime = 6 * time.Nanosecond
	workDuration.Record(ctx, elapsedTime.Nanoseconds())

	stuffCounter.Add(ctx, 1)

	data, err := meter.Dump(ctx)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 1, len(data.ScopeMetrics))
	assert.Equal(t, "CounterAndHistogram", data.ScopeMetrics[0].Scope.Name)

	metrics := data.ScopeMetrics[0].Metrics
	assert.Equal(t, 2, len(metrics))

	// Assert Counter
	assert.Equal(t, "countStuff", metrics[0].Name)
	countMetricData := metrics[0].Data
	sumData, isSum := countMetricData.(metricdata.Sum[int64])
	if !isSum {
		t.Error(err)
	}
	assert.Equal(t, "CumulativeTemporality", sumData.Temporality.String())
	assert.Equal(t, 1, len(sumData.DataPoints))
	assert.Equal(t, int64(13), sumData.DataPoints[0].Value) // 12 + 1

	// Assert Histogram
	assert.Equal(t, "workDuration", metrics[1].Name)

	histMetricData := metrics[1].Data
	histData, isHistData := histMetricData.(metricdata.Histogram[int64])
	if !isHistData {
		t.Error(err)
	}

	assert.Equal(t, 1, len(histData.DataPoints))
	assert.Equal(t, uint64(3), histData.DataPoints[0].Count)
	assert.Equal(t, int64(12), histData.DataPoints[0].Sum) // 2 + 4 + 6
	assert.Equal(
		t,
		[]float64{0, 5, 10, 25, 50, 75, 100, 250, 500, 750, 1000, 2500, 5000, 7500, 10000},
		histData.DataPoints[0].Bounds,
	)
	assert.Equal(
		t,
		[]uint64{0, 2, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		histData.DataPoints[0].BucketCounts,
	)

	if meter.Close(ctx) != nil {
		t.Error(err)
	}
}
