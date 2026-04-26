/**
 * Admin Dashboard Types
 * Types for admin dashboard components and API responses
 */

export interface AdminStats {
  total_users: number
}

export interface AdminUser {
  id: string
  email: string
  is_admin: boolean
  is_test_user: boolean
  email_verified: boolean
  created_at: string
  last_login_at?: string
  last_login_ip?: string
}

/** Log entry from the backend ring buffer */
export interface LogEntry {
  timestamp: string
  level: string
  message: string
  fields?: Record<string, unknown>
}

/** Response from GET /api/v1/admin/logs */
export interface LogsResponse {
  entries: LogEntry[]
  total: number
  capacity: number
}

/** Valid log levels for filtering */
export type LogLevel = 'debug' | 'info' | 'warn' | 'error'

/** Email configuration status from backend */
export interface EmailConfigStatus {
  configured: boolean
  provider: string
  from_address: string
  from_name: string
}

/** Request to send a test email */
export interface TestEmailRequest {
  to: string
  subject?: string
  message?: string
}

/** Response from sending a test email */
export interface TestEmailResponse {
  success: boolean
  message: string
  recipient: string
}

/** Audit log entry from the database */
export interface AuditLog {
  id: string
  event_type: string
  actor_id: string | null
  target_id: string | null
  action: string
  status: string
  ip_address: string | null
  user_agent: string | null
  metadata: Record<string, unknown> | null
  created_at: string
}

/** Query parameters for audit log filtering */
export interface AuditLogQuery {
  event_type?: string
  status?: string
  actor_id?: string
  from?: string
  to?: string
  page?: number
  page_size?: number
}

/** Paginated audit log response */
export interface AuditLogResponse {
  logs: AuditLog[]
  total: number
  page: number
  page_size: number
  total_pages: number
  actors: Record<string, string>
}

/** Database migration status */
export interface MigrationStatus {
  version: string
  name: string
  applied: boolean
  failed: boolean
  applied_at?: string
  error_message?: string
}

/** Request to generate test users */
export interface TestDataGenerateRequest {
  count: number
}

/** Response from test data generation */
export interface TestDataGenerateResponse {
  job_id?: string
  message: string
  count: number
}

/** Progress of an async test data generation job */
export interface TestDataProgress {
  job_id: string
  current: number
  total: number
  done: boolean
  error?: string
}

/** Response from test data count endpoint */
export interface TestDataCountResponse {
  count: number
}

/** Response from test data delete endpoint */
export interface TestDataDeleteResponse {
  message: string
  deleted_count: number
}
