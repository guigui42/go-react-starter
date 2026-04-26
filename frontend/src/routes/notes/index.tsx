import { useAuth } from '@/contexts/AuthContext'
import { NoteList } from '@/features/notes/components/NoteList'
import {
  animateOnMount,
  fadeInUp,
  MotionBox,
  MotionStack,
  staggerContainer,
} from '@/lib/motion'
import { Center, Container, Loader, Title } from '@mantine/core'
import { createFileRoute, redirect, useNavigate } from '@tanstack/react-router'
import { useEffect } from 'react'
import { useTranslation } from 'react-i18next'

export const Route = createFileRoute('/notes/')({
  beforeLoad: ({ context }) => {
    if (!context.auth.isLoading && !context.auth.isAuthenticated) {
      throw redirect({ to: '/login' })
    }
  },
  component: NotesPage,
})

function NotesPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { isLoading, isAuthenticated } = useAuth()

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      navigate({ to: '/login', replace: true })
    }
  }, [isLoading, isAuthenticated, navigate])

  if (isLoading) {
    return (
      <Center h="50dvh">
        <Loader size="lg" />
      </Center>
    )
  }

  if (!isAuthenticated) {
    return (
      <Center h="50dvh">
        <Loader size="lg" />
      </Center>
    )
  }

  return (
    <Container size="lg" my={{ base: 'md', sm: 40 }}>
      <MotionStack variants={staggerContainer} {...animateOnMount} gap="lg">
        <MotionBox variants={fadeInUp}>
          <Title>{t('notes.title')}</Title>
        </MotionBox>
        <MotionBox variants={fadeInUp}>
          <NoteList />
        </MotionBox>
      </MotionStack>
    </Container>
  )
}
