import { apiDownload, API_PREFIX } from '@/lib/api'

/**
 * Export user data for GDPR compliance
 * Downloads all user data as a JSON file
 */
export async function exportUserData(): Promise<void> {
  const date = new Date().toISOString().split('T')[0]
  const filename = `data-export-${date}.json`
  return apiDownload(`${API_PREFIX}/user/export`, filename)
}
