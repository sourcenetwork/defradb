package mocks

import (
	"testing"

	client "github.com/sourcenetwork/defradb/client"
	core "github.com/sourcenetwork/defradb/core"
	mock "github.com/stretchr/testify/mock"
)

func NewStubbedFetcher(t *testing.T) *Fetcher {
	f := NewFetcher(t)
	f.EXPECT().Init(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(nil)
	f.EXPECT().Start(mock.Anything, mock.Anything, mock.Anything).Maybe().Return(nil)
	f.EXPECT().FetchNext(mock.Anything).Maybe().Return(nil, nil)
	f.EXPECT().FetchNextDoc(mock.Anything, mock.Anything).Maybe().
		Return(NewEncodedDocument(t), core.Doc{}, nil)
	f.EXPECT().FetchNextDecoded(mock.Anything).Maybe().Return(&client.Document{}, nil)
	f.EXPECT().Close().Maybe().Return(nil)
	return f
}
