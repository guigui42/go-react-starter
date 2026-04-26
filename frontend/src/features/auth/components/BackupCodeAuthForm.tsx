import { Alert, Button, Group, Stack, Text, TextInput } from '@mantine/core';
import { IconAlertCircle, IconArrowLeft, IconKey } from '@tabler/icons-react';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';

interface BackupCodeAuthFormProps {
  onSuccess: (token: string) => void;
  onError?: (error: Error) => void;
  onBack?: () => void;
}

export function BackupCodeAuthForm({ onSuccess, onError, onBack }: BackupCodeAuthFormProps) {
  const { t } = useTranslation();
  const [code, setCode] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!code.trim()) {
      setError(t('backup_codes.auth.empty_code'));
      return;
    }

    setError(null);
    setIsSubmitting(true);

    try {
      // Remove any formatting (hyphens, spaces)
      const cleanCode = code.replace(/[-\s]/g, '');

      const response = await fetch('/api/auth/backup-code', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ code: cleanCode }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || t('backup_codes.auth.invalid_code'));
      }

      const data = await response.json();
      onSuccess(data.token);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : t('backup_codes.auth.failed');
      setError(errorMessage);
      onError?.(err instanceof Error ? err : new Error(errorMessage));
    } finally {
      setIsSubmitting(false);
    }
  };

  const formatCodeInput = (value: string) => {
    // Auto-format as XXXX-XXXX-XXXX-XXXX
    const cleaned = value.replace(/[^a-zA-Z0-9]/g, '').toUpperCase();
    const parts = cleaned.match(/.{1,4}/g) || [];
    return parts.join('-').substring(0, 19); // Max length: 16 chars + 3 hyphens
  };

  const handleCodeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const formatted = formatCodeInput(e.currentTarget.value);
    setCode(formatted);
    setError(null);
  };

  return (
    <form onSubmit={handleSubmit}>
      <Stack gap="md">
        <div>
          <Text size="sm" fw={500} mb="xs">
            {t('backup_codes.auth.title')}
          </Text>
          <Text size="xs" c="dimmed">
            {t('backup_codes.auth.description')}
          </Text>
        </div>

        <TextInput
          label={t('backup_codes.auth.code_label')}
          placeholder="XXXX-XXXX-XXXX-XXXX"
          value={code}
          onChange={handleCodeChange}
          leftSection={<IconKey size={16} />}
          maxLength={19}
          required
          autoComplete="off"
          error={error}
          styles={{
            input: {
              fontFamily: 'monospace',
              letterSpacing: '0.05em',
            },
          }}
        />

        {error && (
          <Alert icon={<IconAlertCircle size={16} />} color="red" title={t('common.error')}>
            {error}
          </Alert>
        )}

        <Group grow>
          {onBack && (
            <Button variant="default" leftSection={<IconArrowLeft size={16} />} onClick={onBack}>
              {t('common.back')}
            </Button>
          )}
          <Button type="submit" loading={isSubmitting}>
            {t('backup_codes.auth.submit_button')}
          </Button>
        </Group>

        <Alert icon={<IconAlertCircle size={16} />} color="blue" variant="light">
          <Text size="xs">{t('backup_codes.auth.warning')}</Text>
        </Alert>
      </Stack>
    </form>
  );
}
