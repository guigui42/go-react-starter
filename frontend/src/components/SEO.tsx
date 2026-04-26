import { useTranslation } from 'react-i18next'

const SITE_URL = 'http://localhost:5173'

interface SEOProps {
  translationKey: string
  path: string
}

/**
 * SEO component using React 19 native meta tag support.
 * Automatically hoists <title> and <meta> to <head>.
 */
export function SEO({ translationKey, path }: SEOProps) {
  const { t } = useTranslation()
  
  const title = t(`seo.${translationKey}.title`)
  const description = t(`seo.${translationKey}.description`)
  const canonicalUrl = `${SITE_URL}${path}`
  
  return (
    <>
      <title>{title}</title>
      <meta name="description" content={description} />
      <link rel="canonical" href={canonicalUrl} />
    </>
  )
}
