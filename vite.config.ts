import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'
import { VitePWA } from 'vite-plugin-pwa'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react(),
    tailwindcss(),
    // PWA: manifest para instalar la app y service worker para abrirla sin conexión.
    VitePWA({
      // El service worker nuevo se activa solo cuando hay un despliegue.
      registerType: 'autoUpdate',
      includeAssets: ['favicon.ico', 'apple-touch-icon-180x180.png'],
      manifest: {
        name: 'MultiCatálogo',
        short_name: 'MultiCatálogo',
        description: 'Tienda multicatálogo con red de referidos',
        lang: 'es',
        start_url: '/',
        scope: '/',
        display: 'standalone',
        theme_color: '#14b8a6',
        background_color: '#ffffff',
        icons: [
          { src: 'pwa-64x64.png', sizes: '64x64', type: 'image/png' },
          { src: 'pwa-192x192.png', sizes: '192x192', type: 'image/png' },
          { src: 'pwa-512x512.png', sizes: '512x512', type: 'image/png' },
          { src: 'maskable-icon-512x512.png', sizes: '512x512', type: 'image/png', purpose: 'maskable' },
        ],
      },
      workbox: {
        // La interfaz (HTML, JS, CSS, íconos) queda guardada para abrir sin conexión.
        globPatterns: ['**/*.{js,css,html,ico,png,svg}'],
        // Cualquier ruta de React Router abre index.html, salvo la API.
        navigateFallback: '/index.html',
        navigateFallbackDenylist: [/^\/api\//],
        runtimeCaching: [
          {
            // Fotos de productos: se guardan para verlas sin conexión.
            urlPattern: ({ url }) => url.origin === 'https://picsum.photos' || url.hostname.endsWith('.picsum.photos'),
            handler: 'CacheFirst',
            options: {
              cacheName: 'imagenes-productos',
              expiration: { maxEntries: 100, maxAgeSeconds: 60 * 60 * 24 * 30 },
              cacheableResponse: { statuses: [0, 200] },
            },
          },
          // La API no se cachea: precios, stock y pedidos siempre vienen del servidor.
        ],
      },
    }),
  ],
})
