import { defineConfig } from 'vite';
import { octane } from '@octanejs/vite-plugin';
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
  plugins: [octane(), tailwindcss()],
  server: { port: 5173, proxy: { '/api': 'http://127.0.0.1:4173' } },
  build: { outDir: 'dist', emptyOutDir: false },
});
