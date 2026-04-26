/**
 * Admin Dashboard Component
 *
 * Main admin dashboard page showing system stats, user management, and admin tools.
 * Requires admin privileges to access.
 */

import { useAdminStats, useAdminUsers } from '@/features/admin/hooks'
import { ApiError } from '@/lib/api'
import {
    Alert,
    Box,
    Center,
    Container,
    Loader,
    Stack,
    Tabs,
    Text,
    Title,
} from '@mantine/core'
import { IconAlertCircle, IconDatabase, IconHistory, IconLayoutDashboard, IconLogs, IconMail, IconShieldCheck } from '@tabler/icons-react'
import { useState } from 'react'
import { AdminStatsCards } from './AdminStatsCards'
import { AuditLogViewer } from './AuditLogViewer'
import { EmailTester } from './EmailTester'
import { LogViewer } from './LogViewer'
import { MigrationHistory } from './MigrationHistory'
import { UserList } from './UserList'

export function AdminDashboard() {
  const [activeTab, setActiveTab] = useState<string | null>('dashboard')
  
  const isDashboardActive = activeTab === 'dashboard'
  
  const { data: stats, isLoading: statsLoading, error: statsError } = useAdminStats({ enabled: isDashboardActive })
  const { data: users, isLoading: usersLoading, error: usersError } = useAdminUsers({ enabled: isDashboardActive })

  // Check for 403 Forbidden error (non-admin user) using status code
  const isForbidden = [statsError, usersError].some(
    (error) => error instanceof ApiError && error.status === 403
  )

  if (isForbidden) {
    return (
      <Container size="xl" py="xl" px={{ base: 0, sm: 'md' }}>
        <Center>
          <Alert
            icon={<IconAlertCircle size={16} />}
            title="Access Denied"
            color="red"
            variant="light"
          >
            You do not have permission to access the admin dashboard.
            Admin access is restricted to designated administrators.
          </Alert>
        </Center>
      </Container>
    )
  }

  const hasError = statsError || usersError
  if (hasError) {
    return (
      <Container size="xl" py="xl" px={{ base: 0, sm: 'md' }}>
        <Center>
          <Alert
            icon={<IconAlertCircle size={16} />}
            title="Error Loading Dashboard"
            color="red"
            variant="light"
          >
            <Stack gap="xs">
              <Text size="sm">Failed to load admin dashboard data.</Text>
              {statsError && <Text size="xs" c="dimmed">Stats: {statsError.message}</Text>}
              {usersError && <Text size="xs" c="dimmed">Users: {usersError.message}</Text>}
            </Stack>
          </Alert>
        </Center>
      </Container>
    )
  }

  const isInitialLoading = statsLoading && usersLoading
  if (isInitialLoading) {
    return (
      <Container size="xl" py="xl" px={{ base: 0, sm: 'md' }}>
        <Center py="xl">
          <Loader size="lg" />
        </Center>
      </Container>
    )
  }

  return (
    <Box py="xl">
      <Container size="xl">
        <Stack gap="xl">
          {/* Page Header */}
          <Stack gap="xs">
            <Title order={2} style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <IconShieldCheck size={28} />
              Admin Dashboard
            </Title>
            <Text c="dimmed" size="sm">
              System overview and management tools for administrators
            </Text>
          </Stack>

          {/* Tabs */}
          <Tabs value={activeTab} onChange={setActiveTab} keepMounted={false}>
            <Tabs.List style={{ overflowX: 'auto', flexWrap: 'nowrap' }}>
              <Tabs.Tab value="dashboard" leftSection={<IconLayoutDashboard size={16} />}>
                Dashboard
              </Tabs.Tab>
              <Tabs.Tab value="logs" leftSection={<IconLogs size={16} />}>
                Backend Logs
              </Tabs.Tab>
              <Tabs.Tab value="audit" leftSection={<IconHistory size={16} />}>
                Audit Logs
              </Tabs.Tab>
              <Tabs.Tab value="email" leftSection={<IconMail size={16} />}>
                Email
              </Tabs.Tab>
              <Tabs.Tab value="migrations" leftSection={<IconDatabase size={16} />}>
                Migrations
              </Tabs.Tab>
            </Tabs.List>
          </Tabs>
        </Stack>
      </Container>

      {/* Tab Content - rendered outside Container for flexible width */}
      {activeTab === 'dashboard' && (
        <Container size="xl" pt="md">
          <Stack gap="xl">
            {/* Stats Cards */}
            <AdminStatsCards stats={stats} isLoading={statsLoading} />

            {/* User List */}
            <UserList users={users} isLoading={usersLoading} />
          </Stack>
        </Container>
      )}

      {activeTab === 'logs' && (
        <Container fluid px="xl" pt="md">
          <LogViewer />
        </Container>
      )}

      {activeTab === 'audit' && (
        <Container size="xl" pt="md">
          <AuditLogViewer />
        </Container>
      )}

      {activeTab === 'email' && (
        <Container size="md" pt="md">
          <EmailTester />
        </Container>
      )}

      {activeTab === 'migrations' && (
        <Container size="lg" pt="md">
          <MigrationHistory />
        </Container>
      )}
    </Box>
  )
}
