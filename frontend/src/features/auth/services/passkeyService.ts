/**
 * Passkey Service
 * Handles WebAuthn/Passkey registration and authentication flows
 */

import {
  startRegistration as browserStartRegistration,
  startAuthentication as browserStartAuthentication,
  browserSupportsWebAuthn,
} from '@simplewebauthn/browser'
import type {
  RegistrationResponseJSON,
  AuthenticationResponseJSON,
} from '@simplewebauthn/types'
import { apiRequest } from '@/lib/api'
import {
  PasskeyError,
  PasskeyErrorCode,
  type StartRegistrationResponse,
  type FinishRegistrationResponse,
  type StartAuthenticationResponse,
  type FinishAuthenticationResponse,
  type PasskeyCredential,
  type UpdateCredentialRequest,
} from './passkeyService.types'

/**
 * Convert WebAuthn browser errors to PasskeyError
 */
function handleWebAuthnError(error: unknown): never {
  if (error instanceof Error) {
    // User cancelled the operation
    if (error.name === 'NotAllowedError') {
      throw new PasskeyError(
        PasskeyErrorCode.USER_CANCELLED,
        'The operation was cancelled by the user',
        { originalError: error.message }
      )
    }

    // Operation timed out
    if (error.name === 'AbortError') {
      throw new PasskeyError(
        PasskeyErrorCode.TIMEOUT,
        'The operation timed out',
        { originalError: error.message }
      )
    }

    // Invalid state (e.g., credential already registered)
    if (error.name === 'InvalidStateError') {
      throw new PasskeyError(
        PasskeyErrorCode.INVALID_STATE,
        'This passkey is already registered',
        { originalError: error.message }
      )
    }

    // Generic error
    throw new PasskeyError(
      PasskeyErrorCode.UNKNOWN_ERROR,
      error.message || 'An unknown error occurred',
      { originalError: error.message }
    )
  }

  throw new PasskeyError(
    PasskeyErrorCode.UNKNOWN_ERROR,
    'An unknown error occurred',
    { originalError: String(error) }
  )
}

/**
 * Handle network/API errors
 */
function handleNetworkError(error: unknown): never {
  // If it's an API error from the backend, preserve the message and map to appropriate code
  if (error && typeof error === 'object' && 'message' in error && typeof error.message === 'string') {
    const message = error.message.toLowerCase()
    
    // Map specific backend errors to appropriate error codes
    // Generic message from BeginAuthentication when user not found or no credentials
    if (message.includes('unable to start passkey authentication') || 
        message.includes('no passkeys registered') || 
        message.includes('no credentials')) {
      throw new PasskeyError(
        PasskeyErrorCode.NO_CREDENTIALS,
        error.message,
        { originalError: error instanceof Error ? error.message : String(error) }
      )
    }
    
    if (message.includes('validation') || message.includes('invalid') || message.includes('required')) {
      throw new PasskeyError(
        PasskeyErrorCode.VALIDATION_ERROR,
        error.message,
        { originalError: error instanceof Error ? error.message : String(error) }
      )
    }
    
    // Generic API error - preserve the backend message
    throw new PasskeyError(
      PasskeyErrorCode.NETWORK_ERROR,
      error.message,
      { originalError: error instanceof Error ? error.message : String(error) }
    )
  }
  
  // Generic network error
  throw new PasskeyError(
    PasskeyErrorCode.NETWORK_ERROR,
    'Network request failed. Please check your connection and try again.',
    { originalError: error instanceof Error ? error.message : String(error) }
  )
}

class PasskeyService {
  /**
   * Start passkey registration flow
   * Fetches registration options from the server
   */
  async startRegistration(): Promise<StartRegistrationResponse> {
    try {
      return await apiRequest<StartRegistrationResponse>(
        `/auth/passkey/register/begin`,
        { method: 'POST' }
      )
    } catch (error) {
      handleNetworkError(error)
    }
  }

  /**
   * Complete passkey registration flow
   * Creates the credential with the browser and sends to server
   */
  async finishRegistration(
    options: StartRegistrationResponse,
    friendlyName?: string
  ): Promise<FinishRegistrationResponse> {
    let credential: RegistrationResponseJSON

    try {
      // Prompt user to create credential with browser/authenticator
      credential = await browserStartRegistration(options.options)
    } catch (error) {
      handleWebAuthnError(error)
    }

    try {
      // Send credential to server for verification and storage
      // sessionId and friendlyName are query params, credential is in body
      const queryParams = new URLSearchParams({
        sessionId: options.sessionId,
      })
      if (friendlyName) {
        queryParams.append('friendlyName', friendlyName)
      }

      return await apiRequest<FinishRegistrationResponse>(
        `/auth/passkey/register/finish?${queryParams.toString()}`,
        {
          method: 'POST',
          body: JSON.stringify(credential),
        }
      )
    } catch (error) {
      handleNetworkError(error)
    }
  }

  /**
   * Start passkey authentication flow
   * Fetches authentication options from the server
   */
  async startAuthentication(email?: string): Promise<StartAuthenticationResponse> {
    try {
      return await apiRequest<StartAuthenticationResponse>(
        `/auth/passkey/authenticate/begin`,
        { 
          method: 'POST',
          body: email ? JSON.stringify({ email }) : undefined,
        }
      )
    } catch (error) {
      handleNetworkError(error)
    }
  }

  /**
   * Complete passkey authentication flow
   * Gets assertion from browser and sends to server for verification
   */
  async finishAuthentication(
    options: StartAuthenticationResponse,
    useConditionalMediation = false
  ): Promise<FinishAuthenticationResponse> {
    let assertion: AuthenticationResponseJSON

    try {
      // Prompt user to authenticate with browser/authenticator
      assertion = await browserStartAuthentication(
        options.options,
        useConditionalMediation
      )
    } catch (error) {
      handleWebAuthnError(error)
    }

    try {
      // Send assertion to server for verification - sessionId as query param
      return await apiRequest<FinishAuthenticationResponse>(
        `/auth/passkey/authenticate/finish?sessionId=${options.sessionId}`,
        {
          method: 'POST',
          body: JSON.stringify(assertion),
        }
      )
    } catch (error) {
      handleNetworkError(error)
    }
  }

  /**
   * List all registered passkey credentials for the current user
   */
  async listCredentials(): Promise<PasskeyCredential[]> {
    try {
      const response = await apiRequest<{ credentials: PasskeyCredential[] }>(
        `/api/v1/passkeys`
      )
      return response.credentials
    } catch (error) {
      handleNetworkError(error)
    }
  }

  /**
   * Update a passkey credential (e.g., change friendly name)
   */
  async updateCredential(
    credentialId: string,
    updates: UpdateCredentialRequest
  ): Promise<PasskeyCredential> {
    try {
      return await apiRequest<PasskeyCredential>(
        `/api/v1/passkeys/${credentialId}/name`,
        {
          method: 'PUT',
          body: JSON.stringify(updates),
        }
      )
    } catch (error) {
      handleNetworkError(error)
    }
  }

  /**
   * Delete a passkey credential
   */
  async deleteCredential(credentialId: string): Promise<void> {
    try {
      await apiRequest<void>(
        `/api/v1/passkeys/${credentialId}`,
        { method: 'DELETE' }
      )
    } catch (error) {
      handleNetworkError(error)
    }
  }

  /**
   * Check if passkeys are supported in the current browser
   */
  isSupported(): boolean {
    return browserSupportsWebAuthn()
  }

  /**
   * Convenience method: Complete full registration flow in one call
   * Combines startRegistration + finishRegistration
   */
  async register(friendlyName?: string): Promise<FinishRegistrationResponse> {
    const options = await this.startRegistration()
    return await this.finishRegistration(options, friendlyName)
  }

  /**
   * Convenience method: Complete full authentication flow in one call
   * Combines startAuthentication + finishAuthentication
   */
  async authenticate(email: string, useConditionalMediation = false): Promise<FinishAuthenticationResponse> {
    const options = await this.startAuthentication(email)
    return await this.finishAuthentication(options, useConditionalMediation)
  }
}

// Export singleton instance
export const passkeyService = new PasskeyService()
