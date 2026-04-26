import { useAuth } from '@/contexts/AuthContext'
import { OAuthManagementPanel } from '@/features/auth/components/OAuthManagementPanel'
import { PasskeyManagementPanel } from '@/features/auth/components/PasskeyManagementPanel'
import { DeleteAccountSection } from '@/features/user/components/DeleteAccountSection'
import { NotificationSettings } from '@/features/user/components/NotificationSettings'
import { PrivacySettings } from '@/features/user/components/PrivacySettings'
import { ProfileSettings } from '@/features/user/components/ProfileSettings'
import {
    animateOnMount,
    fadeInUp,
    MotionBox,
    MotionStack,
    staggerContainer,
} from '@/lib/motion'
import { Center, Container, Loader, Paper, Stack, Tabs, Title } from '@mantine/core'
import { IconBell, IconLock, IconShieldLock, IconUser } from '@tabler/icons-react'
import { createFileRoute, redirect, useNavigate, useSearch } from '@tanstack/react-router'
import { useEffect } from 'react'
import { useTranslation } from 'react-i18next'

const VALID_TABS = ['profile', 'security', 'privacy', 'notifications'] as const
type SettingsTab = (typeof VALID_TABS)[number]

export const Route = createFileRoute('/settings')({
  validateSearch: (search: Record<string, unknown>): { tab?: SettingsTab } => {
    const tab = search.tab as string | undefined
    if (tab && VALID_TABS.includes(tab as SettingsTab)) {
      return { tab: tab as SettingsTab }
    }
    return {}
  },
  beforeLoad: ({ context }) => {
    // Require authentication
    if (!context.auth.isLoading && !context.auth.isAuthenticated) {
      throw redirect({ to: '/login' })
    }
  },
  component: SettingsPage,
})

function SettingsPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { tab } = useSearch({ from: '/settings' })
  const { isLoading, isAuthenticated } = useAuth()

  const activeTab = tab || 'profile'

  const handleTabChange = (value: string | null) => {
    if (value) {
      navigate({ to: '/settings', search: value === 'profile' ? {} : { tab: value as SettingsTab }, replace: true })
    }
  }

  // Redirect to login if not authenticated after loading completes
  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      navigate({ to: '/login', replace: true })
    }
  }, [isLoading, isAuthenticated, navigate])

  // Show loading state while authentication is being verified
  if (isLoading) {
    return (
      <Center h="100dvh">
        <Loader size="lg" />
      </Center>
    )
  }

  // Don't render content if not authenticated (will redirect)
  if (!isAuthenticated) {
    return (
      <Center h="100dvh">
        <Loader size="lg" />
      </Center>
    )
  }

  return (
    <Container size="lg" my={{ base: 'md', sm: 40 }} px={{ base: 0, sm: 'md' }}>
      <MotionStack variants={staggerContainer} {...animateOnMount} gap="lg">
        <MotionBox variants={fadeInUp}>
          <Title mb="lg">{t('settings.title')}</Title>
        </MotionBox>

        <MotionBox variants={fadeInUp}>
          <Tabs value={activeTab} onChange={handleTabChange}>
            <Tabs.List style={{ overflowX: 'auto', flexWrap: 'nowrap' }}>
              <Tabs.Tab value="profile" leftSection={<IconUser size={16} />}>
                {t('settings.profile')}
              </Tabs.Tab>
              <Tabs.Tab value="security" leftSection={<IconShieldLock size={16} />}>
                {t('settings.security')}
              </Tabs.Tab>
              <Tabs.Tab value="privacy" leftSection={<IconLock size={16} />}>
                {t('settings.privacy')}
              </Tabs.Tab>
              <Tabs.Tab value="notifications" leftSection={<IconBell size={16} />}>
                {t('settings.notifications')}
              </Tabs.Tab>
            </Tabs.List>

            <Tabs.Panel value="profile" pt="md">
              <Paper withBorder shadow="sm" p="md" radius="md">
                <ProfileSettings />
              </Paper>
            </Tabs.Panel>

            <Tabs.Panel value="security" pt="md">
              <Stack gap="lg">
                <Paper withBorder shadow="sm" p="md" radius="md">
                  <PasskeyManagementPanel />
                </Paper>

                <Paper withBorder shadow="sm" p="md" radius="md">
                  <OAuthManagementPanel />
                </Paper>

                <DeleteAccountSection />
              </Stack>
            </Tabs.Panel>

            <Tabs.Panel value="privacy" pt="md">
              <Paper withBorder shadow="sm" p="md" radius="md">
                <PrivacySettings />
              </Paper>
            </Tabs.Panel>

            <Tabs.Panel value="notifications" pt="md">
              <Paper withBorder shadow="sm" p="md" radius="md">
                <NotificationSettings />
              </Paper>
            </Tabs.Panel>
          </Tabs>
        </MotionBox>
      </MotionStack>
    </Container>
  )
}
