import { useAuth } from '@/contexts/AuthContext'
import { useUpdateUserPreferences, useUserPreferences } from '@/features/user/hooks/useUserPreferences'
import type { DigestFrequency } from '@/features/user/types/preferences'
import { apiRequest } from '@/lib/api'
import { Alert, Button, Group, Select, Stack, Text, Title } from '@mantine/core'
import { useForm } from '@mantine/form'
import { IconAlertCircle, IconCheck, IconMail, IconSend } from '@tabler/icons-react'
import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'

const FREQUENCY_OPTIONS: { value: DigestFrequency; labelKey: string }[] = [
  { value: 'never', labelKey: 'settings.notificationSettings.frequencies.never' },
  { value: 'daily', labelKey: 'settings.notificationSettings.frequencies.daily' },
  { value: 'weekly', labelKey: 'settings.notificationSettings.frequencies.weekly' },
  { value: 'monthly', labelKey: 'settings.notificationSettings.frequencies.monthly' },
]

const RESEND_COOLDOWN_SECONDS = 60

export function NotificationSettings() {
  const { t } = useTranslation()
  const { user } = useAuth()
  const { data: preferences, isLoading } = useUserPreferences()
  const { mutate: updatePreferences, isPending, isSuccess, isError } = useUpdateUserPreferences()

  const [isResending, setIsResending] = useState(false)
  const [resendSuccess, setResendSuccess] = useState(false)
  const [resendError, setResendError] = useState(false)
  const [cooldown, setCooldown] = useState(0)

  const form = useForm<{ digest_frequency: DigestFrequency }>({
    initialValues: {
      digest_frequency: 'never',
    },
  })

  const hasInitialized = useRef(false)

  useEffect(() => {
    if (preferences && !hasInitialized.current) {
      form.setValues({
        digest_frequency: preferences.digest_frequency || 'never',
      })
      hasInitialized.current = true
    }
  }, [preferences]) // eslint-disable-line react-hooks/exhaustive-deps

  // Cooldown timer
  useEffect(() => {
    if (cooldown <= 0) return
    const interval = setInterval(() => {
      setCooldown((prev) => {
        if (prev <= 1) {
          clearInterval(interval)
          return 0
        }
        return prev - 1
      })
    }, 1000)
    return () => clearInterval(interval)
  }, [cooldown])

  const handleResendVerification = async () => {
    if (!user?.email || isResending || cooldown > 0) return

    setIsResending(true)
    setResendSuccess(false)
    setResendError(false)

    try {
      await apiRequest('/auth/resend-verification', {
        method: 'POST',
        body: JSON.stringify({ email: user.email }),
      })
      setResendSuccess(true)
      setCooldown(RESEND_COOLDOWN_SECONDS)
    } catch {
      setResendError(true)
    } finally {
      setIsResending(false)
    }
  }

  const handleSubmit = (values: { digest_frequency: DigestFrequency }) => {
    updatePreferences({ digest_frequency: values.digest_frequency })
  }

  const emailVerified = user?.email_verified ?? false
  const selectedFrequency = form.values.digest_frequency

  if (isLoading) {
    return null
  }

  return (
    <form onSubmit={form.onSubmit(handleSubmit)}>
      <Stack gap="lg">
        <div>
          <Title order={4}>{t('settings.notificationSettings.title')}</Title>
          <Text size="sm" c="dimmed" mt={4}>
            {t('settings.notificationSettings.description')}
          </Text>
        </div>

        {!emailVerified && (
          <Alert
            icon={<IconAlertCircle size={16} />}
            title={t('settings.notificationSettings.emailNotVerified')}
            color="yellow"
            variant="light"
          >
            <Stack gap="sm">
              <Text size="sm">{t('settings.notificationSettings.emailNotVerifiedAction')}</Text>
              <Group>
                <Button
                  size="xs"
                  variant="light"
                  color="yellow"
                  leftSection={<IconSend size={14} />}
                  loading={isResending}
                  disabled={cooldown > 0}
                  onClick={handleResendVerification}
                >
                  {cooldown > 0
                    ? t('settings.notificationSettings.resendIn', { seconds: cooldown })
                    : t('settings.notificationSettings.resendVerification')}
                </Button>
              </Group>
              {resendSuccess && (
                <Text size="sm" c="green">{t('settings.notificationSettings.resendSuccess')}</Text>
              )}
              {resendError && (
                <Text size="sm" c="red">{t('settings.notificationSettings.resendError')}</Text>
              )}
            </Stack>
          </Alert>
        )}

        <Select
          label={t('settings.notificationSettings.digestFrequency')}
          description={t('settings.notificationSettings.digestFrequencyDescription')}
          data={FREQUENCY_OPTIONS.map((opt) => ({
            value: opt.value,
            label: t(opt.labelKey),
          }))}
          disabled={!emailVerified}
          {...form.getInputProps('digest_frequency')}
        />

        {selectedFrequency !== 'never' && emailVerified && (
          <Alert icon={<IconMail size={16} />} color="blue" variant="light">
            {t(`settings.notificationSettings.scheduleInfo.${selectedFrequency}`)}
          </Alert>
        )}

        {isSuccess && (
          <Alert icon={<IconCheck size={16} />} color="green" variant="light">
            {t('settings.notificationSettings.saved')}
          </Alert>
        )}

        {isError && (
          <Alert icon={<IconAlertCircle size={16} />} color="red" variant="light">
            {t('settings.notificationSettings.saveError')}
          </Alert>
        )}

        <Group>
          <Button
            type="submit"
            loading={isPending}
            disabled={!emailVerified}
          >
            {t('common.save')}
          </Button>
        </Group>
      </Stack>
    </form>
  )
}
