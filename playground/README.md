# DefraDB Playground

A web based playground for DefraDB.

## Developing

Run a development server bound to `localhost:5173`.

```bash
npm install
npm run dev
```

Start DefraDB with CORS allowed.

```bash
defradb start --allowed-origins="*"
```

## Building

Create a static build and output files to `./dist`.

```bash
npm install
npm run build
```
