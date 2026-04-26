import { Anchor, Group, Text } from '@mantine/core'
import { Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import classes from './Footer.module.css'

/**
 * Footer component displayed at the bottom of all pages.
 * Contains copyright, company info, and links to legal pages.
 */
export function Footer() {
  const { t } = useTranslation()
  const currentYear = new Date().getFullYear()

  return (
    <footer className={classes.footer} role="contentinfo">
      <nav aria-label={t('footer.navigation')}>
        <Group gap="xs" justify="center" wrap="wrap">
          <Anchor component={Link} to="/privacy" size="sm" c="dimmed">
            {t('footer.privacy')}
          </Anchor>
          <Text size="sm" c="dimmed">•</Text>
          <Anchor component={Link} to="/terms" size="sm" c="dimmed">
            {t('footer.terms')}
          </Anchor>
          <Text size="sm" c="dimmed">•</Text>
          <Anchor component={Link} to="/legal" size="sm" c="dimmed">
            {t('footer.legal')}
          </Anchor>
        </Group>
      </nav>
      <Text size="sm" c="dimmed" ta="center" mt="xs">
        © {currentYear} {t('footer.company')} {t('footer.rights')}
      </Text>
    </footer>
  )
}
