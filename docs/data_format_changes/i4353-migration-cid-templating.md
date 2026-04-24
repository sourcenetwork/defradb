# Migration test uses templated collection version IDs

`TestCollectionMigrationQueryMigratesAcrossMultipleVersions` was updated to use
`{{.CollectionVersionIDN}}` templates in place of hardcoded CIDs. Under the change
detector, setup actions run on the source branch and non-setup actions run on the
target branch, so `s.CollectionVersions` is populated differently between phases
and the templates don't resolve to the same values the source branch saw.
