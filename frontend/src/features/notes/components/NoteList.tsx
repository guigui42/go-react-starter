import { useNotes, useDeleteNote } from '@/features/notes/hooks/useNotes'
import { NoteCard } from './NoteCard'
import { Button, Center, Loader, SimpleGrid, Stack, Text } from '@mantine/core'
import { notifications } from '@mantine/notifications'
import { IconPlus } from '@tabler/icons-react'
import { Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

export function NoteList() {
  const { t } = useTranslation()
  const { data, isLoading } = useNotes()
  const deleteNote = useDeleteNote()

  const handleDelete = (id: string) => {
    if (window.confirm(t('notes.deleteConfirm'))) {
      deleteNote.mutate(id, {
        onSuccess: () => {
          notifications.show({
            title: t('common.success'),
            message: t('notes.deleteSuccess'),
            color: 'green',
          })
        },
      })
    }
  }

  if (isLoading) {
    return (
      <Center py="xl">
        <Loader size="lg" />
      </Center>
    )
  }

  const notes = data?.notes || []

  return (
    <Stack gap="lg">
      <Button
        component={Link}
        to="/notes/$noteId"
        params={{ noteId: 'new' }}
        leftSection={<IconPlus size={18} />}
      >
        {t('notes.createNote')}
      </Button>

      {notes.length === 0 ? (
        <Center py="xl">
          <Text c="dimmed">{t('notes.noNotes')}</Text>
        </Center>
      ) : (
        <SimpleGrid cols={{ base: 1, sm: 2, lg: 3 }} spacing="md">
          {notes.map((note) => (
            <NoteCard key={note.id} note={note} onDelete={handleDelete} />
          ))}
        </SimpleGrid>
      )}
    </Stack>
  )
}
