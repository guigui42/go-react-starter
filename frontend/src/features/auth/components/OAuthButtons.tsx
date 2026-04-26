import { getProviderDisplayName, startOAuthLogin, useOAuthProviders } from '@/features/auth/hooks';
import { Alert, Button, Divider, Group, Skeleton, Stack, Text } from '@mantine/core';
import { IconAlertCircle, IconBrandFacebook, IconBrandGithub, IconBrandGoogle } from '@tabler/icons-react';
import { useTranslation } from 'react-i18next';

/**
 * Get the icon component for an OAuth provider
 */
function getProviderIcon(provider: string) {
  switch (provider) {
    case 'github':
      return <IconBrandGithub size={20} />;
    case 'google':
      return <IconBrandGoogle size={20} />;
    case 'facebook':
      return <IconBrandFacebook size={20} />;
    default:
      return null;
  }
}

/**
 * Get button color for an OAuth provider (using Mantine colors)
 */
function getProviderColor(provider: string): string {
  switch (provider) {
    case 'github':
      return 'dark';
    case 'google':
      return 'red';
    case 'facebook':
      return 'indigo';
    default:
      return 'gray';
  }
}

interface OAuthButtonsProps {
  /** Optional loading state from parent component */
  isLoading?: boolean;
  /** Optional callback when OAuth flow starts */
  onOAuthStart?: (provider: string) => void;
}

/**
 * OAuthButtons component renders OAuth login buttons for enabled providers
 * 
 * This component:
 * - Fetches enabled OAuth providers from the backend
 * - Renders a button for each enabled provider
 * - Handles OAuth login flow by redirecting to the backend OAuth endpoint
 * 
 * @example
 * ```tsx
 * <OAuthButtons />
 * ```
 */
export function OAuthButtons({ isLoading: externalLoading, onOAuthStart }: OAuthButtonsProps) {
  const { t } = useTranslation();
  const { data, isLoading, isError } = useOAuthProviders();

  // Get enabled providers (defensive: handle undefined providers)
  const enabledProviders = data?.providers?.filter(p => p.enabled) ?? [];

  // Handle OAuth button click
  const handleOAuthClick = (provider: string) => {
    onOAuthStart?.(provider);
    startOAuthLogin(provider);
  };

  // Loading state
  if (isLoading) {
    return (
      <Stack gap="xs">
        <Skeleton height={42} radius="md" />
      </Stack>
    );
  }

  // Error state - silently fail (OAuth is optional)
  if (isError) {
    return null;
  }

  // No enabled providers
  if (enabledProviders.length === 0) {
    return null;
  }

  return (
    <Stack gap="md">
      {/* OAuth buttons */}
      <Stack gap="xs">
        {enabledProviders.map((provider) => (
          <Button
            key={provider.name}
            variant="default"
            color={getProviderColor(provider.name)}
            leftSection={getProviderIcon(provider.name)}
            onClick={() => handleOAuthClick(provider.name)}
            loading={externalLoading}
            fullWidth
            size="md"
            aria-label={t('auth.oauth.continueWith', { provider: getProviderDisplayName(provider.name) })}
          >
            {t('auth.oauth.continueWith', { provider: getProviderDisplayName(provider.name) })}
          </Button>
        ))}
      </Stack>

      {/* Divider */}
      <Divider label={t('common.or')} labelPosition="center" />
    </Stack>
  );
}

interface LinkedOAuthAccount {
  id: string;
  provider: string;
  provider_email?: string;
  created_at: string;
}

interface OAuthAccountsListProps {
  /** Linked OAuth accounts */
  accounts: LinkedOAuthAccount[];
  /** Callback when unlink is requested */
  onUnlink: (provider: string) => void;
  /** Whether unlink is in progress */
  isUnlinking?: boolean;
  /** Provider currently being unlinked */
  unlinkingProvider?: string;
  /** Whether unlink is disabled (last auth method) */
  unlinkDisabled?: boolean;
}

/**
 * OAuthAccountsList component renders a list of linked OAuth accounts
 * with the ability to unlink them
 */
export function OAuthAccountsList({
  accounts,
  onUnlink,
  isUnlinking,
  unlinkingProvider,
  unlinkDisabled,
}: OAuthAccountsListProps) {
  const { t } = useTranslation();

  if (accounts.length === 0) {
    return (
      <Text c="dimmed" size="sm" ta="center">
        {t('auth.oauth.noLinkedAccounts')}
      </Text>
    );
  }

  return (
    <Stack gap="xs">
      {accounts.map((account) => (
        <Group key={account.id} justify="space-between" wrap="nowrap">
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
            onClick={() => onUnlink(account.provider)}
            loading={isUnlinking && unlinkingProvider === account.provider}
            disabled={unlinkDisabled}
            aria-label={t('auth.oauth.unlink', { provider: getProviderDisplayName(account.provider) })}
          >
            {t('auth.oauth.unlinkButton')}
          </Button>
        </Group>
      ))}
      
      {unlinkDisabled && (
        <Alert color="yellow" icon={<IconAlertCircle size={16} />}>
          <Text size="xs">{t('auth.oauth.cannotUnlinkLast')}</Text>
        </Alert>
      )}
    </Stack>
  );
}

interface LinkOAuthButtonsProps {
  /** Already linked providers */
  linkedProviders: string[];
  /** Optional loading state */
  isLoading?: boolean;
}

/**
 * LinkOAuthButtons component renders buttons to link new OAuth providers
 * Only shows providers that are enabled but not yet linked
 */
export function LinkOAuthButtons({ linkedProviders, isLoading: externalLoading }: LinkOAuthButtonsProps) {
  const { t } = useTranslation();
  const { data, isLoading } = useOAuthProviders();

  // Get providers that are enabled but not linked
  const availableProviders = (data?.providers ?? []).filter(
    p => p.enabled && !linkedProviders.includes(p.name)
  );

  if (isLoading) {
    return <Skeleton height={36} radius="md" />;
  }

  if (availableProviders.length === 0) {
    return null;
  }

  return (
    <Stack gap="xs">
      <Text size="sm" fw={500}>{t('auth.oauth.linkAccount')}</Text>
      <Group gap="xs">
        {availableProviders.map((provider) => (
          <Button
            key={provider.name}
            variant="outline"
            leftSection={getProviderIcon(provider.name)}
            onClick={() => startOAuthLogin(provider.name)}
            loading={externalLoading}
            size="sm"
          >
            {getProviderDisplayName(provider.name)}
          </Button>
        ))}
      </Group>
    </Stack>
  );
}
