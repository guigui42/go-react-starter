/**
 * Migration History Component
 *
 * Displays database migration status showing which migrations have been applied,
 * when they were applied, which are pending, and which have failed with error details.
 * Admin-only component for system health monitoring.
 */

import { useMigrationStatus } from '@/features/admin/hooks'
import type { MigrationStatus } from '@/features/admin/types'
import {
  Alert,
  Badge,
  Code,
  Collapse,
  Group,
  Paper,
  Skeleton,
  Stack,
  Table,
  Text,
  Title,
  Tooltip,
} from '@mantine/core'
import { IconAlertTriangle, IconChevronDown, IconChevronRight } from '@tabler/icons-react'
import { useState } from 'react'

function formatAppliedAt(iso: string): string {
  return new Date(iso).toLocaleString()
}

function StatusBadge({ migration }: { migration: MigrationStatus }) {
  if (migration.failed) {
    return (
      <Badge size="sm" variant="filled" color="red">
        Failed
      </Badge>
    )
  }
  if (migration.applied) {
    return (
      <Badge size="sm" variant="filled" color="green">
        Applied
      </Badge>
    )
  }
  return (
    <Badge size="sm" variant="filled" color="orange">
      Pending
    </Badge>
  )
}

function ErrorDetail({ errorMessage }: { errorMessage: string }) {
  const [opened, setOpened] = useState(false)

  return (
    <>
      <Group
        gap={4}
        style={{ cursor: 'pointer' }}
        onClick={() => setOpened((o) => !o)}
        role="button"
        tabIndex={0}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault()
            setOpened((o) => !o)
          }
        }}
        aria-expanded={opened}
        aria-label="Toggle error details"
      >
        {opened ? <IconChevronDown size={14} /> : <IconChevronRight size={14} />}
        <Text size="xs" c="red">
          View error
        </Text>
      </Group>
      <Collapse expanded={opened}>
        <Code block mt={4} style={{ fontSize: 11 }} c="red">
          {errorMessage}
        </Code>
      </Collapse>
    </>
  )
}

function MigrationRow({ migration }: { migration: MigrationStatus }) {
  return (
    <Table.Tr style={migration.failed ? { backgroundColor: 'var(--mantine-color-red-light)' } : undefined}>
      <Table.Td>
        <Text size="sm" fw={600} style={{ fontFamily: 'monospace' }}>
          {migration.version}
        </Text>
      </Table.Td>
      <Table.Td>
        <Text size="sm">{migration.name}</Text>
      </Table.Td>
      <Table.Td>
        <StatusBadge migration={migration} />
      </Table.Td>
      <Table.Td>
        {migration.applied_at ? (
          <Tooltip label={migration.applied_at} withArrow>
            <Text size="sm">{formatAppliedAt(migration.applied_at)}</Text>
          </Tooltip>
        ) : (
          <Text size="sm" c="dimmed">—</Text>
        )}
      </Table.Td>
      <Table.Td>
        {migration.error_message ? (
          <ErrorDetail errorMessage={migration.error_message} />
        ) : (
          <Text size="sm" c="dimmed">—</Text>
        )}
      </Table.Td>
    </Table.Tr>
  )
}

export function MigrationHistory() {
  const { data: migrations, isLoading } = useMigrationStatus()

  const appliedCount = migrations?.filter((m) => m.applied).length ?? 0
  const pendingCount = migrations?.filter((m) => !m.applied && !m.failed).length ?? 0
  const failedCount = migrations?.filter((m) => m.failed).length ?? 0

  return (
    <Stack gap="md">
      <Group justify="space-between" align="flex-end">
        <Title order={4}>Database Migrations</Title>
        {migrations && (
          <Group gap="sm">
            <Badge variant="light" color="green" size="sm">
              {appliedCount} applied
            </Badge>
            {pendingCount > 0 && (
              <Badge variant="light" color="orange" size="sm">
                {pendingCount} pending
              </Badge>
            )}
            {failedCount > 0 && (
              <Badge variant="light" color="red" size="sm">
                {failedCount} failed
              </Badge>
            )}
          </Group>
        )}
      </Group>

      {failedCount > 0 && (
        <Alert
          icon={<IconAlertTriangle size={16} />}
          title="Migration failures detected"
          color="red"
          variant="light"
        >
          {failedCount} migration{failedCount > 1 ? 's have' : ' has'} failed.
          Check the error details below and re-run migrations after fixing the issue.
        </Alert>
      )}

      <Paper withBorder>
        {isLoading ? (
          <Stack gap="xs" p="md">
            {Array.from({ length: 5 }).map((_, i) => (
              <Skeleton key={i} height={36} />
            ))}
          </Stack>
        ) : !migrations || migrations.length === 0 ? (
          <Text p="xl" ta="center" c="dimmed">
            No migrations registered
          </Text>
        ) : (
          <Table striped highlightOnHover withTableBorder>
            <Table.Thead>
              <Table.Tr>
                <Table.Th w={100}>Version</Table.Th>
                <Table.Th>Name</Table.Th>
                <Table.Th w={100}>Status</Table.Th>
                <Table.Th w={200}>Applied At</Table.Th>
                <Table.Th w={150}>Error</Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {migrations.map((migration) => (
                <MigrationRow key={migration.version} migration={migration} />
              ))}
            </Table.Tbody>
          </Table>
        )}
      </Paper>
    </Stack>
  )
}
