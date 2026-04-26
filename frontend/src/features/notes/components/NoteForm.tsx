import { useCreateNote, useUpdateNote } from '../hooks/useNotes'
import type { Note } from '../types/note'
import { Button, Group, Paper, Stack, Textarea, TextInput } from '@mantine/core'
import { useForm } from '@mantine/form'
import { notifications } from '@mantine/notifications'
import { useNavigate } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

interface NoteFormProps {
  note?: Note | null
}

export function NoteForm({ note }: NoteFormProps) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const isEditing = !!note && note.id !== 'new'
  const createNote = useCreateNote()
  const updateNote = useUpdateNote(note?.id || '')

  const form = useForm({
    initialValues: {
      title: note?.title || '',
      content: note?.content || '',
    },
    validate: {
      title: (value) => (value.trim().length === 0 ? 'Title is required' : null),
    },
  })

  const handleSubmit = form.onSubmit(async (values) => {
    try {
      if (isEditing) {
        await updateNote.mutateAsync(values)
        notifications.show({
          title: t('common.success'),
          message: t('notes.updateSuccess'),
          color: 'green',
        })
      } else {
        await createNote.mutateAsync(values)
        notifications.show({
          title: t('common.success'),
          message: t('notes.createSuccess'),
          color: 'green',
        })
      }
      navigate({ to: '/notes' })
    } catch {
      notifications.show({
        title: t('common.error'),
        message: t('errors.serverError'),
        color: 'red',
      })
    }
  })

  return (
    <Paper withBorder shadow="sm" p="xl" radius="md">
      <form onSubmit={handleSubmit}>
        <Stack gap="md">
          <TextInput
            label={t('notes.noteTitle')}
            placeholder={t('notes.noteTitle')}
            required
            {...form.getInputProps('title')}
          />
          <Textarea
            label={t('notes.noteContent')}
            placeholder={t('notes.noteContent')}
            minRows={6}
            autosize
            {...form.getInputProps('content')}
          />
          <Group justify="flex-end">
            <Button variant="default" onClick={() => navigate({ to: '/notes' })}>
              {t('common.cancel')}
            </Button>
            <Button
              type="submit"
              loading={createNote.isPending || updateNote.isPending}
            >
              {isEditing ? t('common.save') : t('notes.createNote')}
            </Button>
          </Group>
        </Stack>
      </form>
    </Paper>
  )
}
