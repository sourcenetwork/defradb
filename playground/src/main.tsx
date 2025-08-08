// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

import ReactDOM from 'react-dom/client';
import App from './App.tsx';
import './index.css';
import { instantiate } from '@sourcenetwork/acp-js';

const mode = import.meta.env.VITE_PLAYGROUND_MODE;

if (mode === 'wasm') {
  (window as any).globalACPConfig = {
    apiUrl: `${window.location.origin}${import.meta.env.VITE_ACP_API_URL ?? '/api'}`,
    rpcUrl: `${window.location.origin}${import.meta.env.VITE_ACP_RPC_URL ?? '/rpc'}`,
    grpcUrl: `${window.location.origin}${import.meta.env.VITE_ACP_GRPC_URL ?? '/api'}`,
    chainId: import.meta.env.VITE_ACP_CHAIN_ID ?? 'sourcehub-dev',
    denom: import.meta.env.VITE_ACP_DENOM ?? 'uopen',
    useZeroFees: import.meta.env.VITE_ACP_ALLOW_ZERO_FEES === 'true' || false,
  };

  await instantiate('defradb.wasm');
}

ReactDOM.createRoot(document.getElementById('root') as HTMLElement).render(
  <App />,
);
