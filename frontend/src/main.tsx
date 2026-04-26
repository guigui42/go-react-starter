import { MantineProvider } from '@mantine/core';
import '@mantine/core/styles.css';
import '@mantine/dates/styles.css';
import { Notifications } from '@mantine/notifications';
import '@mantine/notifications/styles.css';
import { QueryClientProvider } from '@tanstack/react-query';
import { lazy, StrictMode, Suspense } from 'react';
import { createRoot } from 'react-dom/client';
import { queryClient } from './lib/queryClient';
import { theme, cssVariablesResolver } from './theme';

// Lazy load devtools - only in development
const ReactQueryDevtools = import.meta.env.DEV
  ? lazy(() =>
      import('@tanstack/react-query-devtools').then((mod) => ({
        default: mod.ReactQueryDevtools,
      }))
    )
  : () => null;

// Initialize security migration at startup
import { initSecurityMigration } from './lib/securityMigration';
initSecurityMigration();

// Initialize Web Vitals reporting to Umami (async, non-blocking)
import { initWebVitals } from './lib/webVitals';
initWebVitals();

import App from './App.tsx';
import { PWAUpdatePrompt } from './components/PWAUpdatePrompt';
import { AuthProvider } from './contexts/AuthContext';
import { UserPreferencesProvider } from './contexts/UserPreferencesContext';
import './i18n/config'; // Initialize i18n
import './index.css';
createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <MantineProvider theme={theme} defaultColorScheme="auto" cssVariablesResolver={cssVariablesResolver}>
        <Notifications />
        <PWAUpdatePrompt />
        <AuthProvider>
          <UserPreferencesProvider>
            <App />
          </UserPreferencesProvider>
        </AuthProvider>
      </MantineProvider>
      <Suspense fallback={null}>
        <ReactQueryDevtools initialIsOpen={false} />
      </Suspense>
    </QueryClientProvider>

  </StrictMode>
);
