import { describe, it, expect, beforeEach, vi, type Mock } from 'vitest'
import * as SimpleWebAuthnBrowser from '@simplewebauthn/browser'
import type {
  RegistrationResponseJSON,
  AuthenticationResponseJSON,
  PublicKeyCredentialRequestOptionsJSON,
} from '@simplewebauthn/types'
import { passkeyService } from './passkeyService'
import { PasskeyError, PasskeyErrorCode } from './passkeyService.types'
import type {
  StartRegistrationResponse,
  StartAuthenticationResponse,
  PasskeyCredential,
} from './passkeyService.types'

// Mock the API
vi.mock('@/lib/api', () => ({
  apiRequest: vi.fn(),
  API_PREFIX: '/api/v1',
}))

// Mock SimpleWebAuthn browser library
vi.mock('@simplewebauthn/browser', () => ({
  startRegistration: vi.fn(),
  startAuthentication: vi.fn(),
  browserSupportsWebAuthn: vi.fn(),
}))

import { apiRequest } from '@/lib/api'

const mockApiRequest = apiRequest as Mock

describe('passkeyService', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('startRegistration', () => {
    it('should fetch registration options from API', async () => {
      const mockOptions: StartRegistrationResponse = {
        sessionId: 'test-session-id',
        options: {
          challenge: 'test-challenge',
          rp: { name: 'Go React Starter', id: 'localhost' },
          user: {
            id: 'user-id',
            name: 'test@example.com',
            displayName: 'Test User',
          },
          pubKeyCredParams: [
            { alg: -7, type: 'public-key' },
            { alg: -257, type: 'public-key' },
          ],
          timeout: 60000,
          attestation: 'none',
          authenticatorSelection: {
            residentKey: 'preferred',
            userVerification: 'preferred',
          },
        },
      }

      mockApiRequest.mockResolvedValue(mockOptions)

      const result = await passkeyService.startRegistration()

      expect(mockApiRequest).toHaveBeenCalledWith('/auth/passkey/register/begin', {
        method: 'POST',
      })
      expect(result).toEqual(mockOptions)
    })

    it('should throw PasskeyError on network failure', async () => {
      mockApiRequest.mockRejectedValue(new Error('Network error'))

      await expect(passkeyService.startRegistration()).rejects.toThrow(PasskeyError)
      await expect(passkeyService.startRegistration()).rejects.toMatchObject({
        code: PasskeyErrorCode.NETWORK_ERROR,
      })
    })
  })

  describe('finishRegistration', () => {
    it('should complete registration with browser credential', async () => {
      const mockOptions: StartRegistrationResponse = {
        sessionId: 'session-123',
        options: {
          challenge: 'challenge',
          rp: { name: 'Go React Starter', id: 'localhost' },
          user: { id: 'user-id', name: 'test@example.com', displayName: 'Test' },
          pubKeyCredParams: [{ alg: -7, type: 'public-key' }],
        },
      }

      const mockCredential = {
        id: 'credential-id',
        rawId: 'credential-id',
        response: {
          clientDataJSON: 'client-data',
          attestationObject: 'attestation',
        },
        type: 'public-key',
      }

      const mockFinishResponse = {
        credentialId: 'credential-id',
        friendlyName: 'My Device',
        createdAt: '2025-11-11T17:00:00Z',
      }

      vi.mocked(SimpleWebAuthnBrowser.startRegistration).mockResolvedValue(mockCredential as RegistrationResponseJSON)
      mockApiRequest.mockResolvedValue(mockFinishResponse)

      const result = await passkeyService.finishRegistration(mockOptions, 'My Device')

      expect(SimpleWebAuthnBrowser.startRegistration).toHaveBeenCalledWith(mockOptions.options)
      expect(mockApiRequest).toHaveBeenCalledWith('/auth/passkey/register/finish?sessionId=session-123&friendlyName=My+Device', {
        method: 'POST',
        body: JSON.stringify(mockCredential),
      })
      expect(result).toEqual(mockFinishResponse)
    })

    it('should throw PasskeyError when user cancels', async () => {
      const mockOptions: StartRegistrationResponse = {
        sessionId: 'session-123',
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        options: {} as any,
      }

      const cancelError = new Error('User cancelled')
      cancelError.name = 'NotAllowedError'
      vi.mocked(SimpleWebAuthnBrowser.startRegistration).mockRejectedValue(cancelError)

      await expect(passkeyService.finishRegistration(mockOptions)).rejects.toThrow(PasskeyError)
      await expect(passkeyService.finishRegistration(mockOptions)).rejects.toMatchObject({
        code: PasskeyErrorCode.USER_CANCELLED,
      })
    })

    it('should throw PasskeyError on timeout', async () => {
      const mockOptions: StartRegistrationResponse = {
        sessionId: 'session-123',
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        options: {} as any,
      }

      const timeoutError = new Error('Timeout')
      timeoutError.name = 'AbortError'
      vi.mocked(SimpleWebAuthnBrowser.startRegistration).mockRejectedValue(timeoutError)

      await expect(passkeyService.finishRegistration(mockOptions)).rejects.toThrow(PasskeyError)
      await expect(passkeyService.finishRegistration(mockOptions)).rejects.toMatchObject({
        code: PasskeyErrorCode.TIMEOUT,
      })
    })
  })

  describe('startAuthentication', () => {
    it('should fetch authentication options from API', async () => {
      const mockOptions: StartAuthenticationResponse = {
        sessionId: 'auth-session',
        options: {
          challenge: 'auth-challenge',
          timeout: 60000,
          rpId: 'localhost',
          allowCredentials: [],
          userVerification: 'preferred',
        },
      }

      mockApiRequest.mockResolvedValue(mockOptions)

      const result = await passkeyService.startAuthentication()

      expect(mockApiRequest).toHaveBeenCalledWith('/auth/passkey/authenticate/begin', {
        method: 'POST',
      })
      expect(result).toEqual(mockOptions)
    })

    it('should throw PasskeyError on network failure', async () => {
      mockApiRequest.mockRejectedValue(new Error('Network error'))

      await expect(passkeyService.startAuthentication()).rejects.toThrow(PasskeyError)
      await expect(passkeyService.startAuthentication()).rejects.toMatchObject({
        code: PasskeyErrorCode.NETWORK_ERROR,
      })
    })
  })

  describe('finishAuthentication', () => {
    it('should complete authentication and return token', async () => {
      const mockOptions: StartAuthenticationResponse = {
        sessionId: 'auth-session',
        options: {
          challenge: 'challenge',
          rpId: 'localhost',
        } as PublicKeyCredentialRequestOptionsJSON,
      }

      const mockAssertion = {
        id: 'credential-id',
        rawId: 'credential-id',
        response: {
          clientDataJSON: 'client-data',
          authenticatorData: 'auth-data',
          signature: 'signature',
        },
        type: 'public-key',
      }

      const mockAuthResponse = {
        token: 'jwt-token',
        user: {
          id: 'user-id',
          email: 'test@example.com',
          firstName: 'Test',
          lastName: 'User',
        },
      }

      vi.mocked(SimpleWebAuthnBrowser.startAuthentication).mockResolvedValue(mockAssertion as AuthenticationResponseJSON)
      mockApiRequest.mockResolvedValue(mockAuthResponse)

      const result = await passkeyService.finishAuthentication(mockOptions)

      expect(SimpleWebAuthnBrowser.startAuthentication).toHaveBeenCalledWith(mockOptions.options, false)
      expect(mockApiRequest).toHaveBeenCalledWith('/auth/passkey/authenticate/finish?sessionId=auth-session', {
        method: 'POST',
        body: JSON.stringify(mockAssertion),
      })
      expect(result).toEqual(mockAuthResponse)
    })

    it('should throw PasskeyError when user cancels', async () => {
      const mockOptions: StartAuthenticationResponse = {
        sessionId: 'auth-session',
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        options: {} as any,
      }

      const cancelError = new Error('User cancelled')
      cancelError.name = 'NotAllowedError'
      vi.mocked(SimpleWebAuthnBrowser.startAuthentication).mockRejectedValue(cancelError)

      await expect(passkeyService.finishAuthentication(mockOptions)).rejects.toThrow(PasskeyError)
      await expect(passkeyService.finishAuthentication(mockOptions)).rejects.toMatchObject({
        code: PasskeyErrorCode.USER_CANCELLED,
      })
    })
  })

  describe('listCredentials', () => {
    it('should fetch user credentials from API', async () => {
      const mockCredentials: PasskeyCredential[] = [
        {
          id: 'cred-1',
          friendlyName: 'iPhone 15 Pro',
          authenticatorAttachment: 'platform',
          backupEligible: true,
          backupState: true,
          createdAt: '2025-11-01T10:00:00Z',
          lastUsedAt: '2025-11-11T16:00:00Z',
        },
        {
          id: 'cred-2',
          friendlyName: 'YubiKey 5',
          authenticatorAttachment: 'cross-platform',
          backupEligible: false,
          backupState: false,
          createdAt: '2025-11-05T14:00:00Z',
        },
      ]

      mockApiRequest.mockResolvedValue({ credentials: mockCredentials })

      const result = await passkeyService.listCredentials()

      expect(mockApiRequest).toHaveBeenCalledWith('/api/v1/passkeys')
      expect(result).toEqual(mockCredentials)
    })
  })

  describe('updateCredential', () => {
    it('should update credential friendly name', async () => {
      const credentialId = 'cred-1'
      const newName = 'My MacBook Pro'
      
      const mockUpdatedCredential: PasskeyCredential = {
        id: credentialId,
        friendlyName: newName,
        authenticatorAttachment: 'platform',
        backupEligible: true,
        backupState: true,
        createdAt: '2025-11-01T10:00:00Z',
      }

      mockApiRequest.mockResolvedValue(mockUpdatedCredential)

      const result = await passkeyService.updateCredential(credentialId, { friendlyName: newName })

      expect(mockApiRequest).toHaveBeenCalledWith(`/api/v1/passkeys/${credentialId}/name`, {
        method: 'PUT',
        body: JSON.stringify({ friendlyName: newName }),
      })
      expect(result).toEqual(mockUpdatedCredential)
    })
  })

  describe('deleteCredential', () => {
    it('should delete credential', async () => {
      const credentialId = 'cred-1'

      mockApiRequest.mockResolvedValue(null)

      await passkeyService.deleteCredential(credentialId)

      expect(mockApiRequest).toHaveBeenCalledWith(`/api/v1/passkeys/${credentialId}`, {
        method: 'DELETE',
      })
    })
  })

  describe('isSupported', () => {
    it('should return true when browserSupportsWebAuthn returns true', () => {
      vi.mocked(SimpleWebAuthnBrowser.browserSupportsWebAuthn).mockReturnValue(true)

      expect(passkeyService.isSupported()).toBe(true)
    })

    it('should return false when browserSupportsWebAuthn returns false', () => {
      vi.mocked(SimpleWebAuthnBrowser.browserSupportsWebAuthn).mockReturnValue(false)

      expect(passkeyService.isSupported()).toBe(false)
    })
  })
})
