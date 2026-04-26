import { useAuth } from '@/contexts/AuthContext'
import { useUpdateUserPreferences, useUserPreferences } from '@/features/user/hooks/useUserPreferences'
import type { ColorScheme } from '@/features/user/types/preferences'
import i18n from '@/i18n/config'
import { useMantineColorScheme } from '@mantine/core'
import { createContext, useContext, useEffect, useRef, useState, type ReactNode } from 'react'

interface UserPreferencesContextType {
  /** Current color scheme (light, dark, or auto) */
  colorScheme: ColorScheme
  /** Current language */
  language: 'en' | 'fr'
  /** Update color scheme - syncs to backend when logged in */
  setColorScheme: (scheme: ColorScheme) => void
  /** Update language - syncs to backend when logged in */
  setLanguage: (lang: 'en' | 'fr') => void
  /** Whether preferences are loading */
  isLoading: boolean
}

const UserPreferencesContext = createContext<UserPreferencesContextType | undefined>(undefined)

/**
 * Detects the system's preferred language.
 * Returns 'fr' if the browser language starts with 'fr', otherwise 'en'.
 */
function detectSystemLanguage(): 'en' | 'fr' {
  const browserLang = navigator.language || (navigator as { userLanguage?: string }).userLanguage || 'en'
  return browserLang.startsWith('fr') ? 'fr' : 'en'
}

/**
 * Provider that manages user preferences (theme and language) based on authentication state.
 * 
 * When logged out:
 * - Color scheme defaults to 'auto' (system preference)
 * - Language defaults to browser/system language
 * - Manual changes work immediately but aren't persisted
 * 
 * When logged in:
 * - Color scheme and language sync from user's saved preferences on login
 * - Changes are applied optimistically and persisted to the backend
 * 
 * Note: Using useEffect for data sync is the recommended TanStack Query v5 pattern
 * since onSuccess was removed from queries. Structural sharing prevents unnecessary
 * effect execution on background refetches.
 */
export function UserPreferencesProvider({ children }: { children: ReactNode }) {
  const { isAuthenticated, isLoading: authLoading } = useAuth()
  const { setColorScheme: setMantineColorScheme } = useMantineColorScheme()
  const { data: preferences, isLoading: prefsLoading } = useUserPreferences()
  const updatePreferences = useUpdateUserPreferences()

  // Local state for tracking user's explicit color scheme choice (when not authenticated)
  const [localColorScheme, setLocalColorScheme] = useState<ColorScheme>('auto')

  // Track auth state to detect login/logout transitions
  const previousAuthState = useRef<boolean | null>(null)
  
  // Track if initial language sync has been done
  const hasInitializedLanguage = useRef(false)

  // Sync color scheme when auth state changes (login/logout)
  useEffect(() => {
    if (authLoading) return

    const justLoggedIn = previousAuthState.current === false && isAuthenticated
    const justLoggedOut = previousAuthState.current === true && !isAuthenticated
    const isInitialLoad = previousAuthState.current === null
    
    previousAuthState.current = isAuthenticated

    // Only run initial setup once
    if (isInitialLoad) {
      if (isAuthenticated && preferences?.color_scheme) {
        setMantineColorScheme(preferences.color_scheme)
      }
      // Don't set 'auto' on initial load - let Mantine use its default
      return
    }

    if (justLoggedIn && preferences?.color_scheme) {
      // User just logged in - sync their saved preferences
      setMantineColorScheme(preferences.color_scheme)
    } else if (justLoggedOut) {
      // User logged out - reset to system preference
      setMantineColorScheme('auto')
    }
  }, [isAuthenticated, authLoading, preferences?.color_scheme, setMantineColorScheme])

  // Initialize language on first load (before auth check completes)
  useEffect(() => {
    if (hasInitializedLanguage.current) return
    
    // Set system language immediately for unauthenticated users
    // This runs once on mount, before auth state is known
    const systemLang = detectSystemLanguage()
    if (i18n.language !== systemLang) {
      i18n.changeLanguage(systemLang)
    }
    hasInitializedLanguage.current = true
  }, [])

  // Sync language when user logs in or out
  useEffect(() => {
    if (authLoading) return

    // When authenticated and preferences load, use user's saved language
    if (isAuthenticated && preferences?.language) {
      if (i18n.language !== preferences.language) {
        i18n.changeLanguage(preferences.language)
      }
    }
  }, [isAuthenticated, authLoading, preferences?.language])

  // Handler to update color scheme - applies optimistically, persists if logged in
  const setColorScheme = (scheme: ColorScheme) => {
    // Update local state for tracking
    setLocalColorScheme(scheme)
    // Apply to Mantine immediately
    setMantineColorScheme(scheme)
    
    if (isAuthenticated) {
      updatePreferences.mutate({ color_scheme: scheme })
    }
  }

  // Handler to update language - applies optimistically, persists if logged in
  const setLanguage = (lang: 'en' | 'fr') => {
    // Apply immediately to i18n
    i18n.changeLanguage(lang)
    
    if (isAuthenticated) {
      updatePreferences.mutate({ language: lang })
    }
  }

  // Determine the effective color scheme to expose in context
  // When authenticated: use backend preferences, fallback to local
  // When not authenticated: use local state
  const effectiveColorScheme = isAuthenticated 
    ? (preferences?.color_scheme || localColorScheme) 
    : localColorScheme
  const effectiveLanguage = preferences?.language || (i18n.language as 'en' | 'fr')

  return (
    <UserPreferencesContext.Provider
      value={{
        colorScheme: effectiveColorScheme,
        language: effectiveLanguage,
        setColorScheme,
        setLanguage,
        isLoading: authLoading || (isAuthenticated && prefsLoading),
      }}
    >
      {children}
    </UserPreferencesContext.Provider>
  )
}

/**
 * Default values for user preferences when context is not yet available.
 * This handles edge cases during initial render before providers are mounted.
 */
const defaultPreferences: UserPreferencesContextType = {
  colorScheme: 'auto',
  language: 'en',
  setColorScheme: () => {},
  setLanguage: () => {},
  isLoading: true,
}

/**
 * Hook to access and update user preferences (theme and language).
 * Returns default values if called before UserPreferencesProvider is mounted
 * (e.g., during TanStack Router's initial render).
 */
// eslint-disable-next-line react-refresh/only-export-components
export function useUserPreferencesContext() {
  const context = useContext(UserPreferencesContext)
  // Return defaults if context is not yet available (during initial render)
  // This prevents errors when TanStack Router renders before providers mount
  return context ?? defaultPreferences
}
