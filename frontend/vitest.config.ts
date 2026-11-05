import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  test: {
    environment: 'jsdom',
    globals: true,
    // D-31 بیرون از src است تا tsc برنامه به node:fs نیاز پیدا نکند.
    include: ['src/**/*.test.{ts,tsx}', 'd31-production-deps.test.ts'],
    setupFiles: ['./src/test-setup.ts'],
  },
});
