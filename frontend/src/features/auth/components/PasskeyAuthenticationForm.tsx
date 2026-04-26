import { isWebAuthnSupported, passkeyService } from '@/features/auth/services';
import { Alert, Button, Group, Stack, Text } from '@mantine/core';
import { IconAlertCircle, IconFingerprint, IconKey } from '@tabler/icons-react';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';

interface PasskeyAuthenticationFormProps {
  email: string;
  onSuccess?: (token: string) => void;
  onError?: (error: Error) => void;
  onUseBackupCode?: () => void;
  showBackupCodeOption?: boolean;
}

export function PasskeyAuthenticationForm({
  email,
  onSuccess,
  onError,
  onUseBackupCode,
  showBackupCodeOption = true,
}: PasskeyAuthenticationFormProps) {
  const { t } = useTranslation();
  const [isAuthenticating, setIsAuthenticating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isSupported = isWebAuthnSupported();

  const handleAuthenticate = async () => {
    if (!isSupported) {
      setError(t('passkey.errors.not_supported'));
      return;
    }

    setError(null);
    setIsAuthenticating(true);

    try {
      const result = await passkeyService.authenticate(email);
      onSuccess?.(result.token);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : t('passkey.errors.authentication_failed');
      setError(errorMessage);
      onError?.(err instanceof Error ? err : new Error(errorMessage));
    } finally {
      setIsAuthenticating(false);
    }
  };

  if (!isSupported) {
    return (
      <Alert icon={<IconAlertCircle size={16} />} title={t('passkey.not_supported_title')} color="yellow">
        {t('passkey.not_supported_message')}
      </Alert>
    );
  }

  return (
    <Stack gap="md">
      <Text size="sm" c="dimmed">
        {t('passkey.authentication.description')}
      </Text>

      {error && (
        <Alert icon={<IconAlertCircle size={16} />} title={t('common.error')} color="red">
          {error}
        </Alert>
      )}

      <Button
        leftSection={<IconFingerprint size={20} />}
        onClick={handleAuthenticate}
        loading={isAuthenticating}
        fullWidth
        size="lg"
      >
        {t('passkey.authentication.sign_in_button')}
      </Button>

      {showBackupCodeOption && onUseBackupCode && (
        <Group justify="center">
          <Button
            variant="subtle"
            leftSection={<IconKey size={16} />}
            onClick={onUseBackupCode}
            size="sm"
          >
            {t('passkey.authentication.use_backup_code')}
          </Button>
        </Group>
      )}
    </Stack>
  );
}
