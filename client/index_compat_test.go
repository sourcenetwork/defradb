package client

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// An in-process caller that predates the kind carrying fields sets only the deprecated top-level
// Fields. Nothing normalizes a struct built in memory, so the read has to fall back to it.
func TestIndexDescription_CompatFieldsOnly_ReadsThrough(t *testing.T) {
	d := IndexDescription{
		Name:            "x",
		ID:              1,
		Fields:          []IndexedFieldDescription{{Name: "age"}},
		Kind:            IndexKindOrdered,
		KindDescription: &OrderedIndexDescription{Unique: true},
	}
	require.Equal(t, []string{"age"}, d.fieldNames())
}
