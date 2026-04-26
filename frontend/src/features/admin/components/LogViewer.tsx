/**
 * Log Viewer Component
 *
 * Displays backend logs from the ring buffer with filtering, search, and export capabilities.
 * Admin-only component for debugging and monitoring.
 */

import { useAdminLogs, useClearAdminLogs } from '@/features/admin/hooks'
import type { LogEntry, LogLevel } from '@/features/admin/types'
import {
  ActionIcon,
  Badge,
  Box,
  Button,
  Checkbox,
  Code,
  Collapse,
  Group,
  Paper,
  ScrollArea,
  Skeleton,
  Stack,
  Switch,
  Text,
  TextInput,
  Title,
  Tooltip,
} from '@mantine/core'
import { notifications } from '@mantine/notifications'
import {
  IconChevronDown,
  IconChevronUp,
  IconCopy,
  IconDownload,
  IconSearch,
  IconTrash,
} from '@tabler/icons-react'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'

/** Height of the log viewer scroll area in pixels */
const LOG_VIEWER_HEIGHT = 400

const LOG_LEVEL_COLORS: Record<string, string> = {
  debug: 'gray',
  info: 'blue',
  warn: 'orange',
  error: 'red',
}

function LogLevelBadge({ level }: { level: string }) {
  return (
    <Badge color={LOG_LEVEL_COLORS[level] || 'gray'} size="sm" variant="light">
      {level.toUpperCase()}
    </Badge>
  )
}

function formatTimestamp(timestamp: string): string {
  try {
    const date = new Date(timestamp)
    return date.toLocaleTimeString('en-US', {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      hour12: false,
    })
  } catch {
    return timestamp
  }
}

function formatLogEntry(entry: LogEntry): string {
  const timestamp = formatTimestamp(entry.timestamp)
  const level = entry.level.toUpperCase().padEnd(5)
  let line = `[${timestamp}] ${level} ${entry.message}`
  
  if (entry.fields && Object.keys(entry.fields).length > 0) {
    const fieldsStr = Object.entries(entry.fields)
      .map(([k, v]) => `${k}=${typeof v === 'string' ? v : JSON.stringify(v)}`)
      .join(' ')
    line += ` | ${fieldsStr}`
  }
  
  return line
}

interface LogEntryRowProps {
  entry: LogEntry
}

function LogEntryRow({ entry }: LogEntryRowProps) {
  const hasFields = entry.fields && Object.keys(entry.fields).length > 0
  
  return (
    <Box
      py={4}
      px={8}
      style={{
        borderBottom: '1px solid var(--mantine-color-dark-4)',
        fontFamily: 'monospace',
        fontSize: '12px',
      }}
    >
      <Group gap="xs" wrap="nowrap">
        <Text size="xs" c="dimmed" style={{ minWidth: 65 }}>
          {formatTimestamp(entry.timestamp)}
        </Text>
        <LogLevelBadge level={entry.level} />
        <Text size="xs" style={{ wordBreak: 'break-word' }}>
          {entry.message}
        </Text>
      </Group>
      {hasFields && (
        <Code block mt={4} style={{ fontSize: '11px' }}>
          {JSON.stringify(entry.fields, null, 2)}
        </Code>
      )}
    </Box>
  )
}

function LogViewerSkeleton() {
  return (
    <Stack gap="sm">
      {Array.from({ length: 8 }).map((_, i) => (
        <Skeleton key={i} height={30} radius="sm" />
      ))}
    </Stack>
  )
}

export function LogViewer() {
  const { data, isLoading, error } = useAdminLogs()
  const clearLogs = useClearAdminLogs()
  
  const [isExpanded, setIsExpanded] = useState(true)
  const [search, setSearch] = useState('')
  const [autoScroll, setAutoScroll] = useState(true)
  const [enabledLevels, setEnabledLevels] = useState<Set<LogLevel>>(
    new Set(['debug', 'info', 'warn', 'error'])
  )
  
  const scrollAreaRef = useRef<HTMLDivElement>(null)
  const viewportRef = useRef<HTMLDivElement>(null)
  
  // Extract entries for memoization dependency
  const entries = data?.entries
  
  // Filter entries based on search and level
  const filteredEntries = useMemo(() => {
    if (!entries) return []
    
    return entries.filter((entry) => {
      // Level filter
      if (!enabledLevels.has(entry.level as LogLevel)) return false
      
      // Search filter (case-insensitive)
      if (search) {
        const searchLower = search.toLowerCase()
        const matchesMessage = entry.message.toLowerCase().includes(searchLower)
        const matchesFields = entry.fields 
          ? JSON.stringify(entry.fields).toLowerCase().includes(searchLower)
          : false
        if (!matchesMessage && !matchesFields) return false
      }
      
      return true
    })
  }, [entries, enabledLevels, search])
  
  // Auto-scroll to bottom when new entries arrive
  useEffect(() => {
    if (autoScroll && viewportRef.current) {
      viewportRef.current.scrollTo({
        top: viewportRef.current.scrollHeight,
        behavior: 'smooth',
      })
    }
  }, [filteredEntries.length, autoScroll])
  
  const toggleLevel = useCallback((level: LogLevel) => {
    setEnabledLevels((prev) => {
      const next = new Set(prev)
      if (next.has(level)) {
        next.delete(level)
      } else {
        next.add(level)
      }
      return next
    })
  }, [])
  
  const handleCopyAll = useCallback(() => {
    const text = filteredEntries.map(formatLogEntry).join('\n')
    navigator.clipboard.writeText(text).then(
      () => {
        notifications.show({
          title: 'Copied',
          message: `${filteredEntries.length} log entries copied to clipboard`,
          color: 'green',
        })
      },
      () => {
        notifications.show({
          title: 'Copy Failed',
          message: 'Failed to copy logs to clipboard',
          color: 'red',
        })
      }
    )
  }, [filteredEntries])
  
  const handleDownload = useCallback(() => {
    const text = filteredEntries.map(formatLogEntry).join('\n')
    const blob = new Blob([text], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-')
    link.download = `Go React Starter-logs-${timestamp}.log`
    link.style.display = 'none'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
    
    notifications.show({
      title: 'Downloaded',
      message: `${filteredEntries.length} log entries saved to file`,
      color: 'green',
    })
  }, [filteredEntries])
  
  const handleClearLogs = useCallback(async () => {
    try {
      await clearLogs.mutateAsync()
      notifications.show({
        title: 'Logs Cleared',
        message: 'Log buffer has been cleared',
        color: 'green',
      })
    } catch {
      notifications.show({
        title: 'Clear Failed',
        message: 'Failed to clear log buffer',
        color: 'red',
      })
    }
  }, [clearLogs])
  
  if (error) {
    return (
      <Paper withBorder p="md" radius="md">
        <Text c="red">Failed to load logs: {error.message}</Text>
      </Paper>
    )
  }
  
  return (
    <Paper withBorder p="md" radius="md">
      <Stack gap="md">
        {/* Header */}
        <Group justify="space-between">
          <Group gap="xs">
            <Title order={4}>Backend Logs</Title>
            {data && (
              <Text size="sm" c="dimmed">
                ({data.total} / {data.capacity})
              </Text>
            )}
          </Group>
          <ActionIcon
            variant="subtle"
            onClick={() => setIsExpanded(!isExpanded)}
            aria-label={isExpanded ? 'Collapse logs' : 'Expand logs'}
          >
            {isExpanded ? <IconChevronUp size={18} /> : <IconChevronDown size={18} />}
          </ActionIcon>
        </Group>
        
        <Collapse expanded={isExpanded}>
          <Stack gap="md">
            {/* Controls */}
            <Group justify="space-between" wrap="wrap" gap="sm">
              {/* Level Filters */}
              <Group gap="xs">
                {(['debug', 'info', 'warn', 'error'] as const).map((level) => (
                  <Checkbox
                    key={level}
                    label={
                      <Badge color={LOG_LEVEL_COLORS[level]} size="sm" variant="light">
                        {level.toUpperCase()}
                      </Badge>
                    }
                    checked={enabledLevels.has(level)}
                    onChange={() => toggleLevel(level)}
                    aria-label={`Filter ${level} logs`}
                  />
                ))}
              </Group>
              
              {/* Auto-scroll Toggle */}
              <Switch
                label="Auto-scroll"
                checked={autoScroll}
                onChange={(e) => setAutoScroll(e.currentTarget.checked)}
                size="sm"
              />
            </Group>
            
            {/* Search and Actions */}
            <Group gap="sm">
              <TextInput
                placeholder="Search logs..."
                leftSection={<IconSearch size={16} />}
                value={search}
                onChange={(e) => setSearch(e.currentTarget.value)}
                style={{ flex: 1 }}
                aria-label="Search logs"
              />
              <Tooltip label="Copy all visible logs">
                <ActionIcon
                  variant="light"
                  color="blue"
                  onClick={handleCopyAll}
                  disabled={filteredEntries.length === 0}
                  aria-label="Copy all logs to clipboard"
                >
                  <IconCopy size={16} />
                </ActionIcon>
              </Tooltip>
              <Tooltip label="Download as .log file">
                <ActionIcon
                  variant="light"
                  color="green"
                  onClick={handleDownload}
                  disabled={filteredEntries.length === 0}
                  aria-label="Download logs as file"
                >
                  <IconDownload size={16} />
                </ActionIcon>
              </Tooltip>
              <Tooltip label="Clear log buffer">
                <Button
                  variant="light"
                  color="red"
                  size="xs"
                  leftSection={<IconTrash size={14} />}
                  onClick={handleClearLogs}
                  loading={clearLogs.isPending}
                  aria-label="Clear log buffer"
                >
                  Clear
                </Button>
              </Tooltip>
            </Group>
            
            {/* Log Entries */}
            {isLoading ? (
              <LogViewerSkeleton />
            ) : (
              <ScrollArea
                ref={scrollAreaRef}
                h={LOG_VIEWER_HEIGHT}
                viewportRef={viewportRef}
                type="auto"
                offsetScrollbars
                style={{
                  backgroundColor: 'var(--mantine-color-dark-7)',
                  borderRadius: 'var(--mantine-radius-sm)',
                }}
              >
                {filteredEntries.length === 0 ? (
                  <Box p="md">
                    <Text c="dimmed" ta="center">
                      {data?.entries.length === 0
                        ? 'No log entries'
                        : 'No logs match the current filters'}
                    </Text>
                  </Box>
                ) : (
                  filteredEntries.map((entry, idx) => (
                    <LogEntryRow key={`${entry.timestamp}-${entry.level}-${idx}`} entry={entry} />
                  ))
                )}
              </ScrollArea>
            )}
            
            {/* Footer */}
            {data && (
              <Text size="sm" c="dimmed">
                Showing {filteredEntries.length} of {data.total} log entries
              </Text>
            )}
          </Stack>
        </Collapse>
      </Stack>
    </Paper>
  )
}
