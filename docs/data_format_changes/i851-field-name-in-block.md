# Add support for querying fieldName and fieldID in commits queries

FieldName was added to the blocks, and this changes all the block CIDs and means that anything committed before this change will cause an error to be returned when using the commits queries.
