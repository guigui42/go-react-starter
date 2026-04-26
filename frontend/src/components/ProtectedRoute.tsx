import { type ReactNode } from 'react'
import { Navigate } from '@tanstack/react-router'
import { useUser } from '@/features/auth/hooks/useUser'
import { Loader, Center } from '@mantine/core'

interface ProtectedRouteProps {
  children: ReactNode
}

export function ProtectedRoute({ children }: ProtectedRouteProps) {
  const { data: user, isLoading } = useUser()

  if (isLoading) {
    return (
      <Center h="100dvh">
        <Loader size="lg" />
      </Center>
    )
  }
  
  if (!user) {
    return <Navigate to="/login" />
  }
  
  return <>{children}</>
}
