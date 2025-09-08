# Playground

The playground source code can be found at the repo below.

https://github.com/sourcenetwork/defradb-playground

## Setup

Download the latest playground static assets.

```bash
go generate .
```

Or from the repo root.

```bash
make deps:playground
```

To enable the playground include the `-tags plaground` flag when running or building from source. Then open your browser and navigate to http://localhost:9181.
