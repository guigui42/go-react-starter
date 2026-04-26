import { useUpdateUserPreferences, useUserPreferences } from '@/features/user/hooks/useUserPreferences'
import type { ColorScheme, UpdateUserPreferencesRequest } from '@/features/user/types/preferences'
import i18n from '@/i18n/config'
import { Alert, Button, Group, SegmentedControl, Select, Stack, Text, Title, useMantineColorScheme } from '@mantine/core'
import { useForm } from '@mantine/form'
import { IconAlertCircle, IconCheck, IconDeviceDesktop, IconMoon, IconSun } from '@tabler/icons-react'
import { useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'

const SUPPORTED_LANGUAGES = [
  { value: 'en', label: 'English' },
  { value: 'fr', label: 'Français' },
]

const SUPPORTED_COUNTRIES = [
  { value: 'FR', label: 'France' },
  { value: 'US', label: 'United States' },
  { value: 'GB', label: 'United Kingdom' },
  { value: 'DE', label: 'Germany' },
  { value: 'CH', label: 'Switzerland' },
  { value: 'CA', label: 'Canada' },
  { value: 'AU', label: 'Australia' },
  { value: 'JP', label: 'Japan' },
  { value: 'BE', label: 'Belgium' },
  { value: 'ES', label: 'Spain' },
  { value: 'IT', label: 'Italy' },
  { value: 'NL', label: 'Netherlands' },
]

export function ProfileSettings() {
  const { t } = useTranslation()
  const { data: preferences, isLoading, error } = useUserPreferences()
  const { mutate: updatePreferences, isPending, isSuccess, isError } = useUpdateUserPreferences()
  const { setColorScheme: setMantineColorScheme } = useMantineColorScheme()

  const form = useForm<UpdateUserPreferencesRequest>({
    initialValues: {
      language: 'en',
      country_code: 'FR',
      color_scheme: 'auto',
    },
  })

  // Track if we've done the initial form population
  const hasInitialized = useRef(false)

  // Initialize form ONCE when preferences first load
  // Using a ref prevents re-running after mutations update the cache
  useEffect(() => {
    if (preferences && !hasInitialized.current) {
      hasInitialized.current = true
      form.setValues({
        language: preferences.language,
        country_code: preferences.country_code,
        color_scheme: preferences.color_scheme,
      })
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [preferences])

  const handleSubmit = (values: UpdateUserPreferencesRequest) => {
    if (values.color_scheme) {
      setMantineColorScheme(values.color_scheme)
    }
    if (values.language) {
      i18n.changeLanguage(values.language)
    }
    // Save ALL values to backend in a single request
    updatePreferences(values)
  }

  if (isLoading) {
    return <Text>{t('common.loading')}</Text>
  }

  if (error) {
    return (
      <Alert icon={<IconAlertCircle size={16} />} title={t('common.error')} color="red">
        {t('settings.profileSettings.loadError')}
      </Alert>
    )
  }

  return (
    <Stack gap="lg">
      <div>
        <Title order={3}>{t('settings.profileSettings.title')}</Title>
        <Text size="sm" c="dimmed" mt={4}>
          {t('settings.profileSettings.description')}
        </Text>
      </div>

      {isSuccess && (
        <Alert icon={<IconCheck size={16} />} title={t('common.success')} color="green">
          {t('settings.profileSettings.saveSuccess')}
        </Alert>
      )}

      {isError && (
        <Alert icon={<IconAlertCircle size={16} />} title={t('common.error')} color="red">
          {t('settings.profileSettings.saveError')}
        </Alert>
      )}

      <form onSubmit={form.onSubmit(handleSubmit)}>
        <Stack gap="md">
          <div>
            <Text size="sm" fw={500} mb={4}>
              {t('settings.profileSettings.colorScheme')}
            </Text>
            <Text size="xs" c="dimmed" mb="xs">
              {t('settings.profileSettings.colorSchemeDescription')}
            </Text>
            <SegmentedControl
              value={form.values.color_scheme || 'auto'}
              onChange={(value) => form.setFieldValue('color_scheme', value as ColorScheme)}
              data={[
                {
                  value: 'light',
                  label: (
                    <Group gap={4}>
                      <IconSun size={16} />
                      <span>{t('settings.profileSettings.colorSchemes.light')}</span>
                    </Group>
                  ),
                },
                {
                  value: 'dark',
                  label: (
                    <Group gap={4}>
                      <IconMoon size={16} />
                      <span>{t('settings.profileSettings.colorSchemes.dark')}</span>
                    </Group>
                  ),
                },
                {
                  value: 'auto',
                  label: (
                    <Group gap={4}>
                      <IconDeviceDesktop size={16} />
                      <span>{t('settings.profileSettings.colorSchemes.auto')}</span>
                    </Group>
                  ),
                },
              ]}
              fullWidth
            />
          </div>

          <Select
            label={t('settings.profileSettings.language')}
            description={t('settings.profileSettings.languageDescription')}
            data={SUPPORTED_LANGUAGES}
            {...form.getInputProps('language')}
            required
          />

          <Select
            label={t('settings.profileSettings.country')}
            description={t('settings.profileSettings.countryDescription')}
            data={SUPPORTED_COUNTRIES}
            {...form.getInputProps('country_code')}
            required
            searchable
          />

          <Group justify="flex-end" mt="md">
            <Button type="submit" loading={isPending}>
              {t('common.save')}
            </Button>
          </Group>
        </Stack>
      </form>
    </Stack>
  )
}
