# Move relation field properties onto collection

Field RelationName and secondary relation fields has been made local, and moved off of the schema and onto collection. Field IsPrimary has been removed completely (from the schema).  As a result schema root and schema version id are no longer dependent on them.
