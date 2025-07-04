// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App.tsx';
import './index.css';
import './wasm_exec.js';

const mode = import.meta.env.VITE_PLAYGROUND_MODE;

if (mode === 'wasm') {
  // @ts-expect-error - Go is a global object from wasm_exec.js
  const go = new Go();

  WebAssembly.instantiateStreaming(fetch("defradb.wasm"), go.importObject).then(
    (result) => {
      console.log("defradb.wasm loaded.");
      go.run(result.instance);
    },
  ).catch(err => {
    console.error("Error loading Wasm module:", err);
  })
}

ReactDOM.createRoot(document.getElementById('root') as HTMLElement).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
