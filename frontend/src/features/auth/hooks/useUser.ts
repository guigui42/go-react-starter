import { useQuery } from '@tanstack/react-query'
import { apiRequest } from '@/lib/api'
import { queryKeys } from '@/lib/queryKeys'

export interface User {
  id: string
  email: string
  is_admin: boolean
  email_verified: boolean
  created_at: string
}

/**
 * Fetches the current authenticated user (if any).
 * Caches the result under the auth.user query key.
 * 
 * With httpOnly cookies, authentication is handled automatically by the browser.
 * The query will return the user data if authenticated, or fail with 401 if not.
 */
export function useUser() {
  return useQuery({
    queryKey: queryKeys.auth.user(),
    queryFn: () => apiRequest<User>('/auth/me'),
    // Fresh data immediately for auth state
    staleTime: 0,
    // Keep cache for a short time to prevent unnecessary refetches
    gcTime: 1000 * 60, // 1 minute
    // Always enabled - the cookie will be sent automatically
    // If not authenticated, API returns 401 and query fails gracefully
    enabled: true,
    // Always refetch on mount to ensure fresh auth state
    refetchOnMount: true,
    // Refetch when window regains focus to catch session changes
    refetchOnWindowFocus: true,
    // Retry logic for auth queries
    retry: (failureCount, error: unknown) => {
      // Don't retry on 401 (unauthorized) responses
      if (typeof error === 'object' && error !== null) {
        const status = (error as { status?: number }).status
        if (status === 401) {
          // User is not authenticated - this is expected when logged out
          return false
        }
      }
      return failureCount < 2
    },
    // Network-first strategy for auth queries
    networkMode: 'always',
  })
}
