import {
  getProviderDisplayName,
  startOAuthLogin,
  useCurrentUser,
  useLinkedOAuthAccounts,
  useOAuthProviders,
  useUnlinkOAuth,
} from '@/features/auth/hooks';
import { Alert, Button, Group, Skeleton, Stack, Text, Title } from '@mantine/core';
import { notifications } from '@mantine/notifications';
import { IconAlertCircle, IconBrandFacebook, IconBrandGithub, IconBrandGoogle, IconCheck, IconPlus } from '@tabler/icons-react';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';

/**
 * Get the icon component for an OAuth provider
 */
function getProviderIcon(provider: string, size = 20) {
  switch (provider) {
    case 'github':
      return <IconBrandGithub size={size} />;
    case 'google':
      return <IconBrandGoogle size={size} />;
    case 'facebook':
      return <IconBrandFacebook size={size} />;
    default:
      return null;
  }
}

/**
 * OAuthManagementPanel component for managing linked OAuth accounts
 * 
 * This component allows users to:
 * - View linked OAuth accounts (GitHub, Google, Facebook)
 * - Link new OAuth providers
 * - Unlink OAuth providers (if they have another auth method)
 */
export function OAuthManagementPanel() {
  const { t } = useTranslation();
  const userId = useCurrentUser();
  const { data: linkedAccounts, isLoading: isLoadingAccounts } = useLinkedOAuthAccounts(userId ?? undefined);
  const { data: providers, isLoading: isLoadingProviders } = useOAuthProviders();
  const unlinkMutation = useUnlinkOAuth(userId ?? undefined);
  const [unlinkingProvider, setUnlinkingProvider] = useState<string | null>(null);

  // Get linked provider names
  const linkedProviderNames = linkedAccounts?.accounts.map(a => a.provider) ?? [];

  // Get enabled providers that can be linked
  const enabledProviders = providers?.providers.filter(p => p.enabled) ?? [];
  const availableToLink = enabledProviders.filter(p => !linkedProviderNames.includes(p.name));

  // Handle unlink
  const handleUnlink = async (provider: string) => {
    setUnlinkingProvider(provider);
    try {
      await unlinkMutation.mutateAsync(provider);
      notifications.show({
        title: t('auth.oauth.management.unlinkSuccess'),
        message: t('auth.oauth.management.unlinkSuccessMessage', { provider: getProviderDisplayName(provider) }),
        color: 'green',
        icon: <IconCheck size={16} />,
      });
    } catch (error) {
      const message = error instanceof Error ? error.message : t('auth.oauth.management.unlinkError');
      notifications.show({
        title: t('auth.oauth.management.unlinkErrorTitle'),
        message,
        color: 'red',
        icon: <IconAlertCircle size={16} />,
      });
    } finally {
      setUnlinkingProvider(null);
    }
  };

  // Loading state
  if (isLoadingAccounts || isLoadingProviders) {
    return (
      <Stack gap="md">
        <Skeleton height={24} width="50%" />
        <Skeleton height={40} />
        <Skeleton height={40} />
      </Stack>
    );
  }

  // No enabled OAuth providers
  if (enabledProviders.length === 0) {
    return null;
  }

  // The backend will validate and return an error if the user tries to unlink their last auth method
  // We don't disable the unlink button on the frontend since we don't have complete knowledge
  // of all user's auth methods (password, passkey, other OAuth providers)

  return (
    <Stack gap="md">
      <div>
        <Title order={4}>{t('auth.oauth.management.title')}</Title>
        <Text c="dimmed" size="sm">
          {t('auth.oauth.management.description')}
        </Text>
      </div>

      {/* Linked accounts */}
      {linkedAccounts && linkedAccounts.accounts.length > 0 && (
        <Stack gap="xs">
          <Text size="sm" fw={500}>{t('auth.oauth.management.linkedAccounts')}</Text>
          {linkedAccounts.accounts.map((account) => (
            <Group key={account.id} justify="space-between" wrap="nowrap" p="xs" style={{ border: '1px solid var(--mantine-color-default-border)', borderRadius: 'var(--mantine-radius-md)' }}>
              <Group gap="sm">
                {getProviderIcon(account.provider)}
                <div>
                  <Text size="sm" fw={500}>
                    {getProviderDisplayName(account.provider)}
                  </Text>
                  {account.provider_email && (
                    <Text size="xs" c="dimmed">
                      {account.provider_email}
                    </Text>
                  )}
                </div>
              </Group>
              <Button
                variant="subtle"
                color="red"
                size="xs"
                onClick={() => handleUnlink(account.provider)}
                loading={unlinkMutation.isPending && unlinkingProvider === account.provider}
                aria-label={t('auth.oauth.unlink', { provider: getProviderDisplayName(account.provider) })}
              >
                {t('auth.oauth.unlinkButton')}
              </Button>
            </Group>
          ))}
        </Stack>
      )}

      {/* Link new providers */}
      {availableToLink.length > 0 && (
        <Stack gap="xs">
          <Text size="sm" fw={500}>{t('auth.oauth.management.linkNew')}</Text>
          <Group gap="xs">
            {availableToLink.map((provider) => (
              <Button
                key={provider.name}
                variant="outline"
                leftSection={getProviderIcon(provider.name, 16)}
                rightSection={<IconPlus size={14} />}
                onClick={() => startOAuthLogin(provider.name)}
                size="sm"
              >
                {getProviderDisplayName(provider.name)}
              </Button>
            ))}
          </Group>
        </Stack>
      )}

      {/* No linked accounts message */}
      {(!linkedAccounts || linkedAccounts.accounts.length === 0) && enabledProviders.length > 0 && (
        <Alert color="blue" icon={<IconAlertCircle size={16} />}>
          <Text size="sm">{t('auth.oauth.management.noLinkedYet')}</Text>
        </Alert>
      )}
    </Stack>
  );
}
