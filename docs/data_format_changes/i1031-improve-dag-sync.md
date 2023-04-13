# Change the way the priority is set

The priority of a field is set both in the data store and in the block store. Previously, the data store priority was up by one against the block store. We changed it to be the same which resulted in a breaking change on the priority comparison from one version to the next.