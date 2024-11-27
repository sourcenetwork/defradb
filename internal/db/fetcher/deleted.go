// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package fetcher

import (
	"errors"

	"github.com/sourcenetwork/immutable"
)

// deleted is a fetcher that orchastrates the fetching of deleted and active documents.
type deleted struct {
	activeFetcher fetcher
	activeDocID   immutable.Option[string]

	deletedFetcher fetcher
	deletedDocID   immutable.Option[string]

	currentFetcher fetcher
}

var _ fetcher = (*deleted)(nil)

func newDeletedFetcher(
	activeFetcher fetcher,
	deletedFetcher fetcher,
) *deleted {
	return &deleted{
		activeFetcher:  activeFetcher,
		deletedFetcher: deletedFetcher,
	}
}

func (f *deleted) NextDoc() (immutable.Option[string], error) {
	if !f.activeDocID.HasValue() {
		var err error
		f.activeDocID, err = f.activeFetcher.NextDoc()
		if err != nil {
			return immutable.None[string](), err
		}
	}

	if !f.deletedDocID.HasValue() {
		var err error
		f.deletedDocID, err = f.deletedFetcher.NextDoc()
		if err != nil {
			return immutable.None[string](), err
		}
	}

	if !f.activeDocID.HasValue() || (f.deletedDocID.HasValue() && f.deletedDocID.Value() < f.activeDocID.Value()) {
		f.currentFetcher = f.deletedFetcher
		return f.deletedDocID, nil
	}

	f.currentFetcher = f.activeFetcher
	return f.activeDocID, nil
}

func (f *deleted) GetFields() (immutable.Option[EncodedDocument], error) {
	doc, err := f.currentFetcher.GetFields()
	if err != nil {
		return immutable.None[EncodedDocument](), err
	}

	if f.activeFetcher == f.currentFetcher {
		f.activeDocID = immutable.None[string]()
	} else {
		f.deletedDocID = immutable.None[string]()
	}

	return doc, nil
}

func (f *deleted) Close() error {
	activeErr := f.activeFetcher.Close()
	if activeErr != nil {
		deletedErr := f.deletedFetcher.Close()
		if deletedErr != nil {
			return errors.Join(activeErr, deletedErr)
		}

		return activeErr
	}

	return f.deletedFetcher.Close()
}
