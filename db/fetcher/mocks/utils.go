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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"

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
	).Maybe().Return(nil)
	f.EXPECT().Start(mock.Anything, mock.Anything, mock.Anything).Maybe().Return(nil)
	f.EXPECT().FetchNext(mock.Anything).Maybe().Return(nil, nil)
	f.EXPECT().FetchNextDoc(mock.Anything, mock.Anything).Maybe().
		Return(NewEncodedDocument(t), core.Doc{}, nil)
	f.EXPECT().FetchNextDecoded(mock.Anything).Maybe().Return(&client.Document{}, nil)
	f.EXPECT().Close().Maybe().Return(nil)
	return f
}
