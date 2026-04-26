/**
 * Security Migration Utility
 * 
 * Handles migration from insecure query keys to secure user-specific keys
 * This ensures all existing cached data is properly cleared
 */

import { queryClient } from '@/lib/queryClient'

const SECURITY_MIGRATION_KEY = 'security_migration_complete'
const MIGRATION_VERSION = '2025-10-11-user-context-fix'

/**
 * Migrate from insecure to secure query keys
 * This function clears all potentially insecure cached data
 */
export function migrateSecurityQueryKeys() {
  const migrationComplete = localStorage.getItem(SECURITY_MIGRATION_KEY)
  
  if (migrationComplete === MIGRATION_VERSION) {
    return // Already migrated
  }
  
  // Clear all cache to be safe
  queryClient.clear()
  
  // Mark migration as complete
  localStorage.setItem(SECURITY_MIGRATION_KEY, MIGRATION_VERSION)
}

/**
 * Initialize security migration explicitly at app startup.
 * This avoids running heavy side effects automatically on module import.
 */
export function initSecurityMigration() {
  if (typeof window !== 'undefined') {
    migrateSecurityQueryKeys()
  }
}
