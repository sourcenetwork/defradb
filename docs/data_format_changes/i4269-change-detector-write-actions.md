# Fix change detector to detect breaking changes in write actions

This is not a breaking (production code) change.

This modifies the change detector's automatic action splitting logic to no longer treat CreateDoc and UpdateDoc as setup actions. Previously, these actions were always executed by the source branch and skipped by the target branch, which meant breaking changes in write operations would not be detected.

Now, the split will occur before the first non-SchemaUpdate/Restart action (typically CreateDoc), allowing the target branch to execute write operations and catch any breaking changes.
