import { useCurrentUser } from '@/features/auth/hooks/useCurrentUser'
import { getUserPreferences, updateUserPreferences } from '@/features/user/api/preferences'
import type { UpdateUserPreferencesRequest, UserPreferences } from '@/features/user/types/preferences'
import { queryKeys } from '@/lib/queryKeys'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

/**
 * Hook to fetch current user's preferences
 */
export function useUserPreferences() {
  const userId = useCurrentUser()
  
  return useQuery({
    queryKey: queryKeys.preferences.user(userId || ''),
    queryFn: getUserPreferences,
    enabled: !!userId,
    staleTime: 5 * 60 * 1000, // 5 minutes
  })
}

/**
 * Hook to update user preferences
 */
export function useUpdateUserPreferences() {
  const queryClient = useQueryClient()
  const userId = useCurrentUser()
  
  return useMutation({
    mutationFn: (data: UpdateUserPreferencesRequest) => updateUserPreferences(data),
    onSuccess: (updatedPreferences: UserPreferences) => {
      if (userId) {
        queryClient.setQueryData(
          queryKeys.preferences.user(userId),
          updatedPreferences
        )
      }
    },
  })
}
