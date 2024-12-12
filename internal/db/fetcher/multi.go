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

// multiFetcher is a fetcher that orchastrates the fetching of documents via multiple child fetchers.
//
// The documents are yielded ordered by docID independently of which child fetcher they are sourced from.
type multiFetcher struct {
	children []fetcherDocID

	// The index of the fetcher that last returned an item from `NextDoc`.
	//
	// Used to identify which fetcher to get the rest of the fields from in `GetFields`.
	currentFetcherIndex int
}

var _ fetcher = (*multiFetcher)(nil)

func newMultiFetcher(
	children ...fetcher,
) *multiFetcher {
	fetcherDocIDs := make([]fetcherDocID, len(children))

	for i, fetcher := range children {
		fetcherDocIDs[i] = fetcherDocID{
			fetcher: fetcher,
		}
	}

	return &multiFetcher{
		children: fetcherDocIDs,
	}
}

// fetcherDocID holds a fetcher and the last docID that it returned.
type fetcherDocID struct {
	fetcher fetcher

	// The last docID that this fetcher returned.
	docID immutable.Option[string]
}

func (f *multiFetcher) NextDoc() (immutable.Option[string], error) {
	selectedFetcherIndex := -1
	var selectedDocID immutable.Option[string]

	for i := 0; i < len(f.children); {
		if !f.children[i].docID.HasValue() {
			docID, err := f.children[i].fetcher.NextDoc()
			if err != nil {
				return immutable.None[string](), err
			}
			f.children[i].docID = docID
		}

		if !f.children[i].docID.HasValue() {
			// If `NextDoc` does not return a document, the fetcher is exhausted and can be
			// closed.  There is no point calling `NextDoc` on empty fetchers until all fetchers
			// are exhausted.
			err := f.children[i].fetcher.Close()
			if err != nil {
				return immutable.None[string](), err
			}

			// Remove the fetcher from the list of children
			f.children = append(f.children[:i], f.children[i+1:]...)
			continue
		}

		if !selectedDocID.HasValue() || (f.children[i].docID.Value() < selectedDocID.Value()) {
			selectedFetcherIndex = i
			selectedDocID = f.children[i].docID
		}

		i++
	}

	f.currentFetcherIndex = selectedFetcherIndex
	return selectedDocID, nil
}

func (f *multiFetcher) GetFields() (immutable.Option[EncodedDocument], error) {
	doc, err := f.children[f.currentFetcherIndex].fetcher.GetFields()
	if err != nil {
		return immutable.None[EncodedDocument](), err
	}

	f.children[f.currentFetcherIndex].docID = immutable.None[string]()

	return doc, nil
}

func (f *multiFetcher) Close() error {
	errs := []error{}
	for _, child := range f.children {
		err := child.fetcher.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		return errors.Join(errs...)
	}

	return nil
}
