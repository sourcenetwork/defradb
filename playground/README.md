# DefraDB Playground

A web based playground for DefraDB.

## Developing

The playground supports two modes, selectable via npm scripts:

#### Wasm mode
Runs DefraDB wasm in the browser.

**Steps:**
- Build the **defradb.wasm** binary:
  ```bash
  GOOS=js GOARCH=wasm go build -o playground/defradb.wasm ./cmd/defradb
  ```
- Then run:
  ```bash
  cd playground
  npm install
  npm run dev:wasm
  ```

#### Remote mode
Connects to a running DefraDB node.

**Steps:**
- Start DefraDB with CORS allowed:
  ```bash
  defradb start --allowed-origins="*"
  ```
- Then run:
  ```bash
  cd playground
  npm install
  npm run dev:remote
  ```

## Building

Create a static build and output files to `./dist`.

```bash
npm install
npm run build
```
