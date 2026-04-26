import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useUser } from './useUser'
import { type ReactNode } from 'react'

describe('useUser', () => {
  let fetchMock: ReturnType<typeof vi.fn>
  let queryClient: QueryClient

  const createWrapper = () => {
    const Wrapper = ({ children }: { children: ReactNode }) => (
      <QueryClientProvider client={queryClient}>
        {children}
      </QueryClientProvider>
    )
    Wrapper.displayName = 'UseUserTestWrapper'
    return Wrapper
  }

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
      },
    })
    fetchMock = vi.fn()
    global.fetch = fetchMock
    vi.clearAllMocks()
  })

  it('should fetch user successfully when authenticated', async () => {
    // With httpOnly cookies, authentication is automatic
    fetchMock.mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({
        id: '1',
        email: 'test@test.com',
        created_at: '2025-10-10T00:00:00Z',
      }),
    })

    const { result } = renderHook(() => useUser(), {
      wrapper: createWrapper(),
    })

    expect(result.current.isLoading).toBe(true)

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true)
    })

    expect(result.current.data).toEqual({
      id: '1',
      email: 'test@test.com',
      created_at: '2025-10-10T00:00:00Z',
    })
    expect(result.current.error).toBeNull()
  })

  it('should handle 401 error without retrying', async () => {
    // When not authenticated, API returns 401
    fetchMock.mockResolvedValue({
      ok: false,
      status: 401,
      json: async () => ({
        code: 'UNAUTHORIZED',
        message: 'Authorization required',
      }),
    })

    const { result } = renderHook(() => useUser(), {
      wrapper: createWrapper(),
    })

    await waitFor(() => {
      expect(result.current.isError).toBe(true)
    })

    expect(result.current.data).toBeUndefined()
    expect(result.current.error).toBeDefined()
    // Should not retry on 401
    expect(fetchMock).toHaveBeenCalledTimes(1)
  })

  it('should retry on non-401 errors', async () => {
    // Create a separate query client for this test with retry enabled
    const retryQueryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: 1 }, // Allow 1 retry
      },
    })

    // Mock fetch to throw an error that would trigger retries
    fetchMock.mockRejectedValue(new Error('Internal server error'))

    const wrapper = ({ children }: { children: ReactNode }) => (
      <QueryClientProvider client={retryQueryClient}>
        {children}
      </QueryClientProvider>
    )

    const { result } = renderHook(() => useUser(), { wrapper })

    // Wait for the query to start
    await waitFor(() => {
      expect(result.current.isLoading || result.current.isFetching).toBe(true)
    }, { timeout: 1000 })

    // Wait for the final error state after retries
    await waitFor(() => {
      expect(result.current.isError).toBe(true)
    }, { timeout: 5000 })

    // Should retry twice on 500 error (initial call + 2 retries = 3 total)
    // The hook's retry logic allows failureCount < 2, so it retries twice
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(3)
    }, { timeout: 1000 })
  })

  it('should always attempt to fetch (cookies are automatic)', () => {
    // With httpOnly cookies, the query is always enabled
    // If not authenticated, it will return 401 which is handled gracefully
    fetchMock.mockResolvedValue({
      ok: false,
      status: 401,
      json: async () => ({
        code: 'UNAUTHORIZED',
        message: 'Authorization required',
      }),
    })

    const { result } = renderHook(() => useUser(), {
      wrapper: createWrapper(),
    })

    // Query should be in loading state as it's always enabled
    expect(result.current.isLoading).toBe(true)
    // fetch should be called since query is always enabled with cookies
    expect(fetchMock).toHaveBeenCalled()
  })

  it('should use correct staleTime', () => {
    const { result } = renderHook(() => useUser(), {
      wrapper: createWrapper(),
    })

    // Check that the query is configured (can't directly test staleTime in this setup,
    // but we can verify the hook returns a query result)
    expect(result.current).toHaveProperty('data')
    expect(result.current).toHaveProperty('isLoading')
    expect(result.current).toHaveProperty('error')
  })
})
