/**
 * useCurrentUser Hook
 * 
 * Helper hook to safely get the current authenticated user ID.
 * Used to ensure all query keys include user context for security.
 */

import { useAuth } from '@/contexts/AuthContext';

/**
 * Get the current authenticated user ID
 * 
 * @returns User ID string or null if not authenticated
 * @throws Error if hook is used outside AuthProvider
 * 
 * @example
 * const userId = useCurrentUser();
 * if (userId) {
 *   // Use userId in query keys
 *   queryKey: queryKeys.trades.list(userId, filters)
 * }
 */
export function useCurrentUser(): string | null {
  const { user } = useAuth();
  return user?.id || null;
}

/**
 * Get the current authenticated user ID with requirement check
 * 
 * Waits for authentication to load before checking if user exists.
 * This prevents errors during page refresh when auth state is loading.
 * 
 * @returns User ID string or null if still loading/not authenticated
 * 
 * @example
 * const userId = useRequiredUser();
 * if (userId) {
 *   // Safe to use userId in query keys
 *   queryKey: queryKeys.trades.list(userId, filters)
 * }
 */
export function useRequiredUser(): string | null {
  const { user, isLoading } = useAuth();
  
  // Return null if still loading - let queries handle the enabled state
  if (isLoading) {
    return null;
  }
  
  // Return user ID if authenticated, null if not
  return user?.id || null;
}