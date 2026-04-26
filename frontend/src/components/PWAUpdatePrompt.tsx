/**
 * PWA Update Prompt Component
 *
 * Detects when a new version of the app is available and prompts
 * the user to reload to get the latest version.
 *
 * This solves the cache invalidation problem when deploying new versions.
 * 
 * Best practices applied:
 * - Service worker checks every 5 minutes
 * - Checks on tab visibility change
 * - Compares version number to detect updates
 * - Forces cache clear on update
 */
import { ActionIcon, Button, Group, Paper, Text } from '@mantine/core';
import { IconRefresh, IconX } from '@tabler/icons-react';
import { useCallback, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useRegisterSW } from 'virtual:pwa-register/react';

// Get current app version from build-time env vars
const APP_VERSION = import.meta.env.VITE_APP_VERSION || 'dev';

// Check if there's a version mismatch (computed once at load time)
function getInitialVersionMismatch(): boolean {
  if (typeof window === 'undefined') return false;
  const storedVersion = localStorage.getItem('app_version');
  return !!(storedVersion && storedVersion !== APP_VERSION && APP_VERSION !== 'dev');
}

export function PWAUpdatePrompt() {
  const { t } = useTranslation();
  const [dismissed, setDismissed] = useState(false);
  // Use lazy initializer to compute version mismatch at render time (not in effect)
  const [versionMismatch, setVersionMismatch] = useState(getInitialVersionMismatch);

  const {
    needRefresh: [needRefresh, setNeedRefresh],
  } = useRegisterSW({
    onRegistered(registration: ServiceWorkerRegistration | undefined) {
      // Check for updates every 5 minutes
      if (registration) {
        setInterval(() => {
          registration.update();
        }, 5 * 60 * 1000);
      }
    },
    onRegisterError() {
      // Service worker registration failed - app continues without PWA features
    },
  });

  // Check version on mount and when tab becomes visible
  const checkVersion = useCallback(async () => {
    try {
      // Fetch version from server (bypass cache)
      const response = await fetch('/api/v1/monitoring/health', {
        cache: 'no-store',
        headers: {
          'Cache-Control': 'no-cache',
        },
      });
      
      if (response.ok) {
        // If we got here, the server is reachable
        // The version check will happen via service worker
        navigator.serviceWorker?.getRegistration().then((registration) => {
          registration?.update();
        });
      }
    } catch {
      // Server unreachable, don't show update prompt
    }
  }, []);

  // Check version when tab becomes visible
  useEffect(() => {
    const handleVisibilityChange = () => {
      if (document.visibilityState === 'visible') {
        checkVersion();
      }
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);
    
    // Also check on mount
    checkVersion();
    
    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange);
    };
  }, [checkVersion]);

  // Store current version in localStorage for cross-tab detection
  useEffect(() => {
    localStorage.setItem('app_version', APP_VERSION);
  }, []);

  const handleUpdate = async () => {
    // Tell the waiting service worker to activate via skipWaiting message.
    // With registerType: 'prompt', the new SW waits until we explicitly ask it to take over.
    if ('serviceWorker' in navigator) {
      const registration = await navigator.serviceWorker.getRegistration();
      if (registration?.waiting) {
        // Tell waiting SW to call self.skipWaiting() and take over
        registration.waiting.postMessage({ type: 'SKIP_WAITING' });
      }
    }
    
    // Clear all caches to ensure fresh assets
    if ('caches' in window) {
      const cacheNames = await caches.keys();
      await Promise.all(
        cacheNames.map((cacheName) => caches.delete(cacheName))
      );
    }
    
    // Update localStorage version
    localStorage.setItem('app_version', APP_VERSION);
    
    // Reload after the new SW has taken control
    window.location.reload();
  };

  const handleDismiss = () => {
    setDismissed(true);
    setNeedRefresh(false);
    setVersionMismatch(false);
  };

  // Show prompt if SW detected update OR if version mismatch detected
  if ((!needRefresh && !versionMismatch) || dismissed) {
    return null;
  }

  return (
    <Paper
      shadow="md"
      p="md"
      radius="md"
      withBorder
      style={{
        position: 'fixed',
        bottom: 20,
        right: 20,
        zIndex: 9999,
        maxWidth: 350,
      }}
    >
      <Group justify="space-between" mb="xs">
        <Text fw={600}>{t('pwa.updateAvailable', 'Update Available')}</Text>
        <ActionIcon
          variant="subtle"
          size="sm"
          onClick={handleDismiss}
          aria-label={t('common.dismiss', 'Dismiss')}
        >
          <IconX size={14} />
        </ActionIcon>
      </Group>
      <Text size="sm" c="dimmed" mb="md">
        {t(
          'pwa.updateMessage',
          'A new version of Go React Starter is available. Reload to get the latest features and fixes.'
        )}
      </Text>
      <Group justify="flex-end">
        <Button
          variant="default"
          size="xs"
          onClick={handleDismiss}
        >
          {t('common.later', 'Later')}
        </Button>
        <Button
          leftSection={<IconRefresh size={16} />}
          size="xs"
          onClick={handleUpdate}
        >
          {t('pwa.reload', 'Reload Now')}
        </Button>
      </Group>
    </Paper>
  );
}
