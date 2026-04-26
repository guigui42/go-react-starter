import { useAuth } from '@/contexts/AuthContext'
import { AdminDashboard } from '@/features/admin'
import { Center, Loader } from '@mantine/core'
import { notifications } from '@mantine/notifications'
import { createFileRoute, redirect, useNavigate } from '@tanstack/react-router'
import { useEffect } from 'react'

export const Route = createFileRoute('/admin/')({
  beforeLoad: ({ context }) => {
    // Don't redirect while auth is still loading
    if (!context.auth.isLoading && !context.auth.isAuthenticated) {
      throw redirect({ to: '/login' })
    }
  },
  component: AdminRoute,
})

function AdminRoute() {
  const navigate = useNavigate()
  const { isLoading, isAuthenticated, user } = useAuth()

  // Redirect non-admin users to home
  useEffect(() => {
    if (!isLoading && isAuthenticated && user && !user.is_admin) {
      notifications.show({
        title: 'Access Denied',
        message: 'You do not have permission to access the admin dashboard',
        color: 'red',
      })
      navigate({ to: '/', replace: true })
    }
  }, [isLoading, isAuthenticated, user, navigate])

  // Redirect to login if not authenticated after loading completes
  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      navigate({ to: '/login', replace: true })
    }
  }, [isLoading, isAuthenticated, navigate])

  // Show loading state while authentication is being verified
  if (isLoading) {
    return (
      <Center h="100dvh">
        <Loader size="lg" />
      </Center>
    )
  }

  // Don't render content if not authenticated (will redirect)
  if (!isAuthenticated) {
    return (
      <Center h="100dvh">
        <Loader size="lg" />
      </Center>
    )
  }

  // Don't render content if not admin (will redirect)
  if (user && !user.is_admin) {
    return (
      <Center h="100dvh">
        <Loader size="lg" />
      </Center>
    )
  }

  return <AdminDashboard />
}
