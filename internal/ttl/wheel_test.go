// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package ttl

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewWheelRejectsInvalidConfig(t *testing.T) {
	wheel, err := NewWheel(context.Background(), 0, 1, func(string) {})
	require.ErrorIs(t, err, ErrInvalidTick)
	require.Nil(t, wheel)

	wheel, err = NewWheel(context.Background(), time.Millisecond, 0, func(string) {})
	require.ErrorIs(t, err, ErrInvalidSlotCount)
	require.Nil(t, wheel)
}

func TestWheelRejectsInvalidTTL(t *testing.T) {
	wheel, err := NewWheel(context.Background(), time.Millisecond, 10, func(string) {})
	require.NoError(t, err)

	require.ErrorIs(t, wheel.Add("negative", -time.Millisecond), ErrNegativeTTL)
	require.ErrorIs(t, wheel.Add("beyond", 11*time.Millisecond), ErrBeyondMaxTTL)
}

func TestWheelExpiresExactlyMaxTTLAfterFullWindow(t *testing.T) {
	expired := make(chan string, 1)
	wheel, err := NewWheel(context.Background(), time.Millisecond, 3, func(key string) {
		expired <- key
	})
	require.NoError(t, err)

	require.NoError(t, wheel.Add("key", 3*time.Millisecond))

	wheel.tickOnce()
	wheel.tickOnce()

	select {
	case key := <-expired:
		t.Fatalf("key expired too early: %s", key)
	default:
	}

	wheel.tickOnce()

	select {
	case key := <-expired:
		require.Equal(t, "key", key)
	default:
		t.Fatal("expected key to expire")
	}
}

func TestWheelUpdateTTL(t *testing.T) {
	expired := make(chan string, 1)
	wheel, err := NewWheel(context.Background(), time.Millisecond, 10, func(key string) {
		expired <- key
	})
	require.NoError(t, err)

	require.NoError(t, wheel.Add("key", time.Millisecond))
	require.NoError(t, wheel.UpdateTTL("key", 3*time.Millisecond))

	wheel.tickOnce()
	wheel.tickOnce()

	select {
	case key := <-expired:
		t.Fatalf("key expired too early: %s", key)
	default:
	}

	wheel.tickOnce()

	select {
	case key := <-expired:
		require.Equal(t, "key", key)
	default:
		t.Fatal("expected key to expire")
	}
}

func TestWheelDelete(t *testing.T) {
	expired := make(chan string, 1)
	wheel, err := NewWheel(context.Background(), time.Millisecond, 10, func(key string) {
		expired <- key
	})
	require.NoError(t, err)

	require.NoError(t, wheel.Add("key", time.Millisecond))
	wheel.Delete("key")
	wheel.tickOnce()

	select {
	case key := <-expired:
		t.Fatalf("deleted key expired: %s", key)
	default:
	}
}
