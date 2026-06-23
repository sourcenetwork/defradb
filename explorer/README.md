# Explorer

The explorer source code can be found at the repo below.

https://github.com/sourcenetwork/defradb-explorer

## Setup

Download the latest explorer static assets.

```bash
go generate .
```

Or from the repo root.

```bash
make deps:explorer
```

To enable the explorer include the `-tags explorer` flag when running or building from source. Then open your browser and navigate to http://localhost:9181.
