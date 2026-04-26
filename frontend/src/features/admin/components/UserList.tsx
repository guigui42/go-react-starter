/**
 * User List Component
 *
 * Displays all registered users for admin management
 */

import type { AdminUser } from '@/features/admin/types'
import {
  Badge,
  Group,
  Paper,
  ScrollArea,
  Skeleton,
  Stack,
  Switch,
  Table,
  Text,
  TextInput,
  Title,
  Tooltip,
} from '@mantine/core'
import { IconCheck, IconSearch, IconX } from '@tabler/icons-react'
import { useMemo, useState } from 'react'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'

dayjs.extend(relativeTime)

interface UserListProps {
  users?: AdminUser[]
  isLoading: boolean
}

function TableSkeleton() {
  return (
    <Stack gap="sm">
      {Array.from({ length: 5 }).map((_, i) => (
        <Skeleton key={i} height={40} radius="sm" />
      ))}
    </Stack>
  )
}

export function UserList({ users, isLoading }: UserListProps) {
  const [search, setSearch] = useState('')
  const [showTestUsers, setShowTestUsers] = useState(false)

  const filteredUsers = useMemo(() => {
    if (!users) return []
    const searchLower = search.toLowerCase()

    return users.filter((user) => {
      if (!showTestUsers && user.is_test_user) return false
      if (searchLower) {
        return user.email.toLowerCase().includes(searchLower)
      }
      return true
    })
  }, [users, search, showTestUsers])

  const testUserCount = useMemo(() => {
    if (!users) return 0
    let count = 0
    for (const user of users) {
      if (user.is_test_user) count += 1
    }
    return count
  }, [users])

  return (
    <Paper withBorder p="md" radius="md">
      <Stack gap="md">
        <Title order={4}>Registered Users</Title>

        <Group justify="space-between" align="flex-end">
          <TextInput
            placeholder="Search by email..."
            leftSection={<IconSearch size={16} />}
            value={search}
            onChange={(e) => setSearch(e.currentTarget.value)}
            aria-label="Search users by email"
            style={{ flex: 1 }}
          />
          {testUserCount > 0 && (
            <Switch
              label={`Show test users (${testUserCount})`}
              checked={showTestUsers}
              onChange={(e) => setShowTestUsers(e.currentTarget.checked)}
              aria-label="Show test users"
            />
          )}
        </Group>

        {isLoading ? (
          <TableSkeleton />
        ) : (
          <ScrollArea>
            <Table striped highlightOnHover>
              <Table.Thead>
                <Table.Tr>
                  <Table.Th>Email</Table.Th>
                  <Table.Th>Role</Table.Th>
                  <Table.Th ta="center">Verified</Table.Th>
                  <Table.Th ta="center">Brokers</Table.Th>
                  <Table.Th ta="center">Trades</Table.Th>
                  <Table.Th ta="center">Dividends</Table.Th>
                  <Table.Th>Last Login</Table.Th>
                  <Table.Th>Last IP</Table.Th>
                  <Table.Th>Registered</Table.Th>
                </Table.Tr>
              </Table.Thead>
              <Table.Tbody>
                {filteredUsers.length === 0 ? (
                  <Table.Tr>
                    <Table.Td colSpan={9}>
                      <Text c="dimmed" ta="center" py="md">
                        {search ? 'No users found matching search' : 'No users registered'}
                      </Text>
                    </Table.Td>
                  </Table.Tr>
                ) : (
                  filteredUsers.map((user) => (
                    <Table.Tr key={user.id}>
                      <Table.Td>
                        <Text size="sm">{user.email}</Text>
                      </Table.Td>
                      <Table.Td>
                        {user.is_admin ? (
                          <Badge color="violet" size="sm" variant="light">
                            Admin
                          </Badge>
                        ) : user.is_test_user ? (
                          <Badge color="orange" size="sm" variant="light">
                            Test
                          </Badge>
                        ) : (
                          <Badge color="gray" size="sm" variant="light">
                            User
                          </Badge>
                        )}
                      </Table.Td>
                      <Table.Td ta="center">
                        {user.email_verified ? (
                          <Tooltip label="Email verified">
                            <span
                              tabIndex={0}
                              aria-label="Email verified"
                              style={{ display: 'inline-flex', alignItems: 'center' }}
                            >
                              <IconCheck
                                size={16}
                                color="var(--mantine-color-green-6)"
                                aria-hidden="true"
                              />
                            </span>
                          </Tooltip>
                        ) : (
                          <Tooltip label="Email not verified">
                            <span
                              tabIndex={0}
                              aria-label="Email not verified"
                              style={{ display: 'inline-flex', alignItems: 'center' }}
                            >
                              <IconX
                                size={16}
                                color="var(--mantine-color-red-6)"
                                aria-hidden="true"
                              />
                            </span>
                          </Tooltip>
                        )}
                      </Table.Td>
                      <Table.Td ta="center">
                        <Text size="sm">{user.broker_count}</Text>
                      </Table.Td>
                      <Table.Td ta="center">
                        <Text size="sm">{user.trade_count}</Text>
                      </Table.Td>
                      <Table.Td ta="center">
                        <Text size="sm">{user.dividend_count}</Text>
                      </Table.Td>
                      <Table.Td>
                        <Text size="sm" c="dimmed">
                          {user.last_login_at ? dayjs(user.last_login_at).fromNow() : '—'}
                        </Text>
                      </Table.Td>
                      <Table.Td>
                        <Text size="sm" c="dimmed" ff="monospace">
                          {user.last_login_ip ?? '—'}
                        </Text>
                      </Table.Td>
                      <Table.Td>
                        <Text size="sm" c="dimmed">
                          {dayjs(user.created_at).fromNow()}
                        </Text>
                      </Table.Td>
                    </Table.Tr>
                  ))
                )}
              </Table.Tbody>
            </Table>
          </ScrollArea>
        )}

        {users && (
          <Text size="sm" c="dimmed">
            Showing {filteredUsers.length} of {users.length} users
          </Text>
        )}
      </Stack>
    </Paper>
  )
}
