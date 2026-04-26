import { AuthPageControls } from '@/components/AuthPageControls'
import { SEO } from '@/components/SEO'
import { RegisterForm } from '@/features/auth/components/RegisterForm'
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

export const Route = createFileRoute('/register')({
  beforeLoad: ({ context }) => {
    // If user is already authenticated (and not loading), redirect to home
    if (!context.auth.isLoading && context.auth.isAuthenticated) {
      throw redirect({ to: '/' })
    }
  },
  component: RegisterPage,
})

function RegisterPage() {
  const { t } = useTranslation()

  return (
    <>
      <SEO translationKey="register" path="/register" />
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
            <Title ta="center">{t('auth.register')}</Title>
          </MotionBox>

          <MotionBox variants={fadeInUp}>
            <Paper withBorder shadow="md" p={30} radius="md">
              <RegisterForm />

              <Text ta="center" mt="md">
                {t('auth.haveAccount')}{' '}
                <Anchor component={Link} to="/login">
                  {t('auth.login')}
                </Anchor>
              </Text>
            </Paper>
          </MotionBox>
        </MotionStack>
      </Container>
    </>
  )
}
