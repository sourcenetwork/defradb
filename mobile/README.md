# Mobile

This package contains functions for generating mobile bindings.

All methods defined in the `client.TxnStore` interface are made available as independant functions.

The parameters for each function are modified to work with the supported `gomobile` types.

## Building

Setup gomobile:

```bash
$ go install golang.org/x/mobile/cmd/gomobile@latest
$ gomobile init
```

To verify that the Android build is successful:

```bash
$ gomobile bind -o dist/defradb.aar -target=android/arm64,android/amd64 -androidapi=21 .
```

To verify that the iOS build is successful:

```bash
$ gomobile bind -o dist/defradb.xcframework -target=ios .
```
