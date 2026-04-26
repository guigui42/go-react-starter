import { useLogin, useLogout, useRegister } from '@/features/auth/hooks/useAuthMutations'
import { useUser } from '@/features/auth/hooks/useUser'
import { AUTH_STATE_CHANGED_EVENT } from '@/lib/authEvents'
import { queryClient } from '@/lib/queryClient'
import { queryKeys } from '@/lib/queryKeys'
import { createContext, useContext, useEffect, useRef, useState, type ReactNode } from 'react'

interface User {
  id: string
  email: string
  is_admin: boolean
  email_verified: boolean
  created_at: string
}

interface AuthContextType {
  user: User | null | undefined
  isLoading: boolean
  login: (email: string, password: string) => Promise<void>
  register: (email: string, password: string) => Promise<void>
  logout: () => Promise<void>
  isAuthenticated: boolean
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  // Track authentication state based on user query result
  const [authStateVersion, setAuthStateVersion] = useState(0)

  // Use TanStack Query to fetch current user and derive loading/auth state
  const userQuery = useUser()
  const loginMutation = useLogin()
  const registerMutation = useRegister()
  const logoutMutation = useLogout()
  
  // Track previous user for security monitoring
  const previousUserRef = useRef<string | null>(null)

  const currentUserId = userQuery.data?.id || null

  // Listen for auth state changes (triggered by login/logout)
  useEffect(() => {
    const handleAuthStateChange = () => {
      // Increment version to trigger re-render and refetch
      setAuthStateVersion(v => v + 1)
      // Invalidate user query to refetch auth state
      queryClient.invalidateQueries({ queryKey: queryKeys.auth.user() })
    }

    // Listen for custom auth state changed events
    window.addEventListener(AUTH_STATE_CHANGED_EVENT, handleAuthStateChange)

    return () => {
      window.removeEventListener(AUTH_STATE_CHANGED_EVENT, handleAuthStateChange)
    }
  }, [])

  // Security: Monitor user changes and clear cache on user switch
  useEffect(() => {
    if (currentUserId && previousUserRef.current && previousUserRef.current !== currentUserId) {
      queryClient.clear()
    }
    previousUserRef.current = currentUserId
  }, [currentUserId])

  // Refetch user data when auth state version changes
  useEffect(() => {
    if (authStateVersion > 0) {
      userQuery.refetch()
    }
  }, [authStateVersion, userQuery])

  const login = async (email: string, password: string) => {
    await loginMutation.mutateAsync({ email, password })
    
    // Give a small delay to ensure cookie is set
    await new Promise(resolve => setTimeout(resolve, 100))
  }

  const register = async (email: string, password: string) => {
    await registerMutation.mutateAsync({ email, password })
  }

  const logout = async () => {
    await logoutMutation.mutateAsync()
  }
  
  return (
    <AuthContext.Provider
      value={{
        user: userQuery.data,
        // Only show loading on initial load, not during background refetches
        // This prevents unnecessary re-renders that could cause form state loss
        isLoading: userQuery.isLoading && !userQuery.data,
        login,
        register,
        logout,
        isAuthenticated: !!userQuery.data,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider')
  }
  return context
}