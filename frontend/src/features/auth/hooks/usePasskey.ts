/**
 * Passkey Authentication Hooks
 * 
 * TanStack Query hooks for WebAuthn/Passkey authentication flows
 */

import { passkeyService } from '@/features/auth/services'
import type { PasskeyCredential } from '@/features/auth/services/passkeyService.types'
import { notifyLogin } from '@/lib/authEvents'
import { queryKeys } from '@/lib/queryKeys'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

/**
 * Hook for passkey registration flow
 * Combines startRegistration + finishRegistration
 */
export function usePasskeyRegistration() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (friendlyName?: string) => {
      return await passkeyService.register(friendlyName)
    },
    onSuccess: async () => {
      // User is already authenticated (token was set during account registration)
      // Just invalidate queries to refresh passkey list
      await queryClient.invalidateQueries({ queryKey: queryKeys.auth.passkeys() })

      // Use window.location for a full page reload to ensure router context is updated
      // This avoids race conditions with the router's beforeLoad guards
      // Redirect to broker creation for onboarding flow
      window.location.href = '/'
    },
    onError: () => {
      // Error is already a PasskeyError with user-friendly message
    },
  })
}

/**
 * Hook for passkey authentication flow
 * Combines startAuthentication + finishAuthentication
 */
export function usePasskeyAuthentication() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (email: string) => {
      return await passkeyService.authenticate(email)
    },
    onSuccess: async () => {
      // Token is automatically set in httpOnly cookie by backend
      // Notify auth state change
      notifyLogin()

      // Invalidate auth queries and wait for refetch
      await queryClient.invalidateQueries({ queryKey: queryKeys.auth.user() })

      // Use window.location for a full page reload to ensure router context is updated
      // This avoids race conditions with the router's beforeLoad guards
      window.location.href = '/'
    },
    onError: () => {
      // Error is already a PasskeyError with user-friendly message
    },
  })
}

/**
 * Hook to list all passkey credentials for the current user
 */
export function usePasskeys() {
  return useQuery({
    queryKey: queryKeys.auth.passkeysList(),
    queryFn: () => passkeyService.listCredentials(),
    staleTime: 5 * 60 * 1000, // 5 minutes
    // Query is always enabled - will return 401 if not authenticated
  })
}

/**
 * Hook to delete a passkey credential
 */
export function useDeletePasskey() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (credentialId: string) => {
      return passkeyService.deleteCredential(credentialId)
    },
    onSuccess: () => {
      // Invalidate passkey queries to refetch list
      queryClient.invalidateQueries({ queryKey: queryKeys.auth.passkeys() })
    },
    onError: () => {
      // Error handled by TanStack Query
    },
  })
}

/**
 * Hook to update a passkey credential (rename)
 */
export function useUpdatePasskey() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ 
      credentialId, 
      friendlyName 
    }: { 
      credentialId: string
      friendlyName: string 
    }) => {
      return passkeyService.updateCredential(credentialId, { friendlyName })
    },
    onMutate: async ({ credentialId, friendlyName }) => {
      // Cancel outgoing refetches
      await queryClient.cancelQueries({ queryKey: queryKeys.auth.passkeysList() })

      // Snapshot previous value
      const previousPasskeys = queryClient.getQueryData<PasskeyCredential[]>(
        queryKeys.auth.passkeysList()
      )

      // Optimistically update cache
      if (previousPasskeys) {
        queryClient.setQueryData<PasskeyCredential[]>(
          queryKeys.auth.passkeysList(),
          previousPasskeys.map(pk =>
            pk.id === credentialId
              ? { ...pk, friendly_name: friendlyName }
              : pk
          )
        )
      }

      return { previousPasskeys }
    },
    onError: (error, _variables, context) => {
      // Rollback on error
      if (context?.previousPasskeys) {
        queryClient.setQueryData(
          queryKeys.auth.passkeysList(),
          context.previousPasskeys
        )
      }
      // Error handled by TanStack Query - optimistic update already rolled back above
    },
    onSettled: () => {
      // Refetch to ensure consistency
      queryClient.invalidateQueries({ queryKey: queryKeys.auth.passkeys() })
    },
  })
}

/**
 * Hook to get user migration status (password vs passkey)
 */
export function useMigrationStatus() {
  return useQuery({
    queryKey: queryKeys.auth.migrationStatus(),
    queryFn: async () => {
      // TODO: Implement API call to /auth/migration/status
      // For now, return mock data
      return {
        has_password: true,
        has_passkey: false,
        password_login_enabled: true,
        passkey_login_enabled: false,
        can_disable_password: false,
      }
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
    // Query is always enabled - will return 401 if not authenticated
  })
}

/**
 * Hook to disable password login (after migrating to passkey)
 */
export function useDisablePassword() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (confirmed: boolean) => {
      if (!confirmed) {
        throw new Error('User must confirm password disable action')
      }

      // TODO: Implement API call to /auth/migration/disable-password
      return { 
        password_login_enabled: false,
        backup_codes: [] as string[],
        message: 'Password login disabled'
      }
    },
    onSuccess: () => {
      // Invalidate migration status
      queryClient.invalidateQueries({ queryKey: queryKeys.auth.migration() })
    },
    onError: () => {
      // Error handled by TanStack Query
    },
  })
}
