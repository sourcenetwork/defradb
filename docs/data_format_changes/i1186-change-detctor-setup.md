# Modify change detector setup flow

This is not a breaking (production code) change.

This does however change the way the change detector sets up the target branch in a way that is incompatible with the current develop, this file is required to allow the CI to pass (including comparing develop vs master).
