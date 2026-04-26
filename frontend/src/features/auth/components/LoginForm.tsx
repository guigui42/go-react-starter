import { useCheckAuthMethods, useLogin, usePasskeyAuthentication } from '@/features/auth/hooks'
import { isWebAuthnSupported } from '@/features/auth/services'
import { ApiError } from '@/lib/api'
import { Alert, Button, Divider, Loader, PasswordInput, Stack, Text, TextInput } from '@mantine/core'
import { IconAlertCircle, IconArrowLeft, IconFingerprint, IconKey, IconMail } from '@tabler/icons-react'
import { useNavigate } from '@tanstack/react-router'
import { useCallback, useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { OAuthButtons } from './OAuthButtons'

// LocalStorage key for storing preferred auth method per email
const AUTH_PREFERENCE_KEY = 'Go React Starter_auth_preference'

type AuthPreference = 'passkey' | 'password'
type LoginStep = 'email' | 'checking' | 'auth_options'

interface StoredPreference {
  email: string
  method: AuthPreference
}

/**
 * Get stored auth preference for an email
 */
function getStoredPreference(email: string): AuthPreference | null {
  try {
    const stored = localStorage.getItem(AUTH_PREFERENCE_KEY)
    if (stored) {
      const pref: StoredPreference = JSON.parse(stored)
      if (pref.email === email.toLowerCase()) {
        return pref.method
      }
    }
  } catch {
    // Ignore localStorage errors
  }
  return null
}

/**
 * Store auth preference for an email
 */
function storePreference(email: string, method: AuthPreference): void {
  try {
    const pref: StoredPreference = { email: email.toLowerCase(), method }
    localStorage.setItem(AUTH_PREFERENCE_KEY, JSON.stringify(pref))
  } catch {
    // Ignore localStorage errors
  }
}

export function LoginForm() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const login = useLogin()
  const passkeyAuth = usePasskeyAuthentication()
  const checkAuthMethods = useCheckAuthMethods()
  
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [step, setStep] = useState<LoginStep>('email')
  const [authMethods, setAuthMethods] = useState<{ hasPassword: boolean; hasPasskey: boolean } | null>(null)
  const [showPasswordField, setShowPasswordField] = useState(false)
  const passwordRef = useRef<HTMLInputElement>(null)
  
  const isPasskeySupported = isWebAuthnSupported()
  
  // Handle email submission - check available auth methods
  const handleEmailSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!email || checkAuthMethods.isPending) {
      return
    }
    
    setStep('checking')
    
    try {
      const result = await checkAuthMethods.mutateAsync(email)
      setAuthMethods({
        hasPassword: result.has_password,
        hasPasskey: result.has_passkey,
      })
      
      setStep('auth_options')
      
      // Check user's stored preference
      const preference = getStoredPreference(email)
      
      // If user has passkey and browser supports it, auto-trigger passkey flow
      // Unless they previously preferred password
      if (result.has_passkey && isPasskeySupported && preference !== 'password') {
        // Auto-trigger passkey authentication
        handlePasskeyLogin()
      } else if (!result.has_passkey || !isPasskeySupported) {
        // No passkey available or not supported - show password field directly
        setShowPasswordField(true)
      }
      // Otherwise, show both options and let user choose
      
    } catch {
      // On error, fall back to showing password field
      setAuthMethods({ hasPassword: true, hasPasskey: false })
      setStep('auth_options')
      setShowPasswordField(true)
    }
  }
  
  // Handle password login
  const handlePasswordSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (login.isPending) {
      return
    }
    
    try {
      await login.mutateAsync({ email, password })
      // Store preference on successful login
      storePreference(email, 'password')
      // Use window.location for a full page reload to ensure router context is updated
      window.location.href = '/'
    } catch (error) {
      // Check for EMAIL_NOT_VERIFIED error and redirect to verification page
      if (error instanceof ApiError && error.code === 'EMAIL_NOT_VERIFIED') {
        const verifyEmail = error.details?.email as string || email
        navigate({ to: '/auth/verify-pending', search: { email: verifyEmail } })
        return
      }
      // Other error handling is managed by the mutation
    }
  }
  
  // Handle passkey login
  const handlePasskeyLogin = useCallback(async () => {
    if (passkeyAuth.isPending || !email) {
      return
    }
    
    try {
      await passkeyAuth.mutateAsync(email)
      // Store preference on successful login
      storePreference(email, 'passkey')
      // Navigation is handled by the mutation's onSuccess
    } catch (error) {
      // Check for EMAIL_NOT_VERIFIED error and redirect to verification page
      if (error instanceof ApiError && error.code === 'EMAIL_NOT_VERIFIED') {
        const verifyEmail = error.details?.email as string || email
        navigate({ to: '/auth/verify-pending', search: { email: verifyEmail } })
        return
      }
      // On passkey failure, show password field as fallback
      setShowPasswordField(true)
    }
  }, [email, passkeyAuth, navigate])
  
  // Reset to email step
  const handleBack = () => {
    setStep('email')
    setAuthMethods(null)
    setShowPasswordField(false)
    setPassword('')
    login.reset()
    passkeyAuth.reset()
    checkAuthMethods.reset()
  }
  
  // Show password field for users who can't use their passkey
  const handleCantUsePasskey = () => {
    setShowPasswordField(true)
  }
  
  // Auto-focus password field when it appears
  useEffect(() => {
    if (showPasswordField && step === 'auth_options') {
      passwordRef.current?.focus()
    }
  }, [showPasswordField, step])
  
  return (
    <Stack gap="md">
      {/* Step 1: Email Input */}
      {step === 'email' && (
        <>
          {/* OAuth buttons at the top */}
          <OAuthButtons />
          
          <form onSubmit={handleEmailSubmit}>
            <Stack gap="md">
              <TextInput
                label={t('auth.email')}
                type="email"
                value={email}
                onChange={(e) => setEmail(e.currentTarget.value)}
                required
                autoComplete="email"
                placeholder={t('auth.emailPlaceholder')}
                autoFocus
                leftSection={<IconMail size={18} />}
              />
              
              <Button
                type="submit"
                loading={checkAuthMethods.isPending}
                size="lg"
                fullWidth
              >
                {t('auth.continue')}
              </Button>
            </Stack>
          </form>
        </>
      )}
      
      {/* Step 2: Checking auth methods */}
      {step === 'checking' && (
        <Stack align="center" gap="md" py="xl">
          <Loader size="lg" />
          <Text c="dimmed">{t('auth.checkingAccount')}</Text>
        </Stack>
      )}
      
      {/* Step 3: Authentication Options */}
      {step === 'auth_options' && authMethods && (
        <Stack gap="md">
          {/* Back button */}
          <Button
            variant="subtle"
            leftSection={<IconArrowLeft size={18} />}
            onClick={handleBack}
            size="sm"
            style={{ alignSelf: 'flex-start' }}
          >
            {t('auth.changeEmail')}
          </Button>
          
          {/* Show email being authenticated */}
          <Text size="sm" c="dimmed" ta="center">
            {t('auth.signingInAs', { email })}
          </Text>
          
          {/* Passkey Error */}
          {passkeyAuth.isError && (
            <Alert 
              icon={<IconAlertCircle />} 
              color={
                passkeyAuth.error instanceof Error && 
                'code' in passkeyAuth.error && 
                passkeyAuth.error.code === 'NO_CREDENTIALS' 
                  ? 'yellow' 
                  : 'red'
              }
            >
              {passkeyAuth.error instanceof Error ? passkeyAuth.error.message : t('auth.passkeyLoginError')}
            </Alert>
          )}
          
          {/* Password Login Error */}
          {login.isError && (
            <Alert icon={<IconAlertCircle />} color="red">
              {login.error instanceof Error ? login.error.message : t('auth.loginError')}
            </Alert>
          )}
          
          {/* Passkey not supported warning */}
          {!isPasskeySupported && authMethods.hasPasskey && (
            <Alert color="yellow" icon={<IconAlertCircle />}>
              <Text size="sm">{t('auth.passkeyNotSupported')}</Text>
            </Alert>
          )}
          
          {/* Passkey Option (if available and supported) */}
          {authMethods.hasPasskey && isPasskeySupported && !showPasswordField && (
            <>
              <Button
                leftSection={<IconFingerprint size={20} />}
                onClick={() => handlePasskeyLogin()}
                loading={passkeyAuth.isPending}
                size="lg"
                fullWidth
              >
                {t('auth.loginWithPasskey')}
              </Button>
              
              {/* Option to switch to password */}
              {authMethods.hasPassword && (
                <>
                  <Divider label={t('common.or')} labelPosition="center" />
                  <Button
                    variant="subtle"
                    leftSection={<IconKey size={18} />}
                    onClick={handleCantUsePasskey}
                    fullWidth
                  >
                    {t('auth.cantUsePasskey')}
                  </Button>
                </>
              )}
            </>
          )}
          
          {/* Password Field (shown when no passkey, passkey not supported, or user chose password) */}
          {showPasswordField && (
            <form onSubmit={handlePasswordSubmit}>
              <Stack gap="md">
                <PasswordInput
                  ref={passwordRef}
                  label={t('auth.password')}
                  value={password}
                  onChange={(e) => setPassword(e.currentTarget.value)}
                  required
                  autoComplete="current-password"
                  autoFocus
                />
                
                <Button 
                  type="submit" 
                  loading={login.isPending} 
                  fullWidth
                  size="lg"
                >
                  {t('auth.login')}
                </Button>
                
                {/* Option to try passkey instead */}
                {authMethods.hasPasskey && isPasskeySupported && (
                  <>
                    <Divider label={t('common.or')} labelPosition="center" />
                    <Button
                      variant="subtle"
                      leftSection={<IconFingerprint size={18} />}
                      onClick={() => {
                        setShowPasswordField(false)
                        handlePasskeyLogin()
                      }}
                      fullWidth
                    >
                      {t('auth.tryPasskeyInstead')}
                    </Button>
                  </>
                )}
              </Stack>
            </form>
          )}
          
          {/* Backup code option - always available when user has passkey */}
          {authMethods.hasPasskey && (
            <Text size="xs" c="dimmed" ta="center">
              <a href="/auth/backup-code" style={{ color: 'inherit' }}>
                {t('auth.passkey.authentication.use_backup_code')}
              </a>
            </Text>
          )}
        </Stack>
      )}
    </Stack>
  )
}