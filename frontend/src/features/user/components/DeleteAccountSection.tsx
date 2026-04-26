import { deleteUserAccount } from '@/lib/api'
import { Alert, Button, List, Modal, Paper, Stack, Text, TextInput, Title } from '@mantine/core'
import { useDisclosure } from '@mantine/hooks'
import { IconAlertTriangle, IconTrash } from '@tabler/icons-react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from '@tanstack/react-router'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useAuth } from '@/contexts/AuthContext'
import { notifications } from '@mantine/notifications'

export function DeleteAccountSection() {
  const { t } = useTranslation()
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [opened, { open, close }] = useDisclosure(false)
  const [confirmEmail, setConfirmEmail] = useState('')

  const deleteMutation = useMutation({
    mutationFn: () => deleteUserAccount(confirmEmail),
    onSuccess: async () => {
      // Show success notification first
      notifications.show({
        title: t('common.success'),
        message: t('settings.deleteAccount.deleteSuccess'),
        color: 'green',
        autoClose: 3000,
      })
      
      // Clear all cached data
      queryClient.clear()
      
      // Redirect to home page
      navigate({ to: '/', replace: true })
      
      // Logout last (clears cookie and auth state)
      await logout()
    },
    onError: (error: Error) => {
      notifications.show({
        title: t('common.error'),
        message: error.message || t('settings.deleteAccount.deleteError'),
        color: 'red',
      })
    },
  })

  const handleDelete = () => {
    if (confirmEmail === user?.email) {
      deleteMutation.mutate()
    }
  }

  const handleClose = () => {
    setConfirmEmail('')
    close()
  }

  return (
    <>
      <Paper withBorder shadow="sm" p="md">
        <Stack gap="md">
          <div>
            <Title order={3} c="red">
              {t('settings.deleteAccount.title')}
            </Title>
            <Text size="sm" c="dimmed" mt={4}>
              {t('settings.deleteAccount.description')}
            </Text>
          </div>

          <Button
            color="red"
            variant="outline"
            leftSection={<IconTrash size={16} />}
            onClick={open}
          >
            {t('settings.deleteAccount.button')}
          </Button>
        </Stack>
      </Paper>

      <Modal
        opened={opened}
        onClose={handleClose}
        title={t('settings.deleteAccount.confirmTitle')}
        centered
        size="lg"
      >
        <Stack gap="md">
          <Alert color="red" icon={<IconAlertTriangle size={20} />}>
            <Text size="sm" fw={500}>
              {t('settings.deleteAccount.warning')}
            </Text>
            <List size="sm" mt="sm">
              <List.Item>{t('settings.deleteAccount.dataList.preferences')}</List.Item>
              <List.Item>{t('settings.deleteAccount.dataList.credentials')}</List.Item>
            </List>
          </Alert>

          <TextInput
            label={t('settings.deleteAccount.emailConfirmLabel')}
            placeholder={t('settings.deleteAccount.emailPlaceholder')}
            value={confirmEmail}
            onChange={(e) => setConfirmEmail(e.currentTarget.value)}
            required
            error={
              confirmEmail && confirmEmail !== user?.email
                ? t('settings.deleteAccount.emailMismatch')
                : undefined
            }
          />

          <Stack gap="xs" mt="md">
            <Button
              color="red"
              fullWidth
              disabled={confirmEmail !== user?.email}
              loading={deleteMutation.isPending}
              onClick={handleDelete}
            >
              {t('settings.deleteAccount.confirmButton')}
            </Button>
            <Button variant="default" fullWidth onClick={handleClose}>
              {t('settings.deleteAccount.cancelButton')}
            </Button>
          </Stack>
        </Stack>
      </Modal>
    </>
  )
}
