import { Footer } from '@/components/Footer'
import { Navigation } from '@/components/Navigation'
import { Box, Button, Center, Container, Stack, Text, Title } from '@mantine/core'
import { IconArrowLeft, IconHome } from '@tabler/icons-react'
import { createRootRouteWithContext, Link, Outlet, useMatches, useRouterState } from '@tanstack/react-router'
import { motion } from 'framer-motion'
import { useEffect } from 'react'
import { useTranslation } from 'react-i18next'

interface RouterContext {
  auth: {
    isAuthenticated: boolean
    isLoading: boolean
  }
}

// Routes that should hide main navigation (public/legal pages)
const PUBLIC_ROUTES = ['/', '/login', '/register']

/**
 * SEO-optimized 404 Not Found page
 * - Uses meta robots noindex to prevent indexing of 404 pages
 * - Provides clear navigation back to main content
 * - Supports i18n for French/English
 */
function NotFoundComponent() {
  const { t } = useTranslation()

  useEffect(() => {
    const metaRobots = document.createElement('meta')
    metaRobots.name = 'robots'
    metaRobots.content = 'noindex, nofollow'
    document.head.appendChild(metaRobots)

    document.title = `${t('errors.pageNotFound', 'Page Not Found')} - Go React Starter`

    return () => {
      document.head.removeChild(metaRobots)
    }
  }, [t])

  return (
    <Box style={{ display: 'flex', flexDirection: 'column', minHeight: '100dvh' }}>
      <Container size="sm" py={80}>
        <Center>
          <Stack align="center" gap="xl">
            <Title
              order={1}
              size={120}
              fw={900}
              c="dimmed"
              style={{ lineHeight: 1 }}
            >
              404
            </Title>

            <Stack align="center" gap="sm">
              <Title order={2} ta="center">
                {t('errors.pageNotFound', 'Page Not Found')}
              </Title>
              <Text c="dimmed" ta="center" maw={400}>
                {t('errors.pageNotFoundDescription', "The page you're looking for doesn't exist or has been moved.")}
              </Text>
            </Stack>

            <Stack gap="sm" w="100%" maw={300}>
              <Button
                component={Link}
                to="/"
                size="lg"
                leftSection={<IconHome size={20} />}
                fullWidth
              >
                {t('errors.goHome', 'Go to Homepage')}
              </Button>

              <Button
                variant="light"
                size="lg"
                leftSection={<IconArrowLeft size={20} />}
                onClick={() => window.history.back()}
                fullWidth
              >
                {t('errors.goBack', 'Go Back')}
              </Button>
            </Stack>
          </Stack>
        </Center>
      </Container>
      <Footer />
    </Box>
  )
}

function RootComponent() {
  const matches = useMatches()
  const { i18n } = useTranslation()
  const { location } = useRouterState()

  useEffect(() => {
    document.documentElement.lang = i18n.language
  }, [i18n.language])

  const isAuthRoute = matches.some(match => match.pathname.startsWith('/auth'))

  const currentPath = matches.length > 0 ? matches[matches.length - 1].pathname : ''
  const isPublicRoute = PUBLIC_ROUTES.some(route =>
    currentPath === route || currentPath.startsWith(`${route}/`)
  )

  const hideNavigation = isAuthRoute || isPublicRoute

  return (
    <Box style={{ display: 'flex', flexDirection: 'column', minHeight: '100dvh' }}>
      {!hideNavigation && <Navigation />}
      <Container
        size="xl"
        py="md"
        px={{ base: 'xs', sm: 'md' }}
        style={{
          marginTop: hideNavigation ? 0 : undefined,
          flex: 1,
          width: '100%',
        }}
      >
        <motion.div
          key={location.pathname}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.15, ease: 'easeOut' }}
        >
          <Outlet />
        </motion.div>
      </Container>
      <Footer />
    </Box>
  )
}

export const Route = createRootRouteWithContext<RouterContext>()({
  component: RootComponent,
  notFoundComponent: NotFoundComponent,
})
