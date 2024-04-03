## Accept vs Reject:
- All tests are broken into `accept_*_test.go` and `reject_*_test.go` files.
- Accepted tests are with valid DPIs (hence schema is accepted).
- Rejected tests are with invalid DPIs (hence schema is rejected).
- There are also some Partially-DPI tests that are both accepted and rejected depending on the resource.

Learn more about the DefraDB Policy Interface [DPI](/acp/README.md)
