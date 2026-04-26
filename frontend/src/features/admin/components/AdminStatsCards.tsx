/**
 * Admin Stats Cards Component
 *
 * Displays system overview statistics as stat cards
 */

import type { AdminStats } from '@/features/admin/types'
import { Group, Paper, SimpleGrid, Skeleton, Stack, Text, ThemeIcon } from '@mantine/core'
import { IconFileText, IconUsers } from '@tabler/icons-react'

interface AdminStatsCardsProps {
  stats?: AdminStats
  isLoading: boolean
}

interface StatCardProps {
  label: string
  value: string | number
  icon: React.ReactNode
  color: string
  subtext?: React.ReactNode
}

function StatCard({ label, value, icon, color, subtext }: StatCardProps) {
  return (
    <Paper withBorder p="md" radius="md" shadow="sm">
      <Group justify="space-between" wrap="nowrap">
        <Stack gap={4} style={{ flex: 1 }}>
          <Text size="xs" c="dimmed" tt="uppercase" fw={600}>
            {label}
          </Text>
          <Text size="xl" fw={700}>
            {value}
          </Text>
          {subtext && (
            <Text size="xs" c="dimmed">
              {subtext}
            </Text>
          )}
        </Stack>
        <ThemeIcon color={color} variant="light" size="xl" radius="md">
          {icon}
        </ThemeIcon>
      </Group>
    </Paper>
  )
}

function StatCardSkeleton() {
  return <Skeleton height={100} radius="md" />
}

export function AdminStatsCards({ stats, isLoading }: AdminStatsCardsProps) {
  if (isLoading) {
    return (
      <SimpleGrid cols={{ base: 2, md: 3 }} spacing="md">
        <StatCardSkeleton />
      </SimpleGrid>
    )
  }

  if (!stats) {
    return null
  }

  return (
    <SimpleGrid cols={{ base: 2, md: 3 }} spacing="md">
      <StatCard
        label="Total Users"
        value={stats.total_users}
        icon={<IconUsers size={24} />}
        color="blue"
      />
      <StatCard
        label="Total Notes"
        value={stats.total_notes}
        icon={<IconFileText size={24} />}
        color="teal"
      />
    </SimpleGrid>
  )
}
