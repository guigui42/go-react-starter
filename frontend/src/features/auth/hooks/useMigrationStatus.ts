import { useQuery } from '@tanstack/react-query';
import { queryKeys } from '@/lib/queryKeys';
import { apiRequest } from '@/lib/api';

export interface MigrationStatus {
  has_password: boolean;
  has_passkey: boolean;
  password_login_enabled: boolean;
  passkey_login_enabled: boolean;
  can_disable_password: boolean;
}

/**
 * Hook to fetch user's authentication migration status
 */
export function useMigrationStatus() {
  return useQuery({
    queryKey: queryKeys.auth.migrationStatus(),
    queryFn: () => apiRequest<MigrationStatus>('/auth/migration/status'),
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}
