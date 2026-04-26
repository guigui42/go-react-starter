/**
 * Email Tester Component
 *
 * Admin component for testing email configuration by sending test emails.
 * Shows current email configuration status and allows sending test emails.
 */

import { useEmailConfig, useSendTestEmail } from '@/features/admin/hooks'
import {
    Alert,
    Badge,
    Button,
    Card,
    Group,
    Loader,
    Stack,
    Text,
    TextInput,
    Textarea,
    Title,
} from '@mantine/core'
import { useForm } from '@mantine/form'
import {
    IconAlertCircle,
    IconCheck,
    IconMail,
    IconMailOff,
    IconSend,
    IconX,
} from '@tabler/icons-react'
import { useState } from 'react'

interface TestEmailFormValues {
  to: string
  subject: string
  message: string
}

export function EmailTester() {
  const { data: emailConfig, isLoading: configLoading, error: configError } = useEmailConfig()
  const sendTestEmail = useSendTestEmail()
  const [lastResult, setLastResult] = useState<{ success: boolean; message: string } | null>(null)

  const form = useForm<TestEmailFormValues>({
    initialValues: {
      to: '',
      subject: '',
      message: '',
    },
    validate: {
      to: (value) => {
        if (!value.trim()) return 'Email address is required'
        // Basic email validation
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
        if (!emailRegex.test(value)) return 'Invalid email address'
        return null
      },
    },
  })

  const handleSubmit = async (values: TestEmailFormValues) => {
    setLastResult(null)
    try {
      const result = await sendTestEmail.mutateAsync({
        to: values.to,
        subject: values.subject || undefined,
        message: values.message || undefined,
      })
      setLastResult({
        success: result.success,
        message: result.message,
      })
      if (result.success) {
        form.reset()
      }
    } catch (error) {
      setLastResult({
        success: false,
        message: error instanceof Error ? error.message : 'Failed to send test email',
      })
    }
  }

  if (configLoading) {
    return (
      <Card shadow="sm" padding="lg" radius="md" withBorder>
        <Stack align="center" py="xl">
          <Loader size="md" />
          <Text size="sm" c="dimmed">Loading email configuration...</Text>
        </Stack>
      </Card>
    )
  }

  if (configError) {
    return (
      <Alert
        icon={<IconAlertCircle size={16} />}
        title="Error Loading Configuration"
        color="red"
        variant="light"
      >
        Failed to load email configuration: {configError.message}
      </Alert>
    )
  }

  const isConfigured = emailConfig?.configured ?? false

  return (
    <Stack gap="lg">
      {/* Configuration Status Card */}
      <Card shadow="sm" padding="lg" radius="md" withBorder>
        <Stack gap="md">
          <Group justify="space-between" align="flex-start">
            <Title order={4} style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              {isConfigured ? <IconMail size={20} /> : <IconMailOff size={20} />}
              Email Configuration
            </Title>
            <Badge
              color={isConfigured ? 'green' : 'gray'}
              variant="light"
              size="lg"
              leftSection={isConfigured ? <IconCheck size={12} /> : <IconX size={12} />}
            >
              {isConfigured ? 'Configured' : 'Not Configured'}
            </Badge>
          </Group>

          {isConfigured ? (
            <Stack gap="xs">
              <Group gap="xs">
                <Text size="sm" fw={500} w={100}>Provider:</Text>
                <Text size="sm" c="dimmed">{emailConfig?.provider}</Text>
              </Group>
              <Group gap="xs">
                <Text size="sm" fw={500} w={100}>From:</Text>
                <Text size="sm" c="dimmed">
                  {emailConfig?.from_name ? `${emailConfig.from_name} <${emailConfig.from_address}>` : emailConfig?.from_address}
                </Text>
              </Group>
            </Stack>
          ) : (
            <Alert
              icon={<IconAlertCircle size={16} />}
              color="yellow"
              variant="light"
            >
              <Text size="sm">
                Email is not configured. Set up SMTP configuration in your environment variables to enable email functionality.
              </Text>
            </Alert>
          )}
        </Stack>
      </Card>

      {/* Test Email Form */}
      <Card shadow="sm" padding="lg" radius="md" withBorder>
        <Stack gap="md">
          <Title order={4} style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <IconSend size={20} />
            Send Test Email
          </Title>

          {!isConfigured && (
            <Alert color="gray" variant="light">
              <Text size="sm">
                Email is not configured. Test emails cannot be sent until SMTP is configured.
              </Text>
            </Alert>
          )}

          <form onSubmit={form.onSubmit(handleSubmit)}>
            <Stack gap="md">
              <TextInput
                label="Recipient Email"
                placeholder="test@example.com"
                description="The email address to send the test email to"
                required
                disabled={!isConfigured}
                {...form.getInputProps('to')}
              />

              <TextInput
                label="Subject"
                placeholder="Leave empty for default subject"
                description="Optional: Custom email subject (default: 'Go React Starter Test Email')"
                disabled={!isConfigured}
                {...form.getInputProps('subject')}
              />

              <Textarea
                label="Message"
                placeholder="Leave empty for default message"
                description="Optional: Custom email body content"
                rows={3}
                disabled={!isConfigured}
                {...form.getInputProps('message')}
              />

              {lastResult && (
                <Alert
                  icon={lastResult.success ? <IconCheck size={16} /> : <IconX size={16} />}
                  color={lastResult.success ? 'green' : 'red'}
                  variant="light"
                >
                  {lastResult.message}
                </Alert>
              )}

              <Group justify="flex-end">
                <Button
                  type="submit"
                  leftSection={<IconSend size={16} />}
                  loading={sendTestEmail.isPending}
                  disabled={!isConfigured}
                >
                  Send Test Email
                </Button>
              </Group>
            </Stack>
          </form>
        </Stack>
      </Card>
    </Stack>
  )
}
