import { ActionIcon, Alert, Code, CopyButton, Group, Paper, Stack, Text, Tooltip } from '@mantine/core';
import { IconAlertCircle, IconCheck, IconCopy } from '@tabler/icons-react';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

interface BackupCodeDisplayProps {
  codes: string[];
  onAcknowledge?: () => void;
}

export function BackupCodeDisplay({ codes, onAcknowledge }: BackupCodeDisplayProps) {
  const { t } = useTranslation();
  const [hasAcknowledged, setHasAcknowledged] = useState(false);

  useEffect(() => {
    // Auto-acknowledge after 5 seconds if the callback exists
    if (onAcknowledge && !hasAcknowledged) {
      const timer = setTimeout(() => {
        setHasAcknowledged(true);
      }, 5000);
      return () => clearTimeout(timer);
    }
  }, [hasAcknowledged, onAcknowledge]);

  const formatCode = (code: string) => {
    // Format as XXXX-XXXX-XXXX-XXXX
    return code.match(/.{1,4}/g)?.join('-') || code;
  };

  const allCodesText = codes.map(formatCode).join('\n');

  return (
    <Stack gap="md">
      <Alert icon={<IconAlertCircle size={16} />} title={t('backup_codes.warning_title')} color="yellow">
        <Text size="sm">{t('backup_codes.warning_message')}</Text>
      </Alert>

      <Text size="sm" fw={500}>
        {t('backup_codes.save_instruction')}
      </Text>

      <Paper withBorder p="md" bg="gray.0" style={{ fontFamily: 'monospace' }}>
        <Stack gap="xs">
          {codes.map((code, index) => (
            <Group key={index} justify="space-between">
              <Code style={{ fontSize: '14px' }}>{formatCode(code)}</Code>
              <CopyButton value={code}>
                {({ copied, copy }) => (
                  <Tooltip label={copied ? t('common.copied') : t('common.copy')}>
                    <ActionIcon color={copied ? 'teal' : 'gray'} variant="subtle" onClick={copy}>
                      {copied ? <IconCheck size={16} /> : <IconCopy size={16} />}
                    </ActionIcon>
                  </Tooltip>
                )}
              </CopyButton>
            </Group>
          ))}
        </Stack>
      </Paper>

      <CopyButton value={allCodesText}>
        {({ copied, copy }) => (
          <Group justify="center">
            <ActionIcon.Group>
              <ActionIcon variant="light" size="lg" onClick={copy}>
                {copied ? <IconCheck size={18} /> : <IconCopy size={18} />}
              </ActionIcon>
            </ActionIcon.Group>
            <Text size="sm" c="dimmed">
              {copied ? t('backup_codes.copied_all') : t('backup_codes.copy_all')}
            </Text>
          </Group>
        )}
      </CopyButton>

      <Alert icon={<IconAlertCircle size={16} />} color="blue">
        <Text size="xs">{t('backup_codes.storage_suggestion')}</Text>
      </Alert>
    </Stack>
  );
}
