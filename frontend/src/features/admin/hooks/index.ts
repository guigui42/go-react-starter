import type { AdminStats, AdminUser, AuditLogQuery, AuditLogResponse, EmailConfigStatus, LogsResponse, MigrationStatus, TestDataCountResponse, TestDataDeleteResponse, TestDataGenerateRequest, TestDataGenerateResponse, TestDataProgress, TestEmailRequest, TestEmailResponse } from '@/features/admin/types'
import { API_PREFIX, apiRequest } from '@/lib/api'
import { queryKeys } from '@/lib/queryKeys'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

interface AdminQueryOptions {
  /** Whether to enable the query and polling. Defaults to true. */
  enabled?: boolean
}

/**
 * Hook to fetch admin dashboard statistics
 */
export function useAdminStats(options: AdminQueryOptions = {}) {
  const { enabled = true } = options
  return useQuery({
    queryKey: queryKeys.admin.stats(),
    queryFn: () => apiRequest<AdminStats>(`${API_PREFIX}/admin/stats`),
    staleTime: 1000 * 30, // 30 seconds
    refetchInterval: enabled ? 1000 * 5 : false, // Poll every 5 seconds when enabled
    enabled,
  })
}

/**
 * Hook to fetch all users for admin display
 */
export function useAdminUsers(options: AdminQueryOptions = {}) {
  const { enabled = true } = options
  return useQuery({
    queryKey: queryKeys.admin.users(),
    queryFn: () => apiRequest<AdminUser[]>(`${API_PREFIX}/admin/users`),
    staleTime: 1000 * 30, // 30 seconds
    refetchInterval: enabled ? 1000 * 30 : false, // Poll every 30 seconds when enabled
    enabled,
  })
}

/** Polling interval for admin logs in milliseconds */
const ADMIN_LOGS_POLL_INTERVAL = 5000

/**
 * Hook to fetch backend logs from the admin endpoint
 * Polls every 5 seconds (consistent with other admin hooks)
 */
export function useAdminLogs() {
  return useQuery({
    queryKey: queryKeys.admin.logs(),
    queryFn: () => apiRequest<LogsResponse>(`${API_PREFIX}/admin/logs`),
    staleTime: ADMIN_LOGS_POLL_INTERVAL,
    refetchInterval: ADMIN_LOGS_POLL_INTERVAL,
  })
}

/**
 * Hook to clear the backend log buffer
 */
export function useClearAdminLogs() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: () => 
      apiRequest<{ message: string }>(`${API_PREFIX}/admin/logs`, {
        method: 'DELETE',
      }),
    onSuccess: () => {
      // Invalidate logs query to refresh the display
      queryClient.invalidateQueries({ queryKey: queryKeys.admin.logs() })
    },
  })
}

/**
 * Hook to fetch email configuration status
 */
export function useEmailConfig(options: AdminQueryOptions = {}) {
  const { enabled = true } = options
  return useQuery({
    queryKey: queryKeys.admin.emailConfig(),
    queryFn: () => apiRequest<EmailConfigStatus>(`${API_PREFIX}/admin/email/config`),
    staleTime: 1000 * 60 * 5, // 5 minutes - email config rarely changes
    enabled,
  })
}

/**
 * Hook to send a test email
 */
export function useSendTestEmail() {
  return useMutation({
    mutationFn: (request: TestEmailRequest) => 
      apiRequest<TestEmailResponse>(`${API_PREFIX}/admin/email/test`, {
        method: 'POST',
        body: JSON.stringify(request),
      }),
  })
}

/** Polling interval for audit logs in milliseconds */
const AUDIT_LOGS_POLL_INTERVAL = 10000

/**
 * Hook to fetch audit logs with filtering and pagination
 */
export function useAuditLogs(query: AuditLogQuery = {}, options: AdminQueryOptions = {}) {
  const { enabled = true } = options

  const params = new URLSearchParams()
  if (query.event_type) params.set('event_type', query.event_type)
  if (query.status) params.set('status', query.status)
  if (query.actor_id) params.set('actor_id', query.actor_id)
  if (query.from) params.set('from', query.from)
  if (query.to) params.set('to', query.to)
  if (query.page) params.set('page', String(query.page))
  if (query.page_size) params.set('page_size', String(query.page_size))

  const queryString = params.toString()
  const url = `${API_PREFIX}/admin/audit-logs${queryString ? `?${queryString}` : ''}`

  return useQuery({
    queryKey: [...queryKeys.admin.auditLogs(), query],
    queryFn: () => apiRequest<AuditLogResponse>(url),
    staleTime: AUDIT_LOGS_POLL_INTERVAL,
    refetchInterval: enabled ? AUDIT_LOGS_POLL_INTERVAL : false,
    enabled,
  })
}

/**
 * Hook to fetch database migration status
 */
export function useMigrationStatus(options: AdminQueryOptions = {}) {
  const { enabled = true } = options
  return useQuery({
    queryKey: queryKeys.admin.migrations(),
    queryFn: () => apiRequest<MigrationStatus[]>(`${API_PREFIX}/admin/migrations`),
    staleTime: 1000 * 60 * 5, // 5 minutes — migrations rarely change at runtime
    enabled,
  })
}

/**
 * Hook to fetch current test user count
 */
export function useTestDataCount(options: AdminQueryOptions = {}) {
  const { enabled = true } = options
  return useQuery({
    queryKey: queryKeys.admin.testDataCount(),
    queryFn: () => apiRequest<TestDataCountResponse>(`${API_PREFIX}/admin/test-data/count`),
    staleTime: 1000 * 10,
    enabled,
  })
}

/**
 * Hook to poll test data generation progress
 */
export function useTestDataProgress(jobId: string | null) {
  return useQuery({
    queryKey: queryKeys.admin.testDataProgress(jobId ?? ''),
    queryFn: () => apiRequest<TestDataProgress>(`${API_PREFIX}/admin/test-data/progress/${jobId}`),
    enabled: !!jobId,
    refetchInterval: (query) => {
      const data = query.state.data
      if (data?.done) return false
      return 1000 // Poll every second while in progress
    },
  })
}

/**
 * Hook to generate test users
 */
export function useGenerateTestData() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (request: TestDataGenerateRequest) =>
      apiRequest<TestDataGenerateResponse>(`${API_PREFIX}/admin/test-data/generate`, {
        method: 'POST',
        body: JSON.stringify(request),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.admin.testDataCount() })
      queryClient.invalidateQueries({ queryKey: queryKeys.admin.stats() })
    },
  })
}

/**
 * Hook to delete all test users
 */
export function useDeleteTestData() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: () =>
      apiRequest<TestDataDeleteResponse>(`${API_PREFIX}/admin/test-data`, {
        method: 'DELETE',
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.admin.testDataCount() })
      queryClient.invalidateQueries({ queryKey: queryKeys.admin.stats() })
    },
  })
}
