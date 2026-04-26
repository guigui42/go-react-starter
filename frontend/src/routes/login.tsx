import { AuthPageControls } from '@/components/AuthPageControls'
import { SEO } from '@/components/SEO'
import { LoginForm } from '@/features/auth/components/LoginForm'
import {
  animateOnMount,
  fadeInUp,
  MotionBox,
  MotionStack,
  staggerContainer,
} from '@/lib/motion'
import { Anchor, Container, Paper, Text, Title } from '@mantine/core'
import { createFileRoute, Link, redirect } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

export const Route = createFileRoute('/login')({
  beforeLoad: ({ context }) => {
    // If user is already authenticated (and not loading), redirect to home
    if (!context.auth.isLoading && context.auth.isAuthenticated) {
      throw redirect({ to: '/' })
    }
  },
  component: LoginPage,
})

function LoginPage() {
  const { t } = useTranslation()

  return (
    <>
      <SEO translationKey="login" path="/login" />
      <Container size={420} my={40}>
        <MotionStack
          gap="md"
          variants={staggerContainer}
          {...animateOnMount}
        >
          <MotionBox variants={fadeInUp}>
            <AuthPageControls />
          </MotionBox>

          <MotionBox variants={fadeInUp}>
            <Title ta="center">{t('auth.login')}</Title>
          </MotionBox>

          <MotionBox variants={fadeInUp}>
            <Paper withBorder shadow="md" p={30} radius="md">
              <LoginForm />

              <Text ta="center" mt="md">
                {t('auth.noAccount')}{' '}
                <Anchor component={Link} to="/register">
                  {t('auth.register')}
                </Anchor>
              </Text>
            </Paper>
          </MotionBox>
        </MotionStack>
      </Container>
    </>
  )
}
