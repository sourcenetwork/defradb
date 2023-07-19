# Enable Refactored Explain Tests That Were Always Skipped

Previously we had explain tests always being skipped, the integration of explain setup into the action based testing
setup enabled them, but since they were being skipped previously change detector keeps failing. This isn't a breaking
change.
