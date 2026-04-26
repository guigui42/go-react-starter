/**
 * Audit Log Viewer Component
 *
 * Displays audit log entries from the database with filtering, pagination, and metadata viewing.
 * Admin-only component for security monitoring and compliance.
 */

import { useAuditLogs } from '@/features/admin/hooks'
import type { AuditLog, AuditLogQuery } from '@/features/admin/types'
import {
  Badge,
  Code,
  Collapse,
  Group,
  Pagination,
  Paper,
  Select,
  Skeleton,
  Stack,
  Table,
  Text,
  Title,
  Tooltip,
} from '@mantine/core'
import { IconChevronDown, IconChevronRight } from '@tabler/icons-react'
import { useCallback, useState } from 'react'

const STATUS_COLORS: Record<string, string> = {
  success: 'green',
  failure: 'red',
}

const EVENT_CATEGORY_COLORS: Record<string, string> = {
  auth: 'blue',
  user: 'violet',
  admin: 'orange',
}

const EVENT_TYPE_OPTIONS = [
  { value: '', label: 'All events' },
  { value: 'auth.', label: 'Authentication' },
  { value: 'auth.login', label: 'Login' },
  { value: 'auth.logout', label: 'Logout' },
  { value: 'auth.register', label: 'Registration' },
  { value: 'auth.oauth', label: 'OAuth' },
  { value: 'auth.passkey', label: 'Passkey' },
  { value: 'auth.password', label: 'Password' },
  { value: 'auth.email', label: 'Email verification' },
  { value: 'user.', label: 'User management' },
  { value: 'admin.', label: 'Admin actions' },
]

const STATUS_OPTIONS = [
  { value: '', label: 'All statuses' },
  { value: 'success', label: 'Success' },
  { value: 'failure', label: 'Failure' },
]

const PAGE_SIZE = 25

function getEventCategory(eventType: string): string {
  return eventType.split('.')[0]
}

function formatTimestamp(iso: string): string {
  const date = new Date(iso)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)

  if (diffMins < 1) return 'just now'
  if (diffMins < 60) return `${diffMins}m ago`
  const diffHours = Math.floor(diffMins / 60)
  if (diffHours < 24) return `${diffHours}h ago`
  const diffDays = Math.floor(diffHours / 24)
  if (diffDays < 7) return `${diffDays}d ago`
  return date.toLocaleDateString()
}

function formatFullTimestamp(iso: string): string {
  return new Date(iso).toLocaleString()
}

function MetadataViewer({ metadata }: { metadata: Record<string, unknown> | null }) {
  const [opened, setOpened] = useState(false)

  if (!metadata) return <Text size="xs" c="dimmed">—</Text>

  const parsed = metadata

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
        aria-label="Toggle metadata"
      >
        {opened ? <IconChevronDown size={14} /> : <IconChevronRight size={14} />}
        <Text size="xs" c="blue">
          {Object.keys(parsed).length} field{Object.keys(parsed).length !== 1 ? 's' : ''}
        </Text>
      </Group>
      <Collapse expanded={opened}>
        <Code block mt={4} style={{ fontSize: 11 }}>
          {JSON.stringify(parsed, null, 2)}
        </Code>
      </Collapse>
    </>
  )
}

function AuditLogRow({ log, actors }: { log: AuditLog; actors: Record<string, string> }) {
  const category = getEventCategory(log.event_type)
  const actorEmail = log.actor_id ? actors[log.actor_id] : undefined

  return (
    <Table.Tr>
      <Table.Td>
        <Tooltip label={formatFullTimestamp(log.created_at)} withArrow>
          <Text size="xs">{formatTimestamp(log.created_at)}</Text>
        </Tooltip>
      </Table.Td>
      <Table.Td>
        <Group gap={4}>
          <Badge
            size="xs"
            variant="light"
            color={EVENT_CATEGORY_COLORS[category] ?? 'gray'}
          >
            {category}
          </Badge>
          <Text size="xs">{log.event_type}</Text>
        </Group>
      </Table.Td>
      <Table.Td>
        <Text size="xs" lineClamp={1}>{log.action}</Text>
      </Table.Td>
      <Table.Td>
        <Badge
          size="xs"
          variant="filled"
          color={STATUS_COLORS[log.status] ?? 'gray'}
        >
          {log.status}
        </Badge>
      </Table.Td>
      <Table.Td>
        {actorEmail ? (
          <Tooltip label={log.actor_id ?? ''} withArrow>
            <Text size="xs">{actorEmail}</Text>
          </Tooltip>
        ) : (
          <Text size="xs" c="dimmed">—</Text>
        )}
      </Table.Td>
      <Table.Td>
        <Text size="xs" c="dimmed" style={{ fontFamily: 'monospace' }}>
          {log.ip_address ?? '—'}
        </Text>
      </Table.Td>
      <Table.Td>
        <MetadataViewer metadata={log.metadata} />
      </Table.Td>
    </Table.Tr>
  )
}

export function AuditLogViewer() {
  const [query, setQuery] = useState<AuditLogQuery>({
    page: 1,
    page_size: PAGE_SIZE,
  })

  const { data, isLoading } = useAuditLogs(query)

  const handleEventTypeChange = useCallback((value: string | null) => {
    setQuery((q) => ({ ...q, event_type: value || undefined, page: 1 }))
  }, [])

  const handleStatusChange = useCallback((value: string | null) => {
    setQuery((q) => ({ ...q, status: value || undefined, page: 1 }))
  }, [])

  const handlePageChange = useCallback((page: number) => {
    setQuery((q) => ({ ...q, page }))
  }, [])

  return (
    <Stack gap="md">
      <Group justify="space-between" align="flex-end">
        <Title order={4}>Audit Logs</Title>
        {data && (
          <Text size="sm" c="dimmed">
            {data.total} total entries
          </Text>
        )}
      </Group>

      {/* Filters */}
      <Paper p="sm" withBorder>
        <Group gap="sm">
          <Select
            size="xs"
            label="Event type"
            data={EVENT_TYPE_OPTIONS}
            value={query.event_type ?? ''}
            onChange={handleEventTypeChange}
            clearable={false}
            w={200}
            aria-label="Filter by event type"
          />
          <Select
            size="xs"
            label="Status"
            data={STATUS_OPTIONS}
            value={query.status ?? ''}
            onChange={handleStatusChange}
            clearable={false}
            w={140}
            aria-label="Filter by status"
          />
        </Group>
      </Paper>

      {/* Table */}
      <Paper withBorder>
        {isLoading ? (
          <Stack gap="xs" p="md">
            {Array.from({ length: 5 }).map((_, i) => (
              <Skeleton key={i} height={36} />
            ))}
          </Stack>
        ) : !data || data.logs.length === 0 ? (
          <Text p="xl" ta="center" c="dimmed">
            No audit log entries found
          </Text>
        ) : (
          <>
            <Table striped highlightOnHover withTableBorder>
              <Table.Thead>
                <Table.Tr>
                  <Table.Th w={100}>Time</Table.Th>
                  <Table.Th w={220}>Event</Table.Th>
                  <Table.Th>Action</Table.Th>
                  <Table.Th w={80}>Status</Table.Th>
                  <Table.Th w={100}>Actor</Table.Th>
                  <Table.Th w={120}>IP</Table.Th>
                  <Table.Th w={120}>Metadata</Table.Th>
                </Table.Tr>
              </Table.Thead>
              <Table.Tbody>
                {data.logs.map((log) => (
                  <AuditLogRow key={log.id} log={log} actors={data.actors} />
                ))}
              </Table.Tbody>
            </Table>

            {/* Pagination */}
            {data.total_pages > 1 && (
              <Group justify="center" p="md">
                <Pagination
                  total={data.total_pages}
                  value={data.page}
                  onChange={handlePageChange}
                  size="sm"
                  aria-label="Audit log pagination"
                />
              </Group>
            )}
          </>
        )}
      </Paper>
    </Stack>
  )
}
