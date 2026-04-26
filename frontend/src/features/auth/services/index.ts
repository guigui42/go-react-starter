/**
 * Passkey authentication services
 */

export { passkeyService } from './passkeyService'
export {
  isWebAuthnSupported,
  isConditionalMediationSupported,
  getBrowserCapabilities,
  type BrowserCapabilities,
} from './browserCapabilities'
export type {
  StartRegistrationResponse,
  FinishRegistrationRequest,
  FinishRegistrationResponse,
  StartAuthenticationResponse,
  FinishAuthenticationRequest,
  FinishAuthenticationResponse,
  PasskeyCredential,
  UpdateCredentialRequest,
} from './passkeyService.types'
export {
  PasskeyError,
  PasskeyErrorCode,
} from './passkeyService.types'
