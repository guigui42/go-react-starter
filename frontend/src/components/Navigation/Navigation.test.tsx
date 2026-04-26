import { MantineProvider } from '@mantine/core'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { Navigation } from './Navigation'

// Mock TanStack Router
const mockNavigate = vi.fn()
vi.mock('@tanstack/react-router', async () => {
  const actual = await vi.importActual('@tanstack/react-router')
  return {
    ...actual,
    Link: ({ to, children, className, 'data-active': dataActive }: Record<string, unknown>) => (
      <a href={to as string} className={className as string} data-active={dataActive as boolean | undefined}>
        {children as React.ReactNode}
      </a>
    ),
    useLocation: () => ({ pathname: '/' }),
    useNavigate: () => mockNavigate,
  }
})

// Mock i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'nav.home': 'Home',
        'nav.notes': 'Notes',
        'nav.settings': 'Settings',
        'auth.logout': 'Logout',
        'nav.account': 'Account',
        'nav.toggleNavigation': 'Toggle navigation',
        'nav.language': 'Language',
        'nav.languages.en': 'English',
        'nav.languages.fr': 'Français',
      }
      return translations[key] || key
    },
    i18n: {
      language: 'en',
      changeLanguage: vi.fn(),
    },
  }),
}))

// Mock i18n config
vi.mock('@/i18n/config', () => ({
  default: {
    language: 'en',
    changeLanguage: vi.fn(),
  },
}))

// Mock ColorSchemeToggle
vi.mock('@/components/ColorSchemeToggle', () => ({
  ColorSchemeToggle: () => <div data-testid="color-scheme-toggle">Theme Toggle</div>,
}))

// Mock the entire auth context
const mockLogout = vi.fn()
vi.mock('@/contexts/AuthContext', () => {
  return {
    AuthProvider: ({ children }: { children: React.ReactNode }) => children,
    useAuth: () => {
      return {
        user: { id: '1', email: 'test@example.com', created_at: '2025-01-01' },
        token: 'test-token',
        isLoading: false,
        login: vi.fn(),
        register: vi.fn(),
        logout: mockLogout,
        isAuthenticated: true,
      }
    },
  }
})

// Mock the UserPreferencesContext
const mockSetLanguage = vi.fn()
vi.mock('@/contexts/UserPreferencesContext', () => ({
  UserPreferencesProvider: ({ children }: { children: React.ReactNode }) => children,
  useUserPreferencesContext: () => ({
    colorScheme: 'auto',
    language: 'en',
    setColorScheme: vi.fn(),
    setLanguage: mockSetLanguage,
    isLoading: false,
  }),
}))

const renderNavigation = (props = {}) => {
  return render(
    <MantineProvider>
      <Navigation {...props} />
    </MantineProvider>
  )
}

describe('Navigation', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.clearAllMocks()
    mockNavigate.mockClear()
    mockLogout.mockClear()
  })

  describe('Accessibility', () => {
    it('should have proper navigation landmark', () => {
      renderNavigation()
      const nav = screen.getByRole('navigation')
      expect(nav).toBeInTheDocument()
      expect(nav).toHaveAttribute('aria-label', 'Main navigation')
    })

    it('should have accessible links with proper labels', () => {
      renderNavigation()
      
      expect(screen.getByRole('link', { name: /home/i })).toBeInTheDocument()
      expect(screen.getByRole('link', { name: /notes/i })).toBeInTheDocument()
    })

    it('should have keyboard-accessible user menu button', () => {
      renderNavigation()
      const userButton = screen.getByRole('button', { name: /test@example\.com/i })
      expect(userButton).toBeInTheDocument()
    })

    it('should include theme toggle control', () => {
      renderNavigation()
      expect(screen.getByTestId('color-scheme-toggle')).toBeInTheDocument()
    })
  })

  describe('Navigation Links', () => {
    it('should render all main navigation links', () => {
      renderNavigation()
      
      expect(screen.getByText('Home')).toBeInTheDocument()
      expect(screen.getByText('Notes')).toBeInTheDocument()
    })

    it('should have correct href attributes for all links', () => {
      renderNavigation()
      
      expect(screen.getByRole('link', { name: /home/i })).toHaveAttribute('href', '/')
      expect(screen.getByRole('link', { name: /notes/i })).toHaveAttribute('href', '/notes')
    })
  })

  describe('Authentication State', () => {
    it('should show user menu when authenticated', () => {
      renderNavigation()
      
      const userButton = screen.getByRole('button', { name: /test@example\.com/i })
      expect(userButton).toBeInTheDocument()
    })

    it('should display user email when authenticated', () => {
      renderNavigation()
      
      expect(screen.getByText('test@example.com')).toBeInTheDocument()
    })
  })

  describe('Responsive Behavior', () => {
    it('should render mobile burger menu button', () => {
      renderNavigation()
      
      const burgerButton = screen.getByLabelText(/toggle navigation/i)
      expect(burgerButton).toBeInTheDocument()
    })
  })

  describe('Logo', () => {
    it('should render application logo/title', () => {
      renderNavigation()
      
      expect(screen.getByText('Go React Starter')).toBeInTheDocument()
    })

    it('should link logo to home page', () => {
      renderNavigation()
      
      const logoLink = screen.getByText('Go React Starter').closest('a')
      expect(logoLink).toHaveAttribute('href', '/')
    })
  })

  describe('Logout', () => {
    it('should call logout and navigate to login page when logout is clicked', async () => {
      const user = userEvent.setup()
      
      const locationMock = { href: '' }
      Object.defineProperty(window, 'location', {
        writable: true,
        value: locationMock,
      })
      
      mockLogout.mockResolvedValueOnce(undefined)
      
      renderNavigation()
      
      const userButton = screen.getByRole('button', { name: /test@example\.com/i })
      await user.click(userButton)
      
      const logoutButton = await screen.findByText('Logout')
      await user.click(logoutButton)
      
      expect(mockLogout).toHaveBeenCalledTimes(1)
      
      await vi.waitFor(() => {
        expect(locationMock.href).toBe('/login')
      })
    })
  })
})
