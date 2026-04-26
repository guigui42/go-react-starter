import { useAuth } from '@/contexts/AuthContext'
import { NoteForm } from '@/features/notes/components/NoteForm'
import { useNote } from '@/features/notes/hooks/useNotes'
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

export const Route = createFileRoute('/notes/$noteId')({
  beforeLoad: ({ context }) => {
    if (!context.auth.isLoading && !context.auth.isAuthenticated) {
      throw redirect({ to: '/login' })
    }
  },
  component: NoteDetailPage,
})

function NoteDetailPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { noteId } = Route.useParams()
  const { isLoading: authLoading, isAuthenticated } = useAuth()
  const { data: note, isLoading: noteLoading } = useNote(noteId)

  useEffect(() => {
    if (!authLoading && !isAuthenticated) {
      navigate({ to: '/login', replace: true })
    }
  }, [authLoading, isAuthenticated, navigate])

  if (authLoading || noteLoading) {
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
    <Container size="md" my={{ base: 'md', sm: 40 }}>
      <MotionStack variants={staggerContainer} {...animateOnMount} gap="lg">
        <MotionBox variants={fadeInUp}>
          <Title>{t('notes.editNote')}</Title>
        </MotionBox>
        <MotionBox variants={fadeInUp}>
          <NoteForm note={note} />
        </MotionBox>
      </MotionStack>
    </Container>
  )
}
