## Accept vs Reject:
- All tests are broken into `accept_*_test.go` and `reject_*_test.go` files.
- Accepted tests are with valid DRIs (hence schema is accepted).
- Rejected tests are with invalid DRIs (hence schema is rejected).
- There are also some Partially-DRI tests that are both accepted and rejected depending on the resource.

Learn more about the DefraDB [ACP System](/acp/README.md)
