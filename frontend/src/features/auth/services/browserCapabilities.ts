/**
 * Browser capability detection for WebAuthn/Passkey features
 */

export interface BrowserCapabilities {
  supportsWebAuthn: boolean
  supportsConditionalMediation: boolean
  supportsAutofill: boolean
}

/**
 * Check if the browser supports WebAuthn
 */
export function isWebAuthnSupported(): boolean {
  return typeof window !== 'undefined' && 
         typeof window.PublicKeyCredential !== 'undefined'
}

/**
 * Check if the browser supports conditional mediation (autofill UI)
 * Supported in Chrome 108+, Safari 16+, Edge 108+
 */
export async function isConditionalMediationSupported(): Promise<boolean> {
  if (!isWebAuthnSupported()) {
    return false
  }

  try {
    if (typeof window.PublicKeyCredential.isConditionalMediationAvailable === 'function') {
      return await window.PublicKeyCredential.isConditionalMediationAvailable()
    }
    return false
  } catch {
    return false
  }
}

/**
 * Get comprehensive browser capabilities for passkey features
 */
export async function getBrowserCapabilities(): Promise<BrowserCapabilities> {
  const supportsWebAuthn = isWebAuthnSupported()
  const supportsConditionalMediation = supportsWebAuthn 
    ? await isConditionalMediationSupported() 
    : false

  return {
    supportsWebAuthn,
    supportsConditionalMediation,
    supportsAutofill: supportsConditionalMediation, // Autofill requires conditional mediation
  }
}
