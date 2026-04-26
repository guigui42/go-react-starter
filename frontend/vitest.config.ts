import react from '@vitejs/plugin-react';
import path from 'path';
import { defineConfig } from 'vitest/config';

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      '@tests': path.resolve(__dirname, './tests'),
      // Mock virtual PWA module for tests
      'virtual:pwa-register/react': path.resolve(__dirname, './tests/mocks/pwa-register-react.ts'),
    },
    conditions: ['browser', 'development'],
  },
  test: {
    globals: true,
    watch: false,
    environment: 'happy-dom',
    setupFiles: './tests/setup.ts',
    // Set environment to 'development' to enable React.act
    env: {
      NODE_ENV: 'test',
      VITE_API_URL: 'http://localhost:8080',
    },
    // Exclude E2E tests from Vitest - they should run with Playwright
    exclude: [
      '**/node_modules/**',
      '**/dist/**',
      '**/cypress/**',
      '**/.{idea,git,cache,output,temp}/**',
      '**/{karma,rollup,webpack,vite,vitest,jest,ava,babel,nyc,cypress,tsup,build}.config.*',
      '**/e2e/**', // Exclude Playwright E2E tests
    ],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'html', 'lcov'],
      exclude: [
        'node_modules/',
        'tests/',
        '**/*.test.{ts,tsx}',
        '**/*.spec.{ts,tsx}',
        '*.config.ts',
        'src/routeTree.gen.ts',
      ],
      thresholds: {
        statements: 70,
        branches: 70,
        functions: 70,
        lines: 70,
      },
    },
  },
});
