# Field Short ID Retrieval Bug Fix

Fixed an issue where retrieving field short IDs could result in a corrupted datastore state. The bug occured when more than 9 collections were present as a result of the collection short id being written in the key as a string. Encoding the collection short id as a variable length integer resolves the issue. This however as is changes the way field short ids are stored in the system store.
