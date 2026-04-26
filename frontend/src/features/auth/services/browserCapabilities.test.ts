import { describe, it, expect, beforeEach, vi } from 'vitest'
import {
  isWebAuthnSupported,
  isConditionalMediationSupported,
  getBrowserCapabilities,
  type BrowserCapabilities,
} from './browserCapabilities'

describe('browserCapabilities', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('isWebAuthnSupported', () => {
    it('should return true when PublicKeyCredential is available', () => {
      // Mock browser with WebAuthn support
      global.PublicKeyCredential = class {} as unknown as typeof PublicKeyCredential
      
      expect(isWebAuthnSupported()).toBe(true)
    })

    it('should return false when PublicKeyCredential is not available', () => {
      // Mock browser without WebAuthn support
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      ;(global as any).PublicKeyCredential = undefined
      
      expect(isWebAuthnSupported()).toBe(false)
    })
  })

  describe('isConditionalMediationSupported', () => {
    it('should return true when conditional mediation is supported', async () => {
      // Mock browser with conditional mediation
      global.PublicKeyCredential = {
        isConditionalMediationAvailable: vi.fn().mockResolvedValue(true),
      } as unknown as typeof PublicKeyCredential
      
      const result = await isConditionalMediationSupported()
      expect(result).toBe(true)
    })

    it('should return false when conditional mediation is not supported', async () => {
      // Mock browser without conditional mediation
      global.PublicKeyCredential = {
        isConditionalMediationAvailable: vi.fn().mockResolvedValue(false),
      } as unknown as typeof PublicKeyCredential
      
      const result = await isConditionalMediationSupported()
      expect(result).toBe(false)
    })

    it('should return false when isConditionalMediationAvailable is not available', async () => {
      // Mock browser without the method
      global.PublicKeyCredential = {} as unknown as typeof PublicKeyCredential
      
      const result = await isConditionalMediationSupported()
      expect(result).toBe(false)
    })

    it('should return false when PublicKeyCredential is not available', async () => {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      ;(global as any).PublicKeyCredential = undefined
      
      const result = await isConditionalMediationSupported()
      expect(result).toBe(false)
    })

    it('should return false when isConditionalMediationAvailable throws error', async () => {
      global.PublicKeyCredential = {
        isConditionalMediationAvailable: vi.fn().mockRejectedValue(new Error('Not supported')),
      } as unknown as typeof PublicKeyCredential
      
      const result = await isConditionalMediationSupported()
      expect(result).toBe(false)
    })
  })

  describe('getBrowserCapabilities', () => {
    it('should return full capabilities for modern browser', async () => {
      global.PublicKeyCredential = {
        isConditionalMediationAvailable: vi.fn().mockResolvedValue(true),
      } as unknown as typeof PublicKeyCredential
      
      const capabilities = await getBrowserCapabilities()
      
      expect(capabilities).toEqual<BrowserCapabilities>({
        supportsWebAuthn: true,
        supportsConditionalMediation: true,
        supportsAutofill: true,
      })
    })

    it('should return partial capabilities for browser without conditional mediation', async () => {
      global.PublicKeyCredential = {
        isConditionalMediationAvailable: vi.fn().mockResolvedValue(false),
      } as unknown as typeof PublicKeyCredential
      
      const capabilities = await getBrowserCapabilities()
      
      expect(capabilities).toEqual<BrowserCapabilities>({
        supportsWebAuthn: true,
        supportsConditionalMediation: false,
        supportsAutofill: false,
      })
    })

    it('should return no capabilities for browser without WebAuthn', async () => {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      ;(global as any).PublicKeyCredential = undefined
      
      const capabilities = await getBrowserCapabilities()
      
      expect(capabilities).toEqual<BrowserCapabilities>({
        supportsWebAuthn: false,
        supportsConditionalMediation: false,
        supportsAutofill: false,
      })
    })
  })
})
