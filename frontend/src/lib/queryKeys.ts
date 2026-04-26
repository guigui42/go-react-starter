/**
 * Centralized Query Key Factory for TanStack Query
 */

export const queryKeys = {
  auth: {
    all: (userId?: string) => {
      return userId ? [userId, 'auth'] as const : ['auth'] as const;
    },
    user: () => ['auth', 'user'] as const,
    passkeys: () => ['auth', 'passkeys'] as const,
    passkeysList: () => [...queryKeys.auth.passkeys(), 'list'] as const,
    migration: () => ['auth', 'migration'] as const,
    migrationStatus: () => [...queryKeys.auth.migration(), 'status'] as const,
  },

  notes: {
    all: (userId: string) => [userId, 'notes'] as const,
    lists: (userId: string) => [...queryKeys.notes.all(userId), 'list'] as const,
    list: (userId: string) => [...queryKeys.notes.lists(userId)] as const,
    details: (userId: string) => [...queryKeys.notes.all(userId), 'detail'] as const,
    detail: (userId: string, id: string) => [...queryKeys.notes.details(userId), id] as const,
  },

  preferences: {
    all: (userId: string) => [userId, 'preferences'] as const,
    user: (userId: string) => [...queryKeys.preferences.all(userId), 'user'] as const,
  },

  oauth: {
    providers: () => ['oauth', 'providers'] as const,
    accounts: (userId: string) => [userId, 'oauth', 'accounts'] as const,
  },

  admin: {
    all: ['admin'] as const,
    stats: () => [...queryKeys.admin.all, 'stats'] as const,
    users: () => [...queryKeys.admin.all, 'users'] as const,
    logs: () => [...queryKeys.admin.all, 'logs'] as const,
    emailConfig: () => [...queryKeys.admin.all, 'email', 'config'] as const,
    auditLogs: () => [...queryKeys.admin.all, 'audit-logs'] as const,
    migrations: () => [...queryKeys.admin.all, 'migrations'] as const,
  },
} as const;

export type QueryKeys = typeof queryKeys;
export type AuthKeys = typeof queryKeys.auth;
export type NoteKeys = typeof queryKeys.notes;
export type PreferenceKeys = typeof queryKeys.preferences;
export type AdminKeys = typeof queryKeys.admin;
export type OAuthKeys = typeof queryKeys.oauth;
