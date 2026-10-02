import { defineConfig } from 'vite';

// Ports are distinct from the defaults (5173/5000/8080) so this app can run
// alongside other local projects. Override with FRONTEND_PORT / BACKEND_URL.
const frontendPort = Number(process.env.FRONTEND_PORT ?? 18471);
const backendUrl = process.env.BACKEND_URL ?? 'http://127.0.0.1:18472';

export default defineConfig({
    server: {
        host: '127.0.0.1',
        port: frontendPort,
        strictPort: true,
        proxy: {
            '/api': { target: backendUrl, changeOrigin: true }
        }
    },
    preview: {
        host: '127.0.0.1',
        port: frontendPort,
        strictPort: true,
        proxy: {
            '/api': { target: backendUrl, changeOrigin: true }
        }
    }
});
