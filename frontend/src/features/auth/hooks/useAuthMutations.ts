import { apiRequest } from '@/lib/api';
import { notifyLogin, notifyLogout } from '@/lib/authEvents';
import { queryKeys } from '@/lib/queryKeys';
import { useMutation, useQueryClient } from '@tanstack/react-query';

interface User {
  id: string;
  email: string;
  created_at: string;
}

interface LoginCredentials {
  email: string;
  password: string;
}

interface LoginResponse {
  user: User;
}

interface RegisterCredentials {
  email: string;
  password: string;
}

/**
 * Response from the check-methods endpoint
 * Indicates which authentication methods are available for a given email
 */
interface CheckAuthMethodsResponse {
  has_password: boolean;
  has_passkey: boolean;
}

/**
 * Response when email verification is pending
 */
interface VerificationPendingResponse {
  message: string;
  email: string;
}

/**
 * Check authentication methods mutation hook
 * 
 * Checks which authentication methods (password, passkey) are available
 * for a given email address. Used by the email-first login flow to
 * determine which login options to display.
 * 
 * Note: The backend returns the same structure for non-existent users
 * (with has_password=true) to prevent user enumeration attacks.
 * 
 * @example
 * const checkAuth = useCheckAuthMethods();
 * 
 * const handleEmailSubmit = async (email: string) => {
 *   const result = await checkAuth.mutateAsync(email);
 *   if (result.has_passkey) {
 *     // Show passkey authentication
 *   } else if (result.has_password) {
 *     // Show password input
 *   }
 * };
 */
export function useCheckAuthMethods() {
  return useMutation({
    mutationFn: (email: string) =>
      apiRequest<CheckAuthMethodsResponse>('/auth/check-methods', {
        method: 'POST',
        body: JSON.stringify({ email }),
      }),
    // Don't retry on failure - we want immediate feedback
    retry: false,
  });
}

/**
 * Login mutation hook
 * 
 * Authenticates user. JWT token is automatically set in httpOnly cookie by the backend.
 * Caches user data in TanStack Query for efficient access.
 * 
 * Handles EMAIL_NOT_VERIFIED error by throwing with additional data for redirection.
 * 
 * @example
 * const login = useLogin();
 * 
 * const handleLogin = async () => {
 *   try {
 *     await login.mutateAsync({ email, password });
 *     navigate('/');
 *   } catch (error) {
 *     if (error instanceof ApiError && error.code === 'EMAIL_NOT_VERIFIED') {
 *       // Redirect to verification pending page
 *       navigate(`/auth/verify-pending?email=${encodeURIComponent(error.details?.email as string)}`);
 *     }
 *     // Other errors are available in login.error
 *   }
 * };
 */
export function useLogin() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ email, password }: LoginCredentials) =>
      apiRequest<LoginResponse>('/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      }),
    // Prevent multiple simultaneous login attempts
    retry: false,
    onMutate: () => {
      // Clear any stale auth queries to prevent conflicts
      queryClient.removeQueries({ queryKey: queryKeys.auth.user() });
    },
    onSuccess: async (data) => {
      // Token is automatically set in httpOnly cookie by backend
      // Notify auth state change
      notifyLogin();
      
      // IMMEDIATELY cache user data in query client (optimistic update)
      queryClient.setQueryData(queryKeys.auth.user(), data.user);
      
      // Clear all non-auth queries to prevent cross-user contamination
      queryClient.removeQueries({ 
        predicate: (query) => {
          const key = query.queryKey;
          // Keep auth queries but clear everything else
          return !key.includes('auth');
        }
      });
      
      // Force immediate invalidation and refetch to ensure consistency
      await queryClient.invalidateQueries({ queryKey: queryKeys.auth.user() });
    },
  });
}

/**
 * Register mutation hook
 * 
 * Registers new user and automatically logs them in.
 * Backend returns token in httpOnly cookie directly on registration (auto-login).
 * 
 * @example
 * const register = useRegister();
 * 
 * const handleRegister = async () => {
 *   try {
 *     const result = await register.mutateAsync({ email, password });
 *     if ('message' in result && result.message === 'verification_email_sent') {
 *       // Redirect to verification pending page
 *       navigate(`/auth/verify-pending?email=${encodeURIComponent(result.email)}`);
 *     } else {
 *       navigate('/');
 *     }
 *   } catch (error) {
 *     // Error is available in register.error
 *   }
 * };
 */
export function useRegister() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ email, password }: RegisterCredentials) => {
      // Register user - backend returns either:
      // 1. User data with token in httpOnly cookie (auto-login) when email verification is disabled
      // 2. Verification pending response with email when email verification is enabled
      return await apiRequest<LoginResponse | VerificationPendingResponse>('/auth/register', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      });
    },
    onSuccess: async (data) => {
      // Check if this is a verification pending response
      if ('message' in data && data.message === 'verification_email_sent') {
        // Email verification required - no auto-login
        return;
      }

      // Normal registration success - token is automatically set in httpOnly cookie by backend
      // Notify auth state change
      notifyLogin();
      
      // IMMEDIATELY cache user data in query client (optimistic update)
      if ('user' in data) {
        queryClient.setQueryData(queryKeys.auth.user(), data.user);
      }
      
      // Clear only non-auth queries to prevent cross-user contamination
      queryClient.removeQueries({ 
        predicate: (query) => {
          const key = query.queryKey;
          return !key.includes('auth');
        }
      });
      
      // Force immediate refetch of user query in background
      queryClient.refetchQueries({ queryKey: queryKeys.auth.user() });
    },
  });
}

/**
 * Logout mutation hook
 * 
 * SECURITY: Performs comprehensive cleanup to prevent data leakage:
 * - Backend clears authentication cookie
 * - Clears ALL cached query data from memory
 * - Invalidates all active queries
 * 
 * Always clears local data even if API call fails to ensure security.
 * 
 * @example
 * const logout = useLogout();
 * 
 * const handleLogout = async () => {
 *   await logout.mutateAsync();
 *   navigate('/login');
 * };
 */
export function useLogout() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () =>
      apiRequest('/auth/logout', {
        method: 'POST',
      }),
    onSettled: () => {
      // Notify auth state change (cookie is cleared by backend)
      notifyLogout();
      
      // Remove all auth-related queries from cache
      queryClient.removeQueries({ queryKey: queryKeys.auth.user() });
      queryClient.removeQueries({ queryKey: queryKeys.auth.all() });
      
      // Clear all other cached data
      queryClient.clear();
    },
  });
}
