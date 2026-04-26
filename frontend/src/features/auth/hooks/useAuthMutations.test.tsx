import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useLogin, useRegister, useLogout } from './useAuthMutations';
import { queryKeys } from '@/lib/queryKeys';
import * as api from '@/lib/api';

// Mock the API module
vi.mock('@/lib/api', () => ({
  apiRequest: vi.fn(),
  ApiError: class ApiError extends Error {
    status: number;
    code: string;
    details?: Record<string, unknown>;
    constructor(status: number, code: string, message: string, details?: Record<string, unknown>) {
      super(message);
      this.name = 'ApiError';
      this.status = status;
      this.code = code;
      this.details = details;
    }
  },
}));

describe('useLogin', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
    });
    vi.clearAllMocks();
  });

  const wrapper = ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );

  it('should successfully login and cache user data', async () => {
    const mockResponse = {
      user: { id: '1', email: 'test@example.com', created_at: '2025-01-01' },
      token: 'test-token-123',
    };

    vi.mocked(api.apiRequest).mockResolvedValueOnce(mockResponse);

    const { result } = renderHook(() => useLogin(), { wrapper });

    result.current.mutate({ email: 'test@example.com', password: 'password123' });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(api.apiRequest).toHaveBeenCalledWith('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email: 'test@example.com', password: 'password123' }),
    });

    // Token is set via httpOnly cookie by backend, not in localStorage
    expect(result.current.data).toEqual(mockResponse);
  });

  it('should cache user data in query client on success', async () => {
    const mockResponse = {
      user: { id: '1', email: 'test@example.com', created_at: '2025-01-01' },
      token: 'test-token-123',
    };

    vi.mocked(api.apiRequest).mockResolvedValueOnce(mockResponse);

    const { result } = renderHook(() => useLogin(), { wrapper });

    result.current.mutate({ email: 'test@example.com', password: 'password123' });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    const cachedUser = queryClient.getQueryData(queryKeys.auth.user());
    expect(cachedUser).toEqual(mockResponse.user);
  });

  it('should handle login errors', async () => {
    const mockError = new api.ApiError(401, 'INVALID_CREDENTIALS', 'Invalid email or password');
    vi.mocked(api.apiRequest).mockRejectedValueOnce(mockError);

    const { result } = renderHook(() => useLogin(), { wrapper });

    result.current.mutate({ email: 'wrong@example.com', password: 'wrongpass' });

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });

    expect(result.current.error).toEqual(mockError);
  });

  it('should not retry on authentication failures', async () => {
    const mockError = new api.ApiError(401, 'INVALID_CREDENTIALS', 'Invalid credentials');
    vi.mocked(api.apiRequest).mockRejectedValue(mockError);

    const { result } = renderHook(() => useLogin(), { wrapper });

    result.current.mutate({ email: 'test@example.com', password: 'wrong' });

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });

    // Should only call once (no retries)
    expect(api.apiRequest).toHaveBeenCalledTimes(1);
  });
});

describe('useRegister', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
    });
    vi.clearAllMocks();
  });

  const wrapper = ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );

  it('should successfully register and auto-login', async () => {
    const mockResponse = {
      user: {
        id: '1',
        email: 'newuser@example.com',
        created_at: '2025-01-01',
      },
      token: 'new-token-123',
    };

    vi.mocked(api.apiRequest).mockResolvedValueOnce(mockResponse);

    const { result } = renderHook(() => useRegister(), { wrapper });

    await result.current.mutateAsync({ email: 'newuser@example.com', password: 'password123' });

    expect(api.apiRequest).toHaveBeenCalledWith('/auth/register', {
      method: 'POST',
      body: JSON.stringify({ email: 'newuser@example.com', password: 'password123' }),
    });
    expect(api.apiRequest).toHaveBeenCalledTimes(1); // Only one call
    // Token is set via httpOnly cookie by backend
    expect(queryClient.getQueryData(queryKeys.auth.user())).toEqual(mockResponse.user);
  });

  it('should handle registration errors', async () => {
    const mockError = new api.ApiError(400, 'EMAIL_EXISTS', 'Email already registered');
    vi.mocked(api.apiRequest).mockRejectedValueOnce(mockError);

    const { result } = renderHook(() => useRegister(), { wrapper });

    try {
      await result.current.mutateAsync({ email: 'existing@example.com', password: 'password123' });
      expect.fail('Should have thrown');
    } catch (error) {
      expect(error).toEqual(mockError);
    }
  });
});

describe('useLogout', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
    });
    // Add some cached data to verify it gets cleared
    queryClient.setQueryData(queryKeys.auth.user(), { id: '1', email: 'test@example.com' });
    queryClient.setQueryData(queryKeys.notes.all('test-user'), [{ id: '1', title: 'Test' }]);
    vi.clearAllMocks();
  });

  const wrapper = ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );

  it('should successfully logout and clear all data', async () => {
    vi.mocked(api.apiRequest).mockResolvedValueOnce(undefined);

    const { result } = renderHook(() => useLogout(), { wrapper });

    await result.current.mutateAsync();

    expect(api.apiRequest).toHaveBeenCalledWith('/auth/logout', { method: 'POST' });
    // Cookie is cleared by backend via Set-Cookie header

    // Verify query cache is cleared
    expect(queryClient.getQueryData(queryKeys.auth.user())).toBeUndefined();
    expect(queryClient.getQueryData(queryKeys.notes.all('test-user'))).toBeUndefined();
  });

  it('should clear local data even if API call fails', async () => {
    const mockError = new api.ApiError(500, 'SERVER_ERROR', 'Server error');
    vi.mocked(api.apiRequest).mockRejectedValueOnce(mockError);

    const { result } = renderHook(() => useLogout(), { wrapper });

    // Should not throw since onSettled clears data anyway
    await expect(result.current.mutateAsync()).rejects.toThrow();

    // Should still clear local data even on API error
    expect(queryClient.getQueryData(queryKeys.auth.user())).toBeUndefined();
    expect(queryClient.getQueryData(queryKeys.notes.all('test-user'))).toBeUndefined();
  });

  it('should handle network errors gracefully', async () => {
    vi.mocked(api.apiRequest).mockRejectedValueOnce(new Error('Network error'));

    const { result } = renderHook(() => useLogout(), { wrapper });

    result.current.mutate();

    await waitFor(() => {
      expect(result.current.isError || result.current.isSuccess).toBe(true);
    });

    // Cache should still be cleared
    expect(queryClient.getQueryData(queryKeys.auth.user())).toBeUndefined();
  });
});
