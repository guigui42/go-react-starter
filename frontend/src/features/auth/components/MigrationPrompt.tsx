import { Alert, Button, Group, Stack, Text } from '@mantine/core';
import { IconShieldLock, IconX } from '@tabler/icons-react';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';

interface MigrationPromptProps {
  onMigrate: () => void;
  onDismiss?: () => void;
  isDismissible?: boolean;
}

export function MigrationPrompt({ onMigrate, onDismiss, isDismissible = true }: MigrationPromptProps) {
  const { t } = useTranslation();
  const [isDismissed, setIsDismissed] = useState(false);

  if (isDismissed) {
    return null;
  }

  const handleDismiss = () => {
    setIsDismissed(true);
    onDismiss?.();
  };

  return (
    <Alert
      icon={<IconShieldLock size={20} />}
      title={t('passkey.migration.title')}
      color="blue"
      withCloseButton={isDismissible}
      onClose={isDismissible ? handleDismiss : undefined}
    >
      <Stack gap="sm">
        <Text size="sm">{t('passkey.migration.description')}</Text>
        <Group>
          <Button size="sm" onClick={onMigrate}>
            {t('passkey.migration.upgrade_button')}
          </Button>
          {isDismissible && (
            <Button size="sm" variant="subtle" onClick={handleDismiss} leftSection={<IconX size={14} />}>
              {t('passkey.migration.dismiss_button')}
            </Button>
          )}
        </Group>
      </Stack>
    </Alert>
  );
}
