/**
 * Admin Feature Module
 *
 * Provides admin dashboard functionality for user management and system monitoring.
 * Access is restricted to designated admin users.
 */

export { AdminDashboard, AdminStatsCards, UserList, LogViewer, MigrationHistory } from './components'
export { useAdminStats, useAdminUsers, useAdminLogs, useClearAdminLogs, useMigrationStatus } from './hooks'
export type { AdminStats, AdminUser, LogEntry, LogsResponse, LogLevel, MigrationStatus } from './types'
