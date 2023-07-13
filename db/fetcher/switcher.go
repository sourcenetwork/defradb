package fetcher

import (
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/planner/mapper"
)

type FetcherSwitcher struct {
	Fetcher
}

var _ Fetcher = (*FetcherSwitcher)(nil)

func (f *FetcherSwitcher) Init(
	col *client.CollectionDescription,
	fields []client.FieldDescription,
	filter *mapper.Filter,
	docMapper *core.DocumentMapping,
	reverse bool,
	showDeleted bool,
) error {
	var index client.IndexDescription
	var filterCond any
	var indexedFieldDesc client.FieldDescription
	colIndexes := col.Indexes
	for filterFieldName, cond := range filter.ExternalConditions {
		for i := range colIndexes {
			if filterFieldName == colIndexes[i].Fields[0].Name {
				index = colIndexes[i]
				filterCond = cond

				indexedFields := col.CollectIndexedFields()
				for j := range indexedFields {
					if indexedFields[j].Name == filterFieldName {
						indexedFieldDesc = indexedFields[j]
						break
					}
				}
			}
		}
	}

	if index.ID != 0 {
		f.Fetcher = NewIndexFetcher(new(DocumentFetcher), indexedFieldDesc, index, filterCond)
	} else {
		f.Fetcher = new(DocumentFetcher)
	}

	return f.Fetcher.Init(col, fields, filter, docMapper, reverse, showDeleted)
}
