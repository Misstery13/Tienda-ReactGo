import { defineConfig, minimal2023Preset } from '@vite-pwa/assets-generator/config'

// Genera los íconos de la PWA a partir de public/pwa-icon.svg.
// Regenerar con: npx pwa-assets-generator
export default defineConfig({
  headLinkOptions: { preset: '2023' },
  preset: {
    ...minimal2023Preset,
    maskable: { ...minimal2023Preset.maskable, resizeOptions: { background: '#14b8a6' } },
    apple: { ...minimal2023Preset.apple, resizeOptions: { background: '#14b8a6' } },
  },
  images: ['public/pwa-icon.svg'],
})
