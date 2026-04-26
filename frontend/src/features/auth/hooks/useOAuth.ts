import { apiRequest, API_PREFIX } from '@/lib/api';
import { queryKeys } from '@/lib/queryKeys';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

/**
 * OAuth Provider info from the backend
 */
export interface OAuthProviderInfo {
  name: string;
  enabled: boolean;
}

/**
 * OAuth Providers response
 */
export interface OAuthProvidersResponse {
  providers: OAuthProviderInfo[];
}

/**
 * Linked OAuth account info
 */
export interface OAuthAccountInfo {
  id: string;
  provider: string;
  provider_email?: string;
  created_at: string;
}

/**
 * OAuth Accounts response
 */
export interface OAuthAccountsResponse {
  accounts: OAuthAccountInfo[];
}

/**
 * Hook to fetch enabled OAuth providers
 * This is used to show/hide OAuth buttons on the login page
 */
export function useOAuthProviders() {
  return useQuery({
    queryKey: queryKeys.oauth.providers(),
    queryFn: () => apiRequest<OAuthProvidersResponse>('/auth/oauth/providers'),
    staleTime: 5 * 60 * 1000, // Cache for 5 minutes
    retry: false,
  });
}

/**
 * Hook to fetch linked OAuth accounts for the current user
 */
export function useLinkedOAuthAccounts(userId?: string) {
  return useQuery({
    queryKey: queryKeys.oauth.accounts(userId ?? '__unauthenticated__'),
    queryFn: () => apiRequest<OAuthAccountsResponse>(`${API_PREFIX}/auth/oauth/accounts`),
    enabled: !!userId,
    staleTime: 60 * 1000, // Cache for 1 minute
  });
}

/**
 * Hook to unlink an OAuth provider from the current user
 */
export function useUnlinkOAuth(userId?: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (provider: string) =>
      apiRequest(`${API_PREFIX}/auth/oauth/${provider}`, {
        method: 'DELETE',
      }),
    onSuccess: () => {
      if (userId) {
        // Invalidate the specific user's OAuth accounts cache
        queryClient.invalidateQueries({ queryKey: queryKeys.oauth.accounts(userId) });
      } else {
        // Fallback: invalidate all OAuth accounts caches
        queryClient.invalidateQueries({
          predicate: (query) =>
            Array.isArray(query.queryKey) &&
            query.queryKey.length >= 3 &&
            query.queryKey[1] === 'oauth' &&
            query.queryKey[2] === 'accounts',
        });
      }
    },
  });
}

/**
 * Get the OAuth login URL for a provider
 * This redirects the user to the OAuth provider's login page
 */
export function getOAuthLoginUrl(provider: string): string {
  // In development, we need to use the backend URL directly
  const baseUrl = import.meta.env.DEV ? 'http://localhost:8080' : '';
  return `${baseUrl}/auth/oauth/${provider}`;
}

/**
 * Start OAuth login by redirecting to the OAuth provider
 */
export function startOAuthLogin(provider: string): void {
  window.location.href = getOAuthLoginUrl(provider);
}

/**
 * Provider display names and icons
 */
export const OAUTH_PROVIDER_INFO: Record<string, { displayName: string; icon: string }> = {
  github: {
    displayName: 'GitHub',
    icon: 'github',
  },
  google: {
    displayName: 'Google',
    icon: 'google',
  },
  facebook: {
    displayName: 'Facebook',
    icon: 'facebook',
  },
};

/**
 * Get display name for an OAuth provider
 */
export function getProviderDisplayName(provider: string): string {
  return OAUTH_PROVIDER_INFO[provider]?.displayName ?? provider;
}
