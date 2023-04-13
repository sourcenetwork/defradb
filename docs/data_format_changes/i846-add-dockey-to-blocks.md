# Store dockey field to delta (block) storage

To be able to request dockey field on commits, it had to be stored first.
Composite blocks didn't have dockey field, so it was added to the block struct.
Field blocks had dockey field, but it didn't store the key with it's instance type.
That's why all CIDs of commits needed to be regenerated.
