import { defineConfig, loadEnv } from 'vite';
import react from '@vitejs/plugin-react-swc';

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');

  return {
    plugins: [react()],
    server: {
      proxy: {
        '/api': {
          target: env.VITE_API_TARGET || 'http://localhost:1317',
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/api/, ''),
          configure: (proxy) => {
            proxy.on('proxyReq', (proxyReq, req, res) => {
              res.setHeader('Access-Control-Allow-Origin', '*');
              res.setHeader('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, OPTIONS');
              res.setHeader('Access-Control-Allow-Headers',
                'Content-Type, Authorization, X-Requested-With, Accept, Origin, ' +
                'x-grpc-web, x-grpc-timeout, x-user-agent, x-grpc-encoding, x-grpc-accept-encoding',
              );
              res.setHeader('Access-Control-Allow-Credentials', 'true');
              if (req.method === 'OPTIONS') {
                res.writeHead(200);
                res.end();
                return;
              }
              const grpcHeaders = [
                'x-grpc-web',
                'x-grpc-timeout',
                'x-user-agent',
                'x-grpc-encoding',
                'x-grpc-accept-encoding',
              ];
              grpcHeaders.forEach(header => {
                if (req.headers[header]) {
                  proxyReq.setHeader(header, req.headers[header]);
                }
              });
              if (req.headers['content-type'] === 'application/grpc-web+proto') {
                proxyReq.setHeader('Content-Type', 'application/grpc-web+proto');
              }
            });

            proxy.on('proxyRes', (proxyRes, _req, res) => {
              proxyRes.headers['Access-Control-Allow-Origin'] = '*';
              proxyRes.headers['Access-Control-Allow-Methods'] = 'GET, POST, PUT, DELETE, OPTIONS';
              proxyRes.headers['Access-Control-Allow-Headers'] =
                'Content-Type, Authorization, X-Requested-With, Accept, Origin, ' +
                'x-grpc-web, x-grpc-timeout, x-user-agent, x-grpc-encoding, x-grpc-accept-encoding';
              proxyRes.headers['Access-Control-Allow-Credentials'] = 'true';
              if (proxyRes.headers['x-grpc-status']) {
                res.setHeader('X-Grpc-Status', proxyRes.headers['x-grpc-status']);
              }
              if (proxyRes.headers['x-grpc-message']) {
                res.setHeader('X-Grpc-Message', proxyRes.headers['x-grpc-message']);
              }
            });

            proxy.on('error', (err, _req, res) => {
              console.error('Proxy error:', err);
              if (!res.headersSent) {
                res.writeHead(500, { 'Content-Type': 'text/plain' });
                res.end('Proxy error: ' + err.message);
              }
            });
          },
        },

        '/rpc': {
          target: env.VITE_RPC_TARGET || 'http://localhost:26657',
          changeOrigin: true,
          configure: (proxy) => {
            proxy.on('proxyReq', (_proxyReq, req, res) => {
              res.setHeader('Access-Control-Allow-Origin', '*');
              res.setHeader('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, OPTIONS');
              res.setHeader('Access-Control-Allow-Headers', 'Content-Type, Authorization, X-Requested-With, Accept, Origin');
              res.setHeader('Access-Control-Allow-Credentials', 'true');
              if (req.method === 'OPTIONS') {
                res.writeHead(200);
                res.end();
                return;
              }
            });

            proxy.on('proxyRes', (proxyRes) => {
              proxyRes.headers['Access-Control-Allow-Origin'] = '*';
              proxyRes.headers['Access-Control-Allow-Methods'] = 'GET, POST, PUT, DELETE, OPTIONS';
              proxyRes.headers['Access-Control-Allow-Headers'] = 'Content-Type, Authorization, X-Requested-With, Accept, Origin';
              proxyRes.headers['Access-Control-Allow-Credentials'] = 'true';
            });

            proxy.on('error', (err, _req, res) => {
              console.error('RPC Proxy error:', err);
              if (!res.headersSent) {
                res.writeHead(500, { 'Content-Type': 'text/plain' });
                res.end('RPC Proxy error: ' + err.message);
              }
            });
          },
        },
      },
    },
  };
});
