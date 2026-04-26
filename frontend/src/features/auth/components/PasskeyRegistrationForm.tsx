import { isWebAuthnSupported, passkeyService } from '@/features/auth/services';
import { Alert, Button, Stack, Text, TextInput } from '@mantine/core';
import { IconAlertCircle, IconFingerprint } from '@tabler/icons-react';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';

interface PasskeyRegistrationFormProps {
  onSuccess?: (credentialId: string) => void;
  onError?: (error: Error) => void;
}

export function PasskeyRegistrationForm({ onSuccess, onError }: PasskeyRegistrationFormProps) {
  const { t } = useTranslation();
  const [friendlyName, setFriendlyName] = useState('');
  const [isRegistering, setIsRegistering] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isSupported = isWebAuthnSupported();

  const handleRegister = async () => {
    if (!isSupported) {
      setError(t('passkey.errors.not_supported'));
      return;
    }

    setError(null);
    setIsRegistering(true);

    try {
      const credential = await passkeyService.register(friendlyName || undefined);
      onSuccess?.(credential.credentialId);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : t('passkey.errors.registration_failed');
      setError(errorMessage);
      onError?.(err instanceof Error ? err : new Error(errorMessage));
    } finally {
      setIsRegistering(false);
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
        {t('passkey.registration.description')}
      </Text>

      <TextInput
        label={t('passkey.registration.friendly_name_label')}
        placeholder={t('passkey.registration.friendly_name_placeholder')}
        value={friendlyName}
        onChange={(e) => setFriendlyName(e.currentTarget.value)}
        description={t('passkey.registration.friendly_name_description')}
        maxLength={50}
      />

      {error && (
        <Alert icon={<IconAlertCircle size={16} />} title={t('common.error')} color="red">
          {error}
        </Alert>
      )}

      <Button
        leftSection={<IconFingerprint size={20} />}
        onClick={handleRegister}
        loading={isRegistering}
        fullWidth
      >
        {t('passkey.registration.create_button')}
      </Button>
    </Stack>
  );
}
