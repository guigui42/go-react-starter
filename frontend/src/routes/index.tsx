import { SEO } from '@/components/SEO'
import { useAuth } from '@/contexts/AuthContext'
import {
  animateOnMount,
  fadeInUp,
  MotionBox,
  MotionStack,
  staggerContainer,
} from '@/lib/motion'
import { Button, Container, Paper, Stack, Text, Title } from '@mantine/core'
import { IconNote, IconSettings } from '@tabler/icons-react'
import { createFileRoute, Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

export const Route = createFileRoute('/')({
  component: HomePage,
})

function HomePage() {
  const { t } = useTranslation()
  const { isAuthenticated, user } = useAuth()

  return (
    <>
      <SEO translationKey="home" path="/" />
      <Container size="sm" my={60}>
        <MotionStack
          gap="xl"
          variants={staggerContainer}
          {...animateOnMount}
        >
          <MotionBox variants={fadeInUp}>
            <Stack align="center" gap="md">
              <Title ta="center">
                {t('seo.home.title')}
              </Title>
              <Text c="dimmed" ta="center" size="lg">
                {t('seo.home.description')}
              </Text>
            </Stack>
          </MotionBox>

          {isAuthenticated ? (
            <MotionBox variants={fadeInUp}>
              <Paper withBorder shadow="sm" p="xl" radius="md">
                <Stack gap="md">
                  <Text fw={500}>
                    Welcome back, {user?.email}!
                  </Text>
                  <Button
                    component={Link}
                    to="/notes"
                    leftSection={<IconNote size={18} />}
                    fullWidth
                  >
                    {t('nav.notes')}
                  </Button>
                  <Button
                    component={Link}
                    to="/settings"
                    variant="light"
                    leftSection={<IconSettings size={18} />}
                    fullWidth
                  >
                    {t('nav.settings')}
                  </Button>
                </Stack>
              </Paper>
            </MotionBox>
          ) : (
            <MotionBox variants={fadeInUp}>
              <Stack gap="sm" align="center">
                <Button component={Link} to="/login" size="lg">
                  {t('auth.login')}
                </Button>
                <Button component={Link} to="/register" variant="light" size="lg">
                  {t('auth.register')}
                </Button>
              </Stack>
            </MotionBox>
          )}
        </MotionStack>
      </Container>
    </>
  )
}
