# DefraDB Playground

A web based playground for DefraDB.

## Table of Contents

<!--ts-->
   * [Prerequisites](#prerequisites)
   * [Getting Started](#getting-started)
      * [Remote mode](#remote-mode)
      * [Wasm mode](#wasm-mode)
         * [Local ACP (default)](#local-acp-default)
         * [SourceHub ACP (requires running SourceHub)](#sourcehub-acp-requires-running-sourcehub)
   * [Building](#building)
<!--te-->

## Prerequisites
- Go `v1.23` or later
- Node.js `^20.19.0 || >=22.12.0`

## Getting Started

The playground supports two modes, selectable via npm scripts.

### Remote mode
Connects to a running DefraDB node.

**Steps:**
- Start DefraDB with CORS allowed:
  ```bash
  DEFRA_KEYRING_SECRET=your_secret defradb start --allowed-origins="*"
  ```
- Then run:
  ```bash
  cd playground
  npm install
  npm run dev:remote
  ```

### Wasm mode
Runs DefraDB wasm in the browser.

**Steps:**
- Build the **defradb.wasm** binary:
  ```bash
  GOOS=js GOARCH=wasm go build -o playground/defradb.wasm ./cmd/defradb
  ```
- Then run one of the following commands:

#### Local ACP (default)
```bash
cd playground
npm install
npm run dev:wasm
```

#### SourceHub ACP (requires running SourceHub)
```bash
cd playground
cp .env.example .env
npm install
npm run dev:wasm:sourcehub
```

The `npm run dev:wasm:sourcehub` command automatically attempts to connect to SourceHub. If SourceHub is not running or unreachable, the playground will automatically fall back to Local ACP mode.

**Note:** When SourceHub is available, the playground includes a keypair reset tab to manually generate a new keypair if the SourceHub state was reset.

**Running SourceHub with API endpoint at `http://localhost:1317`**
```bash
git clone https://github.com/sourcenetwork/sourcehub.git
cd sourcehub && git checkout dev && git pull
make build # make build-mac
./scripts/dev-entrypoint.sh start
```

**CORS Configuration**
The playground uses Vite's proxy to route:
- `/api/*` → `http://localhost:1317/*` (SourceHub API)
- `/rpc/*` → `http://localhost:26657/*` (SourceHub RPC)

**Examples**
When running the playground, the DefraDB client is available in the developer console via the `window.defradbClient` object.

```js
// Add a schema
const schema = `
type User {
  name: String
  age: Int
}
`;

window.defradbClient.addSchema(schema);
```

```js
// Add a schema with a policy
const schema = `
type Users @policy(
  id: "policy_id",
  resource: "users"
) {
  name: String
  age: Int
}
`;

window.defradbClient.addSchema(schema);
```

```js
// Add a document
const query = `
mutation {
  create_Users(input: {
    name: "John Doe 1",
    age: 33
  }) {
    _docID
    name
    age
  }
}
`;

window.defradbClient.execRequest(query, {}, {});
```

```js
// Add a policy with ACP configuration
const policy = `
description: A Valid DefraDB Policy Interface

actor:
  name: actor

resources:
  users:
    permissions:
      read:
        expr: owner + reader
      update:
        expr: owner
      delete:
        expr: owner

    relations:
      owner:
        types:
          - actor
      reader:
        types:
          - actor
`;

const context = {
  identity: "pub_key"
};

window.defradbClient.addDACPolicy(policy, context);
```

## Building

Create a static build and output files to `./dist`.

```bash
npm install
npm run build
```