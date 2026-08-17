# Full-text and trigram secondary-index layouts

This change adds two persisted secondary-index kinds to the typed `IndexDescription` union:
`FullText` (enum value 2) and `Trigram` (enum value 3). The values are appended after the existing
ordered and vector kinds; ordered remains the zero value so legacy descriptions still decode as
ordered indexes.

A trigram index uses the canonical secondary-index key layout and stores each distinct lowercased
three-byte window as an indexed value. A BM25 full-text index adds three key families beneath the
canonical collection/index/epoch prefix:

```text
/<collection>/<index>/<epoch>/t/<encoded-term>/<doc-short-id>  term frequency
/<collection>/<index>/<epoch>/d/<doc-short-id>                 field length
/<collection>/<index>/<epoch>/s                                corpus totals
```

This is additive for databases that contain only previously supported index kinds. Databases made
by the earlier experimental BM25/Trigram branch used different kind values/descriptions and should
not be opened directly with this implementation. Drop and recreate those experimental indexes (or
restore the collection without their index descriptions) so DefraDB backfills the canonical
layouts. There is no in-place migration for experimental index entries; they are derived data and
can be rebuilt from documents.
