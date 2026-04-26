import { usePasskeyRegistration, useRegister } from '@/features/auth/hooks'
import { isWebAuthnSupported } from '@/features/auth/services'
import { Alert, Anchor, Badge, Button, Group, Paper, PasswordInput, Stack, Text, TextInput, ThemeIcon } from '@mantine/core'
import { IconAlertCircle, IconCheck, IconFingerprint, IconKey } from '@tabler/icons-react'
import { useNavigate } from '@tanstack/react-router'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { OAuthButtons } from './OAuthButtons'

type RegistrationMethod = 'passkey' | 'password'

export function RegisterForm() {
  const { t } = useTranslation()
  const register = useRegister()
  const passkeyRegistration = usePasskeyRegistration()
  const navigate = useNavigate()

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [friendlyName, setFriendlyName] = useState('')
  const [validationError, setValidationError] = useState<string | null>(null)
  
  const isPasskeySupported = isWebAuthnSupported()
  const [method, setMethod] = useState<RegistrationMethod>(isPasskeySupported ? 'passkey' : 'password')

  const passwordMatch = password === confirmPassword
  const passwordValid = password.length >= 12

  const handlePasswordRegister = async (e: React.FormEvent) => {
    e.preventDefault()
    setValidationError(null)

    if (!passwordValid) {
      setValidationError(t('auth.passwordMin'))
      return
    }

    if (!passwordMatch) {
      setValidationError(t('auth.passwordMismatch'))
      return
    }

    try {
      const result = await register.mutateAsync({ email, password })
      // Check if email verification is required
      if ('message' in result && result.message === 'verification_email_sent') {
        // Redirect to verification pending page
        navigate({ to: '/auth/verify-pending', search: { email: result.email }, replace: true })
        return
      }
      // Use replace to avoid adding register page to history
      // Redirect to broker creation for onboarding flow
      navigate({ to: '/', replace: true })
    } catch {
      // Error is already stored in register.error
    }
  }

  const handlePasskeyRegister = async (e: React.FormEvent) => {
    e.preventDefault()
    setValidationError(null)

    if (!email) {
      setValidationError(t('auth.emailRequired'))
      return
    }

    try {
      // Step 1: Create passwordless account (backend returns token, auto-login)
      const result = await register.mutateAsync({ email, password: '' })
      
      // Check if email verification is required
      if ('message' in result && result.message === 'verification_email_sent') {
        // Redirect to verification pending page
        navigate({ to: '/auth/verify-pending', search: { email: result.email }, replace: true })
        return
      }
      
      // Step 2: Small delay to ensure token is properly stored and propagated
      // This ensures localStorage is updated and the API client can read it
      await new Promise(resolve => setTimeout(resolve, 100))
      
      // Step 3: User is now authenticated, register passkey
      await passkeyRegistration.mutateAsync(friendlyName || undefined)
      
      // Navigation handled by passkey mutation
    } catch {
      // Error is already stored in mutations
    }
  }

  const switchMethod = () => {
    setMethod(method === 'passkey' ? 'password' : 'passkey')
    setValidationError(null)
    // Reset password fields when switching
    setPassword('')
    setConfirmPassword('')
  }

  return (
    <Stack gap="md">
      {/* OAuth buttons (includes divider) */}
      <OAuthButtons />
      
      {/* Method selector - only show if passkey is supported */}
      {isPasskeySupported && (
        <Paper
          p="sm"
          radius="md"
          withBorder
          style={{
            borderColor: method === 'passkey' ? 'var(--mantine-color-blue-6)' : undefined,
            backgroundColor: method === 'passkey' ? 'var(--mantine-color-blue-light)' : undefined,
          }}
        >
          <Group gap="sm" wrap="nowrap">
            <ThemeIcon
              size="lg"
              radius="md"
              variant={method === 'passkey' ? 'filled' : 'light'}
              color={method === 'passkey' ? 'blue' : 'gray'}
            >
              <IconFingerprint size={20} />
            </ThemeIcon>
            <div style={{ flex: 1 }}>
              <Group gap="xs">
                <Text size="sm" fw={500}>
                  {t('auth.registerWithPasskey')}
                </Text>
                <Badge size="xs" variant="light" color="blue">
                  {t('auth.recommended')}
                </Badge>
              </Group>
              <Text size="xs" c="dimmed">
                {t('auth.registerWithPasskeyDesc')}
              </Text>
            </div>
            {method !== 'passkey' && (
              <Anchor
                component="button"
                type="button"
                size="xs"
                onClick={() => setMethod('passkey')}
              >
                {t('auth.useThis')}
              </Anchor>
            )}
          </Group>
        </Paper>
      )}

      {method === 'passkey' && isPasskeySupported ? (
        <form onSubmit={handlePasskeyRegister}>
          <Stack>
            {(validationError || register.isError || passkeyRegistration.isError) && (
              <Alert icon={<IconAlertCircle />} color="red">
                {validationError || 
                 (register.error instanceof Error ? register.error.message : null) ||
                 (passkeyRegistration.error instanceof Error ? passkeyRegistration.error.message : t('auth.registerError'))}
              </Alert>
            )}

            <TextInput
              label={t('auth.email')}
              type="email"
              value={email}
              onChange={(e) => setEmail(e.currentTarget.value)}
              required
              autoComplete="email"
            />

            <TextInput
              label={t('auth.passkeyDeviceName')}
              placeholder={t('auth.passkeyDeviceNamePlaceholder')}
              value={friendlyName}
              onChange={(e) => setFriendlyName(e.currentTarget.value)}
              description={t('auth.passkeyDeviceNameDesc')}
              maxLength={50}
            />

            <Button
              type="submit"
              loading={register.isPending || passkeyRegistration.isPending}
              leftSection={<IconFingerprint size={20} />}
              fullWidth
            >
              {t('auth.registerWithPasskeyButton')}
            </Button>

            <Text ta="center" size="sm" c="dimmed">
              <Anchor component="button" type="button" size="sm" onClick={switchMethod}>
                <Group gap={4} justify="center">
                  <IconKey size={14} />
                  {t('auth.preferPassword')}
                </Group>
              </Anchor>
            </Text>
          </Stack>
        </form>
      ) : (
        <form onSubmit={handlePasswordRegister}>
          <Stack>
            {(validationError || register.isError) && (
              <Alert icon={<IconAlertCircle />} color="red">
                {validationError || (register.error instanceof Error ? register.error.message : t('auth.registerError'))}
              </Alert>
            )}

            <TextInput
              label={t('auth.email')}
              type="email"
              value={email}
              onChange={(e) => setEmail(e.currentTarget.value)}
              required
              autoComplete="email"
            />

            <PasswordInput
              label={t('auth.password')}
              value={password}
              onChange={(e) => setPassword(e.currentTarget.value)}
              required
              autoComplete="new-password"
              description={
                <Text size="xs" c={passwordValid ? 'green' : 'dimmed'} component="span">
                  {passwordValid && <IconCheck size={12} />}
                  {t('auth.passwordMin')}
                </Text>
              }
            />

            <PasswordInput
              label={t('auth.confirmPassword')}
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.currentTarget.value)}
              required
              autoComplete="new-password"
              error={confirmPassword && !passwordMatch ? t('auth.passwordMismatch') : undefined}
            />

            <Button
              type="submit"
              loading={register.isPending}
              fullWidth
              disabled={!passwordValid || !passwordMatch}
            >
              {t('auth.register')}
            </Button>

            {isPasskeySupported && (
              <Text ta="center" size="sm" c="dimmed">
                <Anchor component="button" type="button" size="sm" onClick={switchMethod}>
                  <Group gap={4} justify="center">
                    <IconFingerprint size={14} />
                    {t('auth.preferPasskey')}
                  </Group>
                </Anchor>
              </Text>
            )}
          </Stack>
        </form>
      )}
    </Stack>
  )
}
