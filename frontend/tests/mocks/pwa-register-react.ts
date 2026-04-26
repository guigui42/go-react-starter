/**
 * Mock for virtual:pwa-register/react
 *
 * This mock is used during testing since the virtual module
 * provided by vite-plugin-pwa is not available in the test environment.
 */
import { useCallback, useState } from 'react';

export interface RegisterSWOptions {
  immediate?: boolean;
  onNeedRefresh?: () => void;
  onOfflineReady?: () => void;
  onRegistered?: (registration: ServiceWorkerRegistration | undefined) => void;
  onRegisteredSW?: (swUrl: string, registration: ServiceWorkerRegistration | undefined) => void;
  onRegisterError?: (error: Error) => void;
}

export interface UseRegisterSWReturn {
  needRefresh: [boolean, (value: boolean) => void];
  offlineReady: [boolean, (value: boolean) => void];
  updateServiceWorker: (reloadPage?: boolean) => Promise<void>;
}

export function useRegisterSW(_options?: RegisterSWOptions): UseRegisterSWReturn {
  const [needRefresh, setNeedRefresh] = useState(false);
  const [offlineReady, setOfflineReady] = useState(false);

  const updateServiceWorker = useCallback(async (_reloadPage?: boolean) => {
    // Mock implementation - does nothing in tests
  }, []);

  return {
    needRefresh: [needRefresh, setNeedRefresh],
    offlineReady: [offlineReady, setOfflineReady],
    updateServiceWorker,
  };
}
