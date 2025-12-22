# Enforce unique index for 1:1 relationships

The current change enforces unique index on 1:1 relationships. In order to migrate to this version unique secondary 
indexes must be created manually for all existing 1:1 relations.
