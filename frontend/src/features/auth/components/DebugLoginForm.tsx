import { useAuth } from '@/contexts/AuthContext'
import { useLogin } from '@/features/auth/hooks'
import { Alert, Button, PasswordInput, Stack, TextInput } from '@mantine/core'
import { IconAlertCircle } from '@tabler/icons-react'
import { useNavigate } from '@tanstack/react-router'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

export function LoginForm() {
  const { t } = useTranslation()
  const login = useLogin()
  const auth = useAuth()
  const navigate = useNavigate()
  
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    // Prevent double submission
    if (login.isPending) {
      return
    }
    
    try {
      await auth.login(email, password)
      await navigate({ to: '/', replace: true })
    } catch {
      // Error handling is managed by the mutation and displayed via login.error
    }
  }
  
  return (
    <form onSubmit={handleSubmit}>
      <Stack>
        {login.isError && (
          <Alert icon={<IconAlertCircle />} color="red">
            {login.error instanceof Error ? login.error.message : t('auth.loginError')}
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
          autoComplete="current-password"
        />
        
        <Button type="submit" loading={login.isPending} fullWidth>
          {t('auth.login')}
        </Button>
      </Stack>
    </form>
  )
}