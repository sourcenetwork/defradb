# Fix support for nillable arrays of nillable types in docs

A bug was fixed, where different document strings containing fields of nillable arrays of nillable types would hash to the same doc ID.
This affected tests that used such arrays, and which relied on hardcoded doc IDs, and the new correct docID had to be used there.
The fix also affects the default order of the results, if no order is applied on the query.