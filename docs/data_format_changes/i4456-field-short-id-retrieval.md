# Field Short ID Retrieval Bug Fix

Fixed an issue where retrieving field short IDs could result in a corrupted datastore state. The bug occurred when more than 9 collections were present because the collection short ID was written in the key as a string. Encoding the collection short ID as a variable-length integer resolves the issue. This, however, changes how field short IDs are stored in the system store..
