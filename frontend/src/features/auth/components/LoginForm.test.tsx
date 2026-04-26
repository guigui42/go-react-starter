import { AuthProvider } from '@/contexts/AuthContext'
import { MantineProvider } from '@mantine/core'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor } from '@testing-library/react'
import { userEvent } from '@testing-library/user-event'
import { type ReactNode } from 'react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { LoginForm } from './LoginForm'

// Mock useNavigate from TanStack Router
const mockNavigate = vi.fn()
vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => mockNavigate,
}))

// Mock passkey service - WebAuthn not supported in test environment
vi.mock('@/features/auth/services', () => ({
  isWebAuthnSupported: () => false, // Disable passkey login in tests
}))

// Mock i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, params?: Record<string, string>) => {
      const translations: Record<string, string> = {
        'auth.email': 'Email',
        'auth.password': 'Password',
        'auth.login': 'Login',
        'auth.continue': 'Continue',
        'auth.checkingAccount': 'Checking account...',
        'auth.changeEmail': 'Change email',
        'auth.signingInAs': `Signing in as ${params?.email || ''}`,
        'auth.loginError': 'Invalid credentials',
        'auth.cantUsePasskey': "Can't use your passkey?",
        'auth.tryPasskeyInstead': 'Try passkey instead',
        'auth.loginWithPasskey': 'Sign in with Passkey',
        'auth.passkeyLoginError': 'Passkey login failed',
        'auth.passkeyNotSupported': 'Passkeys are not supported',
        'auth.passkey.authentication.use_backup_code': 'Use a backup code',
        'common.or': 'or',
      }
      return translations[key] || key
    },
  }),
}))

const TestWrapper = ({ children }: { children: ReactNode }) => {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <MantineProvider>{children}</MantineProvider>
      </AuthProvider>
    </QueryClientProvider>
  )
}

describe('LoginForm', () => {
  let fetchMock: ReturnType<typeof vi.fn>
  
  // Store original location and fetch
  const originalLocation = window.location
  const originalFetch = globalThis.fetch
  
  beforeEach(() => {
    fetchMock = vi.fn()
    globalThis.fetch = fetchMock as unknown as typeof fetch
    localStorage.clear()
    vi.clearAllMocks()
    mockNavigate.mockClear()
    
    // Default mock setup
    fetchMock.mockImplementation((url: string) => {
      if (url.includes('/auth/me')) {
        return Promise.resolve({
          ok: false,
          status: 401,
          json: async () => ({ code: 'UNAUTHORIZED', message: 'Not authenticated' }),
        })
      }
      // Default: check-methods returns password only (since WebAuthn is mocked as unsupported)
      if (url.includes('/auth/check-methods')) {
        return Promise.resolve({
          ok: true,
          status: 200,
          json: async () => ({ has_password: true, has_passkey: false }),
        })
      }
      // Default fallback for unmocked requests
      return Promise.resolve({
        ok: false,
        status: 404,
        json: async () => ({ code: 'NOT_FOUND', message: 'Not found' }),
      })
    })
    
    // Mock window.location.href for navigation tests
    Object.defineProperty(window, 'location', {
      value: { ...originalLocation, href: '' },
      writable: true,
      configurable: true,
    })
  })
  
  afterEach(() => {
    // Restore original location and fetch
    Object.defineProperty(window, 'location', {
      value: originalLocation,
      writable: true,
      configurable: true,
    })
    globalThis.fetch = originalFetch
  })

  it('should render email field and continue button on initial load', () => {
    render(
      <TestWrapper>
        <LoginForm />
      </TestWrapper>
    )

    expect(screen.getByLabelText(/email/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /continue/i })).toBeInTheDocument()
    // Password field should NOT be visible initially
    expect(screen.queryByLabelText(/password/i)).not.toBeInTheDocument()
  })

  it('should show password field after entering email and clicking continue', async () => {
    const user = userEvent.setup()
    
    render(
      <TestWrapper>
        <LoginForm />
      </TestWrapper>
    )

    const emailInput = screen.getByLabelText(/email/i)
    const continueButton = screen.getByRole('button', { name: /continue/i })

    await user.type(emailInput, 'test@test.com')
    await user.click(continueButton)

    // Wait for auth options step to show password field
    await waitFor(() => {
      expect(screen.getByLabelText(/^password\b/i)).toBeInTheDocument()
    })
    
    // Login button should now be visible
    expect(screen.getByRole('button', { name: /^login$/i })).toBeInTheDocument()
  })

  it('should allow entering email and then password', async () => {
    const user = userEvent.setup()
    
    render(
      <TestWrapper>
        <LoginForm />
      </TestWrapper>
    )

    // Step 1: Enter email
    const emailInput = screen.getByLabelText(/email/i)
    await user.type(emailInput, 'test@test.com')
    expect(emailInput).toHaveValue('test@test.com')

    // Click continue
    await user.click(screen.getByRole('button', { name: /continue/i }))

    // Wait for password field
    await waitFor(() => {
      expect(screen.getByLabelText(/^password\b/i)).toBeInTheDocument()
    })

    // Step 2: Enter password
    const passwordInput = screen.getByLabelText(/^password\b/i)
    await user.type(passwordInput, 'SecureP@ss123!')
    expect(passwordInput).toHaveValue('SecureP@ss123!')
  })

  it('should show error on failed login', async () => {
    const user = userEvent.setup()
    
    // Override the mock to handle login failure
    fetchMock.mockImplementation((url: string) => {
      if (url.includes('/auth/me')) {
        return Promise.resolve({
          ok: false,
          status: 401,
          json: async () => ({ code: 'UNAUTHORIZED', message: 'Not authenticated' }),
        })
      }
      if (url.includes('/auth/check-methods')) {
        return Promise.resolve({
          ok: true,
          status: 200,
          json: async () => ({ has_password: true, has_passkey: false }),
        })
      }
      if (url.includes('/auth/login')) {
        return Promise.resolve({
          ok: false,
          status: 401,
          json: async () => ({ code: 'UNAUTHORIZED', message: 'Invalid credentials' }),
        })
      }
      return Promise.resolve({
        ok: false,
        status: 404,
        json: async () => ({ code: 'NOT_FOUND', message: 'Not found' }),
      })
    })

    render(
      <TestWrapper>
        <LoginForm />
      </TestWrapper>
    )

    // Step 1: Enter email and continue
    await user.type(screen.getByLabelText(/email/i), 'test@test.com')
    await user.click(screen.getByRole('button', { name: /continue/i }))

    // Wait for password field
    await waitFor(() => {
      expect(screen.getByLabelText(/^password\b/i)).toBeInTheDocument()
    })

    // Step 2: Enter password and submit
    await user.type(screen.getByLabelText(/^password\b/i), 'wrongpass')
    await user.click(screen.getByRole('button', { name: /^login$/i }))

    // Should show error
    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeInTheDocument()
      expect(screen.getByText(/invalid credentials/i)).toBeInTheDocument()
    })
  })

  it('should navigate to dashboard on successful login', async () => {
    const user = userEvent.setup()
    
    // Override the mock to handle successful login
    fetchMock.mockImplementation((url: string) => {
      if (url.includes('/auth/me')) {
        return Promise.resolve({
          ok: false,
          status: 401,
          json: async () => ({ code: 'UNAUTHORIZED', message: 'Not authenticated' }),
        })
      }
      if (url.includes('/auth/check-methods')) {
        return Promise.resolve({
          ok: true,
          status: 200,
          json: async () => ({ has_password: true, has_passkey: false }),
        })
      }
      if (url.includes('/auth/login')) {
        return Promise.resolve({
          ok: true,
          status: 200,
          json: async () => ({
            user: { id: '1', email: 'test@test.com', created_at: '2025-10-10T00:00:00Z' },
            token: 'test-token',
          }),
        })
      }
      return Promise.resolve({
        ok: false,
        status: 404,
        json: async () => ({ code: 'NOT_FOUND', message: 'Not found' }),
      })
    })

    render(
      <TestWrapper>
        <LoginForm />
      </TestWrapper>
    )

    // Step 1: Enter email and continue
    await user.type(screen.getByLabelText(/email/i), 'test@test.com')
    await user.click(screen.getByRole('button', { name: /continue/i }))

    // Wait for password field
    await waitFor(() => {
      expect(screen.getByLabelText(/^password\b/i)).toBeInTheDocument()
    })

    // Step 2: Enter password and submit
    await user.type(screen.getByLabelText(/^password\b/i), 'SecureP@ss123!')
    await user.click(screen.getByRole('button', { name: /^login$/i }))

    await waitFor(() => {
      // LoginForm uses window.location.href for full page reload after login
      expect(window.location.href).toBe('/')
    })
  })

  it('should show loading state during login', async () => {
    const user = userEvent.setup()
    
    // Override mock to delay /auth/login
    fetchMock.mockImplementation((url: string) => {
      if (url.includes('/auth/me')) {
        return Promise.resolve({
          ok: false,
          status: 401,
          json: async () => ({ code: 'UNAUTHORIZED', message: 'Not authenticated' }),
        })
      }
      if (url.includes('/auth/check-methods')) {
        return Promise.resolve({
          ok: true,
          status: 200,
          json: async () => ({ has_password: true, has_passkey: false }),
        })
      }
      if (url.includes('/auth/login')) {
        // Delay the response to keep loading state visible
        return new Promise((resolve) => {
          setTimeout(
            () =>
              resolve({
                ok: true,
                status: 200,
                json: async () => ({
                  user: { id: '1', email: 'test@test.com', created_at: '2025-10-10T00:00:00Z' },
                  token: 'test-token',
                }),
              }),
            100
          )
        })
      }
      return Promise.resolve({
        ok: false,
        status: 404,
        json: async () => ({ code: 'NOT_FOUND', message: 'Not found' }),
      })
    })

    render(
      <TestWrapper>
        <LoginForm />
      </TestWrapper>
    )

    // Step 1: Enter email and continue
    await user.type(screen.getByLabelText(/email/i), 'test@test.com')
    await user.click(screen.getByRole('button', { name: /continue/i }))

    // Wait for password field
    await waitFor(() => {
      expect(screen.getByLabelText(/^password\b/i)).toBeInTheDocument()
    })

    // Step 2: Enter password and submit
    await user.type(screen.getByLabelText(/^password\b/i), 'SecureP@ss123!')
    
    const submitButton = screen.getByRole('button', { name: /^login$/i })
    const clickPromise = user.click(submitButton)

    // Check for loading state - Mantine Button shows loading state
    await waitFor(() => {
      expect(submitButton).toHaveAttribute('data-loading', 'true')
    })

    // Wait for the login process to complete
    await clickPromise

    await waitFor(
      () => {
        expect(window.location.href).toBe('/')
      },
      { timeout: 2000 }
    )
  })

  it('should have required attribute on email field', () => {
    render(
      <TestWrapper>
        <LoginForm />
      </TestWrapper>
    )

    const emailInput = screen.getByLabelText(/email/i)
    expect(emailInput).toBeRequired()
  })

  it('should have required attribute on password field after continue', async () => {
    const user = userEvent.setup()
    
    render(
      <TestWrapper>
        <LoginForm />
      </TestWrapper>
    )

    // Enter email and continue
    await user.type(screen.getByLabelText(/email/i), 'test@test.com')
    await user.click(screen.getByRole('button', { name: /continue/i }))

    // Wait for password field
    await waitFor(() => {
      const passwordInput = screen.getByLabelText(/^password\b/i)
      expect(passwordInput).toBeRequired()
    })
  })

  it('should have email type on email input', () => {
    render(
      <TestWrapper>
        <LoginForm />
      </TestWrapper>
    )

    const emailInput = screen.getByLabelText(/email/i)
    expect(emailInput).toHaveAttribute('type', 'email')
  })

  it('should have autocomplete attributes for accessibility', async () => {
    const user = userEvent.setup()
    
    render(
      <TestWrapper>
        <LoginForm />
      </TestWrapper>
    )

    // Check email autocomplete
    const emailInput = screen.getByLabelText(/email/i)
    expect(emailInput).toHaveAttribute('autocomplete', 'email')

    // Enter email and continue to get password field
    await user.type(emailInput, 'test@test.com')
    await user.click(screen.getByRole('button', { name: /continue/i }))

    // Wait for password field and check autocomplete
    await waitFor(() => {
      const passwordInput = screen.getByLabelText(/^password\b/i)
      expect(passwordInput).toHaveAttribute('autocomplete', 'current-password')
    })
  })

  it('should clear error on new submission attempt', async () => {
    const user = userEvent.setup()
    
    // Track login attempts
    let loginAttempts = 0
    
    // Override the default mock
    fetchMock.mockImplementation((url: string) => {
      if (url.includes('/auth/me')) {
        return Promise.resolve({
          ok: false,
          status: 401,
          json: async () => ({ code: 'UNAUTHORIZED', message: 'Not authenticated' }),
        })
      }
      if (url.includes('/auth/check-methods')) {
        return Promise.resolve({
          ok: true,
          status: 200,
          json: async () => ({ has_password: true, has_passkey: false }),
        })
      }
      if (url.includes('/auth/login')) {
        loginAttempts++
        if (loginAttempts === 1) {
          // First login attempt fails
          return Promise.resolve({
            ok: false,
            status: 401,
            json: async () => ({ code: 'UNAUTHORIZED', message: 'Invalid credentials' }),
          })
        } else {
          // Second login attempt succeeds
          return Promise.resolve({
            ok: true,
            status: 200,
            json: async () => ({
              user: { id: '1', email: 'test@test.com', created_at: '2025-10-10T00:00:00Z' },
              token: 'test-token',
            }),
          })
        }
      }
      return Promise.resolve({
        ok: false,
        status: 404,
        json: async () => ({ code: 'NOT_FOUND', message: 'Not found' }),
      })
    })

    render(
      <TestWrapper>
        <LoginForm />
      </TestWrapper>
    )

    // Step 1: Enter email and continue
    await user.type(screen.getByLabelText(/email/i), 'test@test.com')
    await user.click(screen.getByRole('button', { name: /continue/i }))

    // Wait for password field
    await waitFor(() => {
      expect(screen.getByLabelText(/^password\b/i)).toBeInTheDocument()
    })

    // First attempt: wrong password
    const passwordInput = screen.getByLabelText(/^password\b/i)
    await user.type(passwordInput, 'wrongpass')
    await user.click(screen.getByRole('button', { name: /^login$/i }))

    // Should show error
    await waitFor(() => {
      expect(screen.getByText(/invalid credentials/i)).toBeInTheDocument()
    })

    // Second attempt: correct password
    await user.clear(passwordInput)
    await user.type(passwordInput, 'correctpass')
    await user.click(screen.getByRole('button', { name: /^login$/i }))

    // Error should be cleared and login should succeed
    await waitFor(() => {
      expect(screen.queryByText(/invalid credentials/i)).not.toBeInTheDocument()
    })
  })

  it('should show change email button and allow going back', async () => {
    const user = userEvent.setup()
    
    render(
      <TestWrapper>
        <LoginForm />
      </TestWrapper>
    )

    // Enter email and continue
    await user.type(screen.getByLabelText(/email/i), 'test@test.com')
    await user.click(screen.getByRole('button', { name: /continue/i }))

    // Wait for auth options step
    await waitFor(() => {
      expect(screen.getByLabelText(/^password\b/i)).toBeInTheDocument()
    })

    // Should show "Change email" button
    const changeEmailButton = screen.getByRole('button', { name: /change email/i })
    expect(changeEmailButton).toBeInTheDocument()

    // Click change email to go back
    await user.click(changeEmailButton)

    // Should be back to email step
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /continue/i })).toBeInTheDocument()
      expect(screen.queryByLabelText(/password/i)).not.toBeInTheDocument()
    })
  })

  it('should show signing in as email text', async () => {
    const user = userEvent.setup()
    
    render(
      <TestWrapper>
        <LoginForm />
      </TestWrapper>
    )

    // Enter email and continue
    await user.type(screen.getByLabelText(/email/i), 'test@test.com')
    await user.click(screen.getByRole('button', { name: /continue/i }))

    // Wait for auth options step
    await waitFor(() => {
      expect(screen.getByText(/signing in as test@test.com/i)).toBeInTheDocument()
    })
  })
})
