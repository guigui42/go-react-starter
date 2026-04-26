import { passkeyService, type PasskeyCredential } from '@/features/auth/services';
import { ActionIcon, Alert, Badge, Button, Card, Group, Loader, Modal, Stack, Text } from '@mantine/core';
import { IconAlertCircle, IconDevices, IconFingerprint, IconPlus, IconTrash } from '@tabler/icons-react';
import { useCallback, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { PasskeyRegistrationForm } from './PasskeyRegistrationForm';

export function PasskeyManagementPanel() {
  const { t } = useTranslation();
  const [credentials, setCredentials] = useState<PasskeyCredential[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [deleteModalOpen, setDeleteModalOpen] = useState(false);
  const [credentialToDelete, setCredentialToDelete] = useState<string | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);
  const [registerModalOpen, setRegisterModalOpen] = useState(false);
  const [hasLoaded, setHasLoaded] = useState(false);

  const loadCredentials = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const creds = await passkeyService.listCredentials();
      setCredentials(creds);
    } catch (err) {
      setError(err instanceof Error ? err.message : t('passkey.errors.load_failed'));
    } finally {
      setIsLoading(false);
    }
  }, [t]);

  // Trigger initial load once during render (React-recommended pattern)
  if (!hasLoaded) {
    setHasLoaded(true);
    loadCredentials();
  }

  const handleDelete = async () => {
    if (!credentialToDelete) return;

    setIsDeleting(true);
    try {
      await passkeyService.deleteCredential(credentialToDelete);
      setCredentials(credentials.filter((c) => c.id !== credentialToDelete));
      setDeleteModalOpen(false);
      setCredentialToDelete(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : t('passkey.errors.delete_failed'));
    } finally {
      setIsDeleting(false);
    }
  };

  const openDeleteModal = (credentialId: string) => {
    setCredentialToDelete(credentialId);
    setDeleteModalOpen(true);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString(undefined, {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  const handleRegistrationSuccess = () => {
    setRegisterModalOpen(false);
    loadCredentials();
  };

  if (isLoading) {
    return (
      <Group justify="center" p="xl">
        <Loader size="sm" />
      </Group>
    );
  }

  return (
    <Stack gap="md">
      <Group justify="space-between">
        <div>
          <Text size="lg" fw={600}>
            {t('passkey.management.title')}
          </Text>
          <Text size="sm" c="dimmed">
            {t('passkey.management.description')}
          </Text>
        </div>
        <Button
          leftSection={<IconPlus size={16} />}
          onClick={() => setRegisterModalOpen(true)}
        >
          {t('passkey.management.add_button')}
        </Button>
      </Group>

      {error && (
        <Alert icon={<IconAlertCircle size={16} />} title={t('common.error')} color="red" onClose={() => setError(null)}>
          {error}
        </Alert>
      )}

      {credentials.length === 0 ? (
        <Card withBorder p="xl">
          <Stack align="center" gap="md">
            <IconDevices size={48} stroke={1.5} color="var(--mantine-color-dimmed)" />
            <Text c="dimmed" ta="center">
              {t('passkey.management.no_credentials')}
            </Text>
            <Button
              leftSection={<IconPlus size={16} />}
              onClick={() => setRegisterModalOpen(true)}
            >
              {t('passkey.management.add_first')}
            </Button>
          </Stack>
        </Card>
      ) : (
        <Stack gap="xs">
          {credentials.map((credential) => (
            <Card key={credential.id} withBorder p="md">
              <Group justify="space-between">
                <Group>
                  <IconFingerprint size={24} stroke={1.5} />
                  <div>
                    <Text fw={500}>{credential.friendlyName || t('passkey.management.unnamed_device')}</Text>
                    <Group gap="xs">
                      <Text size="xs" c="dimmed">
                        {t('passkey.management.created')}: {formatDate(credential.createdAt)}
                      </Text>
                      {credential.lastUsedAt && (
                        <>
                          <Text size="xs" c="dimmed">
                            •
                          </Text>
                          <Text size="xs" c="dimmed">
                            {t('passkey.management.last_used')}: {formatDate(credential.lastUsedAt)}
                          </Text>
                        </>
                      )}
                    </Group>
                    <Group gap="xs" mt={4}>
                      {credential.backupEligible && (
                        <Badge size="xs" variant="light" color="blue">
                          {t('passkey.management.synced')}
                        </Badge>
                      )}
                      <Badge size="xs" variant="light">
                        {credential.authenticatorAttachment === 'platform'
                          ? t('passkey.management.platform')
                          : t('passkey.management.cross_platform')}
                      </Badge>
                    </Group>
                  </div>
                </Group>
                <ActionIcon
                  color="red"
                  variant="subtle"
                  onClick={() => openDeleteModal(credential.id)}
                  aria-label={t('passkey.management.delete_credential')}
                >
                  <IconTrash size={18} />
                </ActionIcon>
              </Group>
            </Card>
          ))}
        </Stack>
      )}

      <Modal
        opened={deleteModalOpen}
        onClose={() => setDeleteModalOpen(false)}
        title={t('passkey.management.delete_confirm_title')}
      >
        <Stack gap="md">
          <Text size="sm">{t('passkey.management.delete_confirm_message')}</Text>
          <Group justify="flex-end">
            <Button variant="default" onClick={() => setDeleteModalOpen(false)}>
              {t('common.cancel')}
            </Button>
            <Button color="red" onClick={handleDelete} loading={isDeleting}>
              {t('common.delete')}
            </Button>
          </Group>
        </Stack>
      </Modal>

      <Modal
        opened={registerModalOpen}
        onClose={() => setRegisterModalOpen(false)}
        title={t('passkey.registration.modal_title')}
      >
        <PasskeyRegistrationForm onSuccess={handleRegistrationSuccess} />
      </Modal>
    </Stack>
  );
}
