import { apiRequest, API_PREFIX } from '@/lib/api'
import { queryKeys } from '@/lib/queryKeys'
import { useAuth } from '@/contexts/AuthContext'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type { CreateNoteRequest, Note, NoteListResponse, UpdateNoteRequest } from '../types/note'

export function useNotes() {
  const { user } = useAuth()
  const userId = user?.id || ''

  return useQuery({
    queryKey: queryKeys.notes.list(userId),
    queryFn: () => apiRequest<NoteListResponse>(`${API_PREFIX}/notes`),
    enabled: !!userId,
  })
}

export function useNote(id: string) {
  const { user } = useAuth()
  const userId = user?.id || ''

  return useQuery({
    queryKey: queryKeys.notes.detail(userId, id),
    queryFn: () => apiRequest<Note>(`${API_PREFIX}/notes/${id}`),
    enabled: !!userId && !!id,
  })
}

export function useCreateNote() {
  const queryClient = useQueryClient()
  const { user } = useAuth()
  const userId = user?.id || ''

  return useMutation({
    mutationFn: (data: CreateNoteRequest) =>
      apiRequest<Note>(`${API_PREFIX}/notes`, {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.notes.all(userId) })
    },
  })
}

export function useUpdateNote(id: string) {
  const queryClient = useQueryClient()
  const { user } = useAuth()
  const userId = user?.id || ''

  return useMutation({
    mutationFn: (data: UpdateNoteRequest) =>
      apiRequest<Note>(`${API_PREFIX}/notes/${id}`, {
        method: 'PUT',
        body: JSON.stringify(data),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.notes.all(userId) })
      queryClient.invalidateQueries({ queryKey: queryKeys.notes.detail(userId, id) })
    },
  })
}

export function useDeleteNote() {
  const queryClient = useQueryClient()
  const { user } = useAuth()
  const userId = user?.id || ''

  return useMutation({
    mutationFn: (id: string) =>
      apiRequest<void>(`${API_PREFIX}/notes/${id}`, {
        method: 'DELETE',
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.notes.all(userId) })
    },
  })
}
