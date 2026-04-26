import { ColorSchemeToggle } from '@/components/ColorSchemeToggle'
import { useUserPreferencesContext } from '@/contexts/UserPreferencesContext'
import { ActionIcon, Group, Menu, rem, Tooltip } from '@mantine/core'
import { IconArrowLeft, IconCheck, IconLanguage } from '@tabler/icons-react'
import { useRouter } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

interface AuthPageControlsProps {
  /** Whether to show the back button. Default: true */
  showBackButton?: boolean
}

/**
 * Controls for unauthenticated pages (login, register, legal pages).
 * Provides back navigation, theme toggle and language switcher.
 */
export function AuthPageControls({ showBackButton = true }: AuthPageControlsProps) {
  const { t } = useTranslation()
  const { language, setLanguage } = useUserPreferencesContext()
  const router = useRouter()

  const handleBack = () => {
    // Use history.back() if there's history, otherwise go to home
    if (window.history.length > 1) {
      router.history.back()
    } else {
      router.navigate({ to: '/' })
    }
  }

  return (
    <Group gap="xs" justify="center">
      {showBackButton && (
        <Tooltip label={t('nav.back')}>
          <ActionIcon
            variant="default"
            size="lg"
            aria-label={t('nav.back')}
            onClick={handleBack}
          >
            <IconArrowLeft size={20} />
          </ActionIcon>
        </Tooltip>
      )}
      <ColorSchemeToggle />
      
      <Menu shadow="md" width={150} position="bottom">
        <Menu.Target>
          <ActionIcon
            variant="default"
            size="lg"
            aria-label={t('nav.language')}
          >
            <IconLanguage size={20} />
          </ActionIcon>
        </Menu.Target>

        <Menu.Dropdown>
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
        </Menu.Dropdown>
      </Menu>
    </Group>
  )
}
