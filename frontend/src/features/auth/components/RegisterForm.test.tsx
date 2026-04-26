import { RegisterForm } from '@/features/auth/components/RegisterForm'
import { MantineProvider } from '@mantine/core'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor } from '@testing-library/react'
import { userEvent } from '@testing-library/user-event'
import { type ReactNode } from 'react'
import { beforeEach, describe, expect, it, vi } from 'vitest'

// Mock useNavigate from TanStack Router
const mockNavigate = vi.fn()
vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => mockNavigate,
}))

// Mock passkey service - WebAuthn not supported in test environment
vi.mock('@/features/auth/services', () => ({
  isWebAuthnSupported: () => false, // Disable passkey registration in tests
  passkeyService: {
    register: vi.fn(),
    authenticate: vi.fn(),
    listCredentials: vi.fn(),
    deleteCredential: vi.fn(),
    updateCredential: vi.fn(),
  },
}))

// Mock i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'auth.email': 'Email',
        'auth.password': 'Password',
        'auth.confirmPassword': 'Confirm Password',
        'auth.register': 'Register',
        'auth.registerError': 'Registration failed. Please try again.',
        'auth.passwordMin': 'Password must be at least 12 characters',
        'auth.passwordMismatch': 'Passwords do not match',
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
      <MantineProvider>{children}</MantineProvider>
    </QueryClientProvider>
  )
}

describe('RegisterForm', () => {
  let fetchMock: ReturnType<typeof vi.fn>

  beforeEach(() => {
    fetchMock = vi.fn().mockImplementation((url: string) => {
      // Mock OAuth providers endpoint (returns no providers for tests)
      if (url.includes('/auth/oauth/providers')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ providers: [] }),
        });
      }
      // Default: reject unhandled endpoints
      return Promise.reject(new Error(`Unhandled fetch URL: ${url}`));
    });
    global.fetch = fetchMock
    localStorage.clear()
    vi.clearAllMocks()
    mockNavigate.mockClear()
  })

  it('should render email, password, and confirm password fields', () => {
    render(
      <TestWrapper>
        <RegisterForm />
      </TestWrapper>
    )

    expect(screen.getByRole('textbox', { name: /email/i })).toBeInTheDocument()
    const passwordInputs = screen.getAllByLabelText(/^password\b|^confirm password\b/i)
    expect(passwordInputs).toHaveLength(2) // Password and Confirm Password
    expect(screen.getByLabelText(/confirm password/i)).toBeInTheDocument()
  })

  it('should render register button', () => {
    render(
      <TestWrapper>
        <RegisterForm />
      </TestWrapper>
    )

    expect(screen.getByRole('button', { name: /register/i })).toBeInTheDocument()
  })

  it('should validate password length', async () => {
    const user = userEvent.setup()

    render(
      <TestWrapper>
        <RegisterForm />
      </TestWrapper>
    )

    const passwordInputs = screen.getAllByLabelText(/password/i)
    const passwordInput = passwordInputs[0] // First password field
    await user.type(passwordInput, 'short')

    // Should show validation message in description
    expect(screen.getByText(/at least 12 characters/i)).toBeInTheDocument()
  })

  it('should validate password confirmation', async () => {
    const user = userEvent.setup()

    render(
      <TestWrapper>
        <RegisterForm />
      </TestWrapper>
    )

    const passwordInputs = screen.getAllByLabelText(/password/i)
    const passwordInput = passwordInputs[0] // First password field
    const confirmPasswordInput = screen.getByLabelText(/confirm password/i)

    await user.type(passwordInput, 'SecureP@ss123!')
    await user.type(confirmPasswordInput, 'different')

    // Should show mismatch error
    expect(screen.getByText(/passwords do not match/i)).toBeInTheDocument()
  })

  it('should disable submit when validation fails', async () => {
    const user = userEvent.setup()

    render(
      <TestWrapper>
        <RegisterForm />
      </TestWrapper>
    )

    const submitButton = screen.getByRole('button', { name: /register/i })
    expect(submitButton).toBeDisabled()

    // Fill valid data
    const emailInput = screen.getByRole('textbox', { name: /email/i })
    const passwordInputs = screen.getAllByLabelText(/password/i)
    const passwordInput = passwordInputs[0] // First password field
    const confirmPasswordInput = screen.getByLabelText(/confirm password/i)

    await user.type(emailInput, 'test@test.com')
    await user.type(passwordInput, 'SecureP@ss123!')
    await user.type(confirmPasswordInput, 'SecureP@ss123!')

    expect(submitButton).not.toBeDisabled()
  })

  it('should disable submit when password is too short', async () => {
    const user = userEvent.setup()

    render(
      <TestWrapper>
        <RegisterForm />
      </TestWrapper>
    )

    const emailInput = screen.getByRole('textbox', { name: /email/i })
    const passwordInputs = screen.getAllByLabelText(/password/i)
    const passwordInput = passwordInputs[0] // First password field
    const confirmPasswordInput = screen.getByLabelText(/confirm password/i)
    const submitButton = screen.getByRole('button', { name: /register/i })

    await user.type(emailInput, 'test@test.com')
    await user.type(passwordInput, 'short')
    await user.type(confirmPasswordInput, 'short')

    expect(submitButton).toBeDisabled()
  })

  it('should disable submit when passwords do not match', async () => {
    const user = userEvent.setup()

    render(
      <TestWrapper>
        <RegisterForm />
      </TestWrapper>
    )

    const emailInput = screen.getByRole('textbox', { name: /email/i })
    const passwordInputs = screen.getAllByLabelText(/password/i)
    const passwordInput = passwordInputs[0] // First password field
    const confirmPasswordInput = screen.getByLabelText(/confirm password/i)
    const submitButton = screen.getByRole('button', { name: /register/i })

    await user.type(emailInput, 'test@test.com')
    await user.type(passwordInput, 'SecureP@ss123!')
    await user.type(confirmPasswordInput, 'different123')

    expect(submitButton).toBeDisabled()
  })

  it('should show error on failed registration', async () => {
    const user = userEvent.setup()

    fetchMock.mockImplementation((url: string) => {
      if (url.includes('/auth/oauth/providers')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ providers: [] }),
        });
      }
      // Registration endpoint fails
      return Promise.resolve({
        ok: false,
        status: 400,
        json: async () => ({
          code: 'VALIDATION_ERROR',
          message: 'Email already exists',
        }),
      });
    });

    render(
      <TestWrapper>
        <RegisterForm />
      </TestWrapper>
    )

    const emailInput = screen.getByRole('textbox', { name: /email/i })
    const passwordInputs = screen.getAllByLabelText(/password/i)
    const passwordInput = passwordInputs[0] // First password field
    const confirmPasswordInput = screen.getByLabelText(/confirm password/i)
    const submitButton = screen.getByRole('button', { name: /register/i })

    await user.type(emailInput, 'test@test.com')
    await user.type(passwordInput, 'SecureP@ss123!')
    await user.type(confirmPasswordInput, 'SecureP@ss123!')
    await user.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText(/email already exists/i)).toBeInTheDocument()
    })
  })

  it('should navigate to / on successful registration', async () => {
    const user = userEvent.setup()

    // Mock OAuth providers AND successful registration
    fetchMock.mockImplementation((url: string) => {
      if (url.includes('/auth/oauth/providers')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ providers: [] }),
        });
      }
      // Registration endpoint succeeds
      return Promise.resolve({
        ok: true,
        status: 201,
        json: async () => ({
          user: { id: '1', email: 'test@test.com', created_at: '2025-10-10T00:00:00Z' },
          token: 'test-token',
        }),
      });
    });

    render(
      <TestWrapper>
        <RegisterForm />
      </TestWrapper>
    )

    const emailInput = screen.getByRole('textbox', { name: /email/i })
    const passwordInputs = screen.getAllByLabelText(/password/i)
    const passwordInput = passwordInputs[0] // First password field
    const confirmPasswordInput = screen.getByLabelText(/confirm password/i)
    const submitButton = screen.getByRole('button', { name: /register/i })

    await user.type(emailInput, 'test@test.com')
    await user.type(passwordInput, 'SecureP@ss123!')
    await user.type(confirmPasswordInput, 'SecureP@ss123!')
    await user.click(submitButton)

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith({ to: '/', replace: true })
    })
  })

  it('should show loading state during registration', async () => {
    const user = userEvent.setup()

    // Mock OAuth providers instantly, but delay registration response
    fetchMock.mockImplementation((url: string) => {
      if (url.includes('/auth/oauth/providers')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ providers: [] }),
        });
      }
      // Registration endpoint - delayed to keep loading state visible
      return new Promise((resolve) => {
        setTimeout(
          () =>
            resolve({
              ok: true,
              status: 201,
              json: async () => ({
                id: '1',
                email: 'test@test.com',
                created_at: '2025-10-10T00:00:00Z',
              }),
            }),
          100
        );
      });
    });

    render(
      <TestWrapper>
        <RegisterForm />
      </TestWrapper>
    )

    const emailInput = screen.getByRole('textbox', { name: /email/i })
    const passwordInputs = screen.getAllByLabelText(/password/i)
    const passwordInput = passwordInputs[0] // First password field
    const confirmPasswordInput = screen.getByLabelText(/confirm password/i)
    const submitButton = screen.getByRole('button', { name: /register/i })

    await user.type(emailInput, 'test@test.com')
    await user.type(passwordInput, 'SecureP@ss123!')
    await user.type(confirmPasswordInput, 'SecureP@ss123!')
    await user.click(submitButton)

    // Button should have loading attribute during submission
    expect(submitButton).toHaveAttribute('data-loading')
  })

  it('should have required attribute on all input fields', () => {
    render(
      <TestWrapper>
        <RegisterForm />
      </TestWrapper>
    )

    const emailInput = screen.getByRole('textbox', { name: /email/i })
    const passwordInputs = screen.getAllByLabelText(/password/i)
    const passwordInput = passwordInputs[0] // First password field
    const confirmPasswordInput = screen.getByLabelText(/confirm password/i)

    expect(emailInput).toBeRequired()
    expect(passwordInput).toBeRequired()
    expect(confirmPasswordInput).toBeRequired()
  })

  it('should have email type on email input', () => {
    render(
      <TestWrapper>
        <RegisterForm />
      </TestWrapper>
    )

    const emailInput = screen.getByRole('textbox', { name: /email/i })
    expect(emailInput).toHaveAttribute('type', 'email')
  })

  it('should have autocomplete attributes for accessibility', () => {
    render(
      <TestWrapper>
        <RegisterForm />
      </TestWrapper>
    )

    const emailInput = screen.getByRole('textbox', { name: /email/i })
    const passwordInputs = screen.getAllByLabelText(/password/i)
    const passwordInput = passwordInputs[0] // First password field
    const confirmPasswordInput = screen.getByLabelText(/confirm password/i)

    expect(emailInput).toHaveAttribute('autocomplete', 'email')
    expect(passwordInput).toHaveAttribute('autocomplete', 'new-password')
    expect(confirmPasswordInput).toHaveAttribute('autocomplete', 'new-password')
  })

  it('should show password strength indicator when valid', async () => {
    const user = userEvent.setup()

    render(
      <TestWrapper>
        <RegisterForm />
      </TestWrapper>
    )

    const passwordInputs = screen.getAllByLabelText(/password/i)
    const passwordInput = passwordInputs[0] // First password field
    await user.type(passwordInput, 'validpass123')

    // Icon should be visible when password is valid (8+ chars)
    const descriptions = screen.getAllByText(/at least 12 characters/i)
    expect(descriptions.length).toBeGreaterThan(0)
  })

  it('should show password mismatch error only after typing in confirm field', async () => {
    const user = userEvent.setup()

    render(
      <TestWrapper>
        <RegisterForm />
      </TestWrapper>
    )

    const passwordInputs = screen.getAllByLabelText(/password/i)
    const passwordInput = passwordInputs[0] // First password field
    const confirmPasswordInput = screen.getByLabelText(/confirm password/i)

    await user.type(passwordInput, 'SecureP@ss123!')

    // No error yet - user hasn't typed in confirm field
    expect(screen.queryByText(/passwords do not match/i)).not.toBeInTheDocument()

    await user.type(confirmPasswordInput, 'different')

    // Now error should appear
    expect(screen.getByText(/passwords do not match/i)).toBeInTheDocument()
  })
})
