/// <reference types="vite/client" />
/// <reference types="vite-plugin-pwa/react" />

interface ImportMetaEnv {
  /** API base URL (e.g., http://localhost:8080) */
  readonly VITE_API_BASE_URL: string
  /** Environment name (development, production) - legacy */
  readonly VITE_ENV: string
  /** Application environment (local-dev, azure-dev, prod) */
  readonly VITE_APP_ENV: string
  /** Application version (release tag for prod, 'dev' for others) */
  readonly VITE_APP_VERSION: string
  /** Git commit SHA */
  readonly VITE_GIT_SHA: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
