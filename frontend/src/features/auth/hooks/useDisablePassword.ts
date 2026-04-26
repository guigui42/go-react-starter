import { useMutation, useQueryClient } from '@tanstack/react-query';
import { queryKeys } from '@/lib/queryKeys';
import { apiRequest } from '@/lib/api';

export interface DisablePasswordResponse {
  password_login_enabled: boolean;
  backup_codes: string[];
  message: string;
}

/**
 * Hook to disable password login (after user has registered passkey)
 */
export function useDisablePassword() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (confirmed: boolean) =>
      apiRequest<DisablePasswordResponse>('/auth/migration/disable-password', {
        method: 'POST',
        body: JSON.stringify({ confirmed }),
      }),
    onSuccess: () => {
      // Invalidate migration status
      queryClient.invalidateQueries({ queryKey: queryKeys.auth.migrationStatus() });
      queryClient.invalidateQueries({ queryKey: queryKeys.auth.user() });
    },
  });
}
