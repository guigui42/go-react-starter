import { ApiError } from '@/lib/api'
import { useExportUserData } from '@/features/user/hooks/useExportUserData'
import { Alert, Button, Stack, Text, Title } from '@mantine/core'
import { IconAlertCircle, IconCheck, IconDownload } from '@tabler/icons-react'
import { useTranslation } from 'react-i18next'

export function PrivacySettings() {
  const { t } = useTranslation()
  const { mutate: exportData, isPending, isSuccess, isError, error, reset } = useExportUserData()

  const handleExport = () => {
    reset() // Clear previous state
    exportData()
  }

  // Check if error is a rate limit error by checking the error code
  const isRateLimited = isError && error instanceof ApiError && error.code === 'RATE_LIMITED'

  return (
    <Stack gap="lg">
      <div>
        <Title order={3}>{t('settings.privacySettings.title')}</Title>
        <Text size="sm" c="dimmed" mt={4}>
          {t('settings.privacySettings.description')}
        </Text>
      </div>

      {isSuccess && (
        <Alert icon={<IconCheck size={16} />} title={t('common.success')} color="green">
          {t('settings.privacySettings.exportSuccess')}
        </Alert>
      )}

      {isError && (
        <Alert icon={<IconAlertCircle size={16} />} title={t('common.error')} color="red">
          {isRateLimited
            ? t('settings.privacySettings.exportRateLimited')
            : t('settings.privacySettings.exportError')}
        </Alert>
      )}

      <div>
        <Title order={4} mb="xs">{t('settings.privacySettings.dataExport.title')}</Title>
        <Text size="sm" c="dimmed" mb="md">
          {t('settings.privacySettings.dataExport.description')}
        </Text>
        <Button
          leftSection={<IconDownload size={16} />}
          onClick={handleExport}
          loading={isPending}
          disabled={isPending}
        >
          {isPending
            ? t('settings.privacySettings.dataExport.downloading')
            : t('settings.privacySettings.dataExport.button')}
        </Button>
      </div>
    </Stack>
  )
}
