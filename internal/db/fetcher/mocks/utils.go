// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package mocks

import (
	"testing"

	"github.com/stretchr/testify/mock"
)

func NewStubbedFetcher(t *testing.T) *Fetcher {
	f := NewFetcher(t)
	f.EXPECT().Init(
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Maybe().Return(nil)
	f.EXPECT().Start(mock.Anything, mock.Anything).Maybe().Return(nil)
	f.EXPECT().FetchNext(mock.Anything).Maybe().Return(nil, nil)
	f.EXPECT().Close().Maybe().Return(nil)
	return f
}
