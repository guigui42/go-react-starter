/**
 * Auth event utilities for consistent authentication state management
 * 
 * With httpOnly cookies, tokens are managed by the browser automatically.
 * This module provides event-based signaling for authentication state changes.
 */

/**
 * Event name for auth state changes
 */
export const AUTH_STATE_CHANGED_EVENT = 'auth-state-changed';

/**
 * Legacy event name for backward compatibility
 * @deprecated Use AUTH_STATE_CHANGED_EVENT instead
 */
export const AUTH_TOKEN_CHANGED_EVENT = AUTH_STATE_CHANGED_EVENT;

/**
 * Trigger auth state change event
 * Use this after login/logout to notify components of authentication changes
 */
export function triggerAuthStateChanged(): void {
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new Event(AUTH_STATE_CHANGED_EVENT));
  }
}

/**
 * Legacy function for backward compatibility
 * @deprecated Use triggerAuthStateChanged instead
 */
export function triggerAuthTokenChanged(): void {
  triggerAuthStateChanged();
}

/**
 * Notify that user has logged in
 * Called after successful login/register
 */
export function notifyLogin(): void {
  triggerAuthStateChanged();
}

/**
 * Notify that user has logged out
 * Called after logout
 */
export function notifyLogout(): void {
  triggerAuthStateChanged();
}

// Legacy functions - kept for backward compatibility during migration
// These are no-ops since we no longer use localStorage for tokens

/**
 * @deprecated Tokens are now managed via httpOnly cookies
 */
export function setAuthToken(_token: string): void {
  // No-op: tokens are now set via httpOnly cookies by the backend
  triggerAuthStateChanged();
}

/**
 * @deprecated Tokens are now managed via httpOnly cookies
 */
export function removeAuthToken(): void {
  // No-op: tokens are now cleared via httpOnly cookies by the backend
  triggerAuthStateChanged();
}

/**
 * @deprecated Tokens are now managed via httpOnly cookies
 * Always returns null since we cannot access httpOnly cookies from JavaScript
 */
export function getAuthToken(): string | null {
  // Cannot access httpOnly cookies from JavaScript
  // Return null - authentication status should be determined by API calls
  return null;
}