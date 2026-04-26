import type { Note } from '@/features/notes/types/note'
import { ActionIcon, Card, Group, Text, Title } from '@mantine/core'
import { IconEdit, IconTrash } from '@tabler/icons-react'
import { Link } from '@tanstack/react-router'
import dayjs from 'dayjs'

interface NoteCardProps {
  note: Note
  onDelete: (id: string) => void
}

export function NoteCard({ note, onDelete }: NoteCardProps) {
  return (
    <Card withBorder shadow="sm" padding="lg" radius="md">
      <Group justify="space-between" mb="xs">
        <Title order={4}>{note.title}</Title>
        <Group gap="xs">
          <ActionIcon
            component={Link}
            to="/notes/$noteId"
            params={{ noteId: note.id }}
            variant="subtle"
            color="blue"
            aria-label="Edit note"
          >
            <IconEdit size={18} />
          </ActionIcon>
          <ActionIcon
            variant="subtle"
            color="red"
            onClick={() => onDelete(note.id)}
            aria-label="Delete note"
          >
            <IconTrash size={18} />
          </ActionIcon>
        </Group>
      </Group>
      <Text c="dimmed" size="sm" lineClamp={3}>
        {note.content}
      </Text>
      <Text c="dimmed" size="xs" mt="md">
        {dayjs(note.updated_at).format('MMM D, YYYY h:mm A')}
      </Text>
    </Card>
  )
}
