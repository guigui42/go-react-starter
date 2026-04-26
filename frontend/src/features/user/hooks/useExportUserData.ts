import { exportUserData } from '@/features/user/api/export'
import { useMutation } from '@tanstack/react-query'

/**
 * Hook to trigger user data export (GDPR compliance)
 * Downloads all user data as a JSON file
 */
export function useExportUserData() {
  return useMutation({
    mutationFn: exportUserData,
  })
}
