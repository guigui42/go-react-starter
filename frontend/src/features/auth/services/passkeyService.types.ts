/**
 * Type definitions for passkey service
 */

import type {
  PublicKeyCredentialCreationOptionsJSON,
  PublicKeyCredentialRequestOptionsJSON,
  RegistrationResponseJSON,
  AuthenticationResponseJSON,
} from '@simplewebauthn/types'

/**
 * Registration flow types
 */
export interface StartRegistrationResponse {
  options: PublicKeyCredentialCreationOptionsJSON
  sessionId: string
}

export interface FinishRegistrationRequest {
  sessionId: string
  credential: RegistrationResponseJSON
  friendlyName?: string
}

export interface FinishRegistrationResponse {
  credentialId: string
  friendlyName: string
  createdAt: string
}

/**
 * Authentication flow types
 */
export interface StartAuthenticationResponse {
  options: PublicKeyCredentialRequestOptionsJSON
  sessionId: string
}

export interface FinishAuthenticationRequest {
  sessionId: string
  credential: AuthenticationResponseJSON
}

export interface FinishAuthenticationResponse {
  token: string
  user: {
    id: string
    email: string
    firstName: string
    lastName: string
  }
}

/**
 * Credential management types
 */
export interface PasskeyCredential {
  id: string
  friendlyName: string
  authenticatorAttachment: 'platform' | 'cross-platform'
  backupEligible: boolean
  backupState: boolean
  createdAt: string
  lastUsedAt?: string
}

export interface UpdateCredentialRequest {
  friendlyName: string
}

/**
 * Error types for passkey operations
 */
export class PasskeyError extends Error {
  code: string
  details?: Record<string, unknown>

  constructor(code: string, message: string, details?: Record<string, unknown>) {
    super(message)
    this.name = 'PasskeyError'
    this.code = code
    this.details = details
  }
}

/**
 * Error codes
 */
export const PasskeyErrorCode = {
  NOT_SUPPORTED: 'NOT_SUPPORTED',
  USER_CANCELLED: 'USER_CANCELLED',
  TIMEOUT: 'TIMEOUT',
  INVALID_STATE: 'INVALID_STATE',
  NETWORK_ERROR: 'NETWORK_ERROR',
  NO_CREDENTIALS: 'NO_CREDENTIALS',
  VALIDATION_ERROR: 'VALIDATION_ERROR',
  UNKNOWN_ERROR: 'UNKNOWN_ERROR',
} as const
