/**
 * User preferences types
 */

export type ColorScheme = 'light' | 'dark' | 'auto'

export type DigestFrequency = 'never' | 'daily' | 'weekly' | 'monthly'

export interface UserPreferences {
  id: string
  user_id: string
  language: 'en' | 'fr'
  base_currency: string
  country_code: string
  color_scheme: ColorScheme
  has_completed_onboarding: boolean
  digest_frequency: DigestFrequency
  created_at: string
  updated_at: string
}

export interface UpdateUserPreferencesRequest {
  language?: 'en' | 'fr'
  base_currency?: string
  country_code?: string
  color_scheme?: ColorScheme
  has_completed_onboarding?: boolean
  digest_frequency?: DigestFrequency
}
