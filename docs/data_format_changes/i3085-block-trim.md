# Remove duplication of block heads on delete

The structure of blocks in the blockstore was reworked slightly - head links have been extracted to a separate property, and fieldName has been removed from composite blocks. 
