const API_BASE = import.meta.env.DEV ? 'http://localhost:8080' : ''

/**
 * API version prefix
 * Centralized constant for API versioning - update here to change across the entire app
 */
export const API_PREFIX = '/api/v1'

/**
 * CSRF token cookie name (must match backend)
 */
const CSRF_TOKEN_COOKIE_NAME = 'csrf_token'

/**
 * CSRF token header name (must match backend)
 */
const CSRF_TOKEN_HEADER_NAME = 'X-CSRF-Token'

/**
 * Get the CSRF token from the cookie
 * Returns an empty string if the cookie is not found
 */
export function getCSRFToken(): string {
  if (typeof document === 'undefined') {
    return ''
  }
  const cookie = document.cookie
    .split('; ')
    .find(row => row.startsWith(`${CSRF_TOKEN_COOKIE_NAME}=`))
  if (!cookie) return ''
  // Use substring instead of split('=')[1] to preserve '=' characters in base64 values
  return cookie.substring(CSRF_TOKEN_COOKIE_NAME.length + 1)
}

export class ApiError extends Error {
  status: number
  code: string
  details?: Record<string, unknown>

  constructor(
    status: number,
    code: string,
    message: string,
    details?: Record<string, unknown>
  ) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
    this.details = details
  }
}

export async function apiRequest<T>(
  endpoint: string,
  options?: RequestInit
): Promise<T> {
  // Build headers with CSRF token for non-GET requests
  const method = options?.method?.toUpperCase() || 'GET'
  const headers: Record<string, string> = {}

  // Only set Content-Type for non-FormData bodies (browser sets it automatically for FormData)
  if (!(options?.body instanceof FormData)) {
    headers['Content-Type'] = 'application/json'
  }
  
  // Include CSRF token for state-changing requests
  if (method !== 'GET' && method !== 'HEAD' && method !== 'OPTIONS') {
    const csrfToken = getCSRFToken()
    if (csrfToken) {
      headers[CSRF_TOKEN_HEADER_NAME] = csrfToken
    }
  }

  const response = await fetch(`${API_BASE}${endpoint}`, {
    ...options,
    credentials: 'include', // Send cookies with requests
    headers: {
      ...headers,
      ...options?.headers,
    },
  })
  
  if (!response.ok) {
    let error
    try {
      error = await response.json()
    } catch (e) {
      // Response body is not valid JSON (e.g., from a panic/crash)
      throw new ApiError(
        response.status,
        'INVALID_RESPONSE',
        `Server returned ${response.status} with invalid JSON response`,
        { originalError: e instanceof Error ? e.message : String(e) }
      )
    }
    throw new ApiError(
      response.status,
      error.code || 'UNKNOWN_ERROR',
      error.message || 'An error occurred',
      error.details
    )
  }
  
  if (response.status === 204) {
    return null as T
  }
  
  const json = await response.json()
  // Backend wraps successful responses in { data: ... }
  return json.data !== undefined ? json.data : json
}

/**
 * Download a file from an API endpoint
 * Handles authentication and triggers browser download
 */
export async function apiDownload(
  endpoint: string,
  defaultFilename: string = 'download'
): Promise<void> {
  const response = await fetch(`${API_BASE}${endpoint}`, {
    method: 'GET',
    credentials: 'include', // Send cookies with requests
  })
  
  if (!response.ok) {
    let error
    try {
      error = await response.json()
    } catch {
      throw new ApiError(
        response.status,
        'DOWNLOAD_FAILED',
        `Download failed: ${response.statusText}`
      )
    }
    throw new ApiError(
      response.status,
      error.code || 'DOWNLOAD_FAILED',
      error.message || 'Download failed',
      error.details
    )
  }
  
  // Get filename from Content-Disposition header
  const contentDisposition = response.headers.get('content-disposition')
  let filename = defaultFilename
  
  if (contentDisposition) {
    const filenameMatch = contentDisposition.match(/filename="?(.+)"?/)
    if (filenameMatch) {
      filename = filenameMatch[1]
    }
  }
  
  // Get blob and trigger download
  const blob = await response.blob()
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  link.style.display = 'none'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}

/**
 * Delete user account and all associated data
 * Requires email confirmation for security
 */
export async function deleteUserAccount(confirmEmail: string): Promise<void> {
  return apiRequest(`${API_PREFIX}/user`, {
    method: 'DELETE',
    body: JSON.stringify({ confirmEmail }),
  })
}
