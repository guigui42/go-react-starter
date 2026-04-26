import { apiRequest, API_PREFIX } from '@/lib/api'
import type { UserPreferences, UpdateUserPreferencesRequest } from '@/features/user/types/preferences'

/**
 * Get current user's preferences
 */
export async function getUserPreferences(): Promise<UserPreferences> {
  return apiRequest<UserPreferences>(`${API_PREFIX}/preferences`)
}

/**
 * Update current user's preferences
 */
export async function updateUserPreferences(
  data: UpdateUserPreferencesRequest
): Promise<UserPreferences> {
  return apiRequest<UserPreferences>(`${API_PREFIX}/preferences`, {
    method: 'PUT',
    body: JSON.stringify(data),
  })
}
