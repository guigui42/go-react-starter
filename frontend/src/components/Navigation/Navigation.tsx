import { ColorSchemeToggle } from '@/components/ColorSchemeToggle'
import { useAuth } from '@/contexts/AuthContext'
import { useUserPreferencesContext } from '@/contexts/UserPreferencesContext'
import {
  Avatar,
  Box,
  Burger,
  Group,
  Menu,
  NavLink,
  Paper,
  rem,
  Text,
  UnstyledButton,
} from '@mantine/core'
import {
  IconCheck,
  IconChevronDown,
  IconHome,
  IconLanguage,
  IconLogout,
  IconNote,
  IconSettings,
  IconShieldCheck,
} from '@tabler/icons-react'
import { Link, useLocation } from '@tanstack/react-router'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import classes from './Navigation.module.css'

interface NavigationLink {
  icon: typeof IconHome
  label: string
  href: string
  translationKey: string
}

const navigationLinks: NavigationLink[] = [
  {
    icon: IconHome,
    label: 'Home',
    href: '/',
    translationKey: 'nav.home',
  },
  {
    icon: IconNote,
    label: 'Notes',
    href: '/notes',
    translationKey: 'nav.notes',
  },
]

export function Navigation() {
  const [opened, setOpened] = useState(false)
  const { t } = useTranslation()
  const { isAuthenticated, user, logout, isLoading } = useAuth()
  const { language, setLanguage } = useUserPreferencesContext()
  const location = useLocation()

  // Don't show navigation while loading auth state or when not authenticated
  if (isLoading || !isAuthenticated) {
    return null
  }

  const handleLogout = async () => {
    try {
      await logout()
      window.location.href = '/login'
    } catch {
      window.location.href = '/login'
    }
  }

  return (
    <Paper component="header" className={classes.header} role="banner">
      <Group h="100%" px="md" justify="space-between" wrap="nowrap">
        {/* Logo and Burger */}
        <Group style={{ flex: 1 }}>
          <Burger
            opened={opened}
            onClick={() => setOpened((o) => !o)}
            hiddenFrom="sm"
            size="sm"
            aria-label={t('nav.toggleNavigation') || 'Toggle navigation'}
          />
          <Link to="/" className={classes.logo}>
            <Text size="xl" fw={700}>
              Go React Starter
            </Text>
          </Link>
        </Group>

        {/* Desktop Navigation Links */}
        <Group gap="xs" visibleFrom="sm" component="nav" aria-label="Main navigation">
          {navigationLinks.map((link) => {
            const Icon = link.icon
            const isActive = location.pathname === link.href

            return (
              <Link
                key={link.href}
                to={link.href}
                className={classes.link}
                data-active={isActive || undefined}
              >
                <Icon style={{ width: rem(16), height: rem(16) }} stroke={1.5} />
                <span>{t(link.translationKey)}</span>
              </Link>
            )
          })}
        </Group>

        {/* User Menu and Theme Toggle */}
        <Group gap="sm" style={{ flex: 1 }} justify="flex-end" wrap="nowrap">
          <ColorSchemeToggle />

          <Menu shadow="md" width={200} position="bottom-end">
            <Menu.Target>
              <UnstyledButton className={classes.user}>
                <Group gap={7}>
                  <Avatar size={30} radius="xl" color="blue">
                    {user?.email?.charAt(0).toUpperCase()}
                  </Avatar>
                  <Box visibleFrom="sm">
                    <Text size="sm" fw={500}>
                      {user?.email}
                    </Text>
                  </Box>
                  <IconChevronDown
                    style={{ width: rem(12), height: rem(12) }}
                    stroke={1.5}
                  />
                </Group>
              </UnstyledButton>
            </Menu.Target>

            <Menu.Dropdown>
              <Menu.Label>{t('nav.account') || 'Account'}</Menu.Label>
              <Menu.Item
                component={Link}
                to="/settings"
                leftSection={
                  <IconSettings style={{ width: rem(14), height: rem(14) }} />
                }
              >
                {t('nav.settings')}
              </Menu.Item>

              {/* Admin link - only shown for admin users */}
              {user?.is_admin && (
                <Menu.Item
                  component={Link}
                  to="/admin"
                  leftSection={
                    <IconShieldCheck style={{ width: rem(14), height: rem(14) }} />
                  }
                  color="violet"
                >
                  {t('nav.admin') || 'Admin Dashboard'}
                </Menu.Item>
              )}

              <Menu.Divider />

              <Menu.Label>{t('nav.language')}</Menu.Label>
              <Menu.Item
                leftSection={
                  <IconLanguage style={{ width: rem(14), height: rem(14) }} />
                }
                rightSection={
                  language === 'en' ? (
                    <IconCheck style={{ width: rem(14), height: rem(14) }} />
                  ) : null
                }
                onClick={() => setLanguage('en')}
              >
                {t('nav.languages.en')}
              </Menu.Item>
              <Menu.Item
                leftSection={
                  <IconLanguage style={{ width: rem(14), height: rem(14) }} />
                }
                rightSection={
                  language === 'fr' ? (
                    <IconCheck style={{ width: rem(14), height: rem(14) }} />
                  ) : null
                }
                onClick={() => setLanguage('fr')}
              >
                {t('nav.languages.fr')}
              </Menu.Item>

              <Menu.Divider />

              <Menu.Item
                color="red"
                leftSection={
                  <IconLogout style={{ width: rem(14), height: rem(14) }} />
                }
                onClick={handleLogout}
              >
                {t('auth.logout')}
              </Menu.Item>
            </Menu.Dropdown>
          </Menu>
        </Group>
      </Group>

      {/* Mobile Navigation Drawer */}
      {opened && (
        <Box className={classes.mobileNav} hiddenFrom="sm">
          <nav aria-label="Main navigation">
            {navigationLinks.map((link) => {
              const Icon = link.icon
              const isActive = location.pathname === link.href

              return (
                <NavLink
                  key={link.href}
                  component={Link}
                  to={link.href}
                  label={t(link.translationKey)}
                  leftSection={<Icon size="1rem" stroke={1.5} />}
                  active={isActive}
                  onClick={() => setOpened(false)}
                />
              )
            })}
          </nav>
        </Box>
      )}
    </Paper>
  )
}
