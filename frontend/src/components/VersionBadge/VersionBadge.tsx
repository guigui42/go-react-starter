import { useAuth } from '@/contexts/AuthContext';
import { Anchor, Badge, Group, Text } from '@mantine/core';
import classes from './VersionBadge.module.css';

/**
 * Version information injected at build time via Vite environment variables.
 * 
 * Environment detection:
 * - local-dev: Local development (default)
 * - azure-dev: Azure development environment
 * - prod: Azure production environment
 * 
 * Version display:
 * - Local/Dev: Shows git commit SHA
 * - Prod: Shows release version (YYYY.MM.DD.XXXX format)
 * 
 * Visibility:
 * - Production: Only visible to admin users
 * - Other environments: Visible to everyone
 */
interface VersionInfo {
  env: string;
  version: string;
  gitSha: string;
}

function getVersionInfo(): VersionInfo {
  return {
    env: import.meta.env.VITE_APP_ENV || 'local-dev',
    version: import.meta.env.VITE_APP_VERSION || 'dev',
    gitSha: import.meta.env.VITE_GIT_SHA || 'unknown',
  };
}

function getEnvColor(env: string): string {
  switch (env) {
    case 'prod':
      return 'green';
    case 'azure-dev':
      return 'orange';
    case 'local-dev':
    default:
      return 'blue';
  }
}

function getEnvLabel(env: string): string {
  switch (env) {
    case 'prod':
      return 'Production';
    case 'azure-dev':
      return 'Azure Dev';
    case 'local-dev':
    default:
      return 'Local Dev';
  }
}

export function VersionBadge() {
  const { user } = useAuth();
  const { env, version, gitSha } = getVersionInfo();
  const isProd = env === 'prod';
  
  // In production, only show to admin users
  if (isProd && !user?.is_admin) {
    return null;
  }
  const displayVersion = isProd ? version : gitSha.substring(0, 7);
  
  // Build GitHub links
  const repoUrl = 'https://github.com/your-org/go-react-starter';
  const versionUrl = isProd 
    ? `${repoUrl}/releases/tag/${version}`
    : `${repoUrl}/commit/${gitSha}`;

  return (
    <Group gap="xs" className={classes.container}>
      <Badge 
        variant="light" 
        color={getEnvColor(env)} 
        size="sm"
        radius="sm"
      >
        {getEnvLabel(env)}
      </Badge>
      <Text size="xs" c="dimmed">
        {isProd ? 'v' : ''}
        <Anchor
          href={versionUrl}
          target="_blank"
          rel="noopener noreferrer"
          size="xs"
          c="dimmed"
          underline="hover"
        >
          {displayVersion}
        </Anchor>
      </Text>
    </Group>
  );
}
