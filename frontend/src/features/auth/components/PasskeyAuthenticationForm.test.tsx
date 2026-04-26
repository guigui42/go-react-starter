import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { userEvent } from '@testing-library/user-event';
import { PasskeyAuthenticationForm } from './PasskeyAuthenticationForm';
import { passkeyService, isWebAuthnSupported } from '@/features/auth/services';
import { MantineProvider } from '@mantine/core';
import type { ReactNode } from 'react';

// Mock the services
vi.mock('@/features/auth/services');

// Mock i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'passkey.not_supported_title': 'Not Supported',
        'passkey.not_supported_message': 'Your browser does not support passkeys',
        'passkey.authentication.description': 'Sign in using your passkey',
        'passkey.authentication.sign_in_button': 'Sign in with Passkey',
        'passkey.authentication.use_backup_code': 'Use Backup Code',
        'passkey.errors.not_supported': 'Passkeys are not supported in this browser',
        'passkey.errors.authentication_failed': 'Authentication failed',
        'common.error': 'Error',
      };
      return translations[key] || key;
    },
  }),
}));

const TestWrapper = ({ children }: { children: ReactNode }) => {
  return <MantineProvider>{children}</MantineProvider>;
};

const renderWithProviders = (ui: ReactNode) => {
  return render(<TestWrapper>{ui}</TestWrapper>);
};

describe('PasskeyAuthenticationForm', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render authentication form when browser is supported', () => {
    vi.mocked(isWebAuthnSupported).mockReturnValue(true);

    renderWithProviders(<PasskeyAuthenticationForm email="test@example.com" />);

    expect(screen.getByRole('button', { name: /sign in with passkey/i })).toBeInTheDocument();
  });

  it('should show not supported message when browser is not supported', () => {
    vi.mocked(isWebAuthnSupported).mockReturnValue(false);

    renderWithProviders(<PasskeyAuthenticationForm email="test@example.com" />);

    expect(screen.getByText(/not supported/i)).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /sign in/i })).not.toBeInTheDocument();
  });

  it('should call authenticate when button is clicked', async () => {
    vi.mocked(isWebAuthnSupported).mockReturnValue(true);
    const mockAuthenticate = vi.mocked(passkeyService.authenticate);
    mockAuthenticate.mockResolvedValue({
      token: 'jwt-token-123',
      user: { id: 'user-1', email: 'test@example.com', firstName: 'Test', lastName: 'User' },
    });

    const onSuccess = vi.fn();
    const user = userEvent.setup();

    renderWithProviders(<PasskeyAuthenticationForm email="test@example.com" onSuccess={onSuccess} />);

    const button = screen.getByRole('button', { name: /sign in with passkey/i });
    await user.click(button);

    await waitFor(() => {
      expect(mockAuthenticate).toHaveBeenCalledWith('test@example.com');
      expect(onSuccess).toHaveBeenCalledWith('jwt-token-123');
    });
  });

  it('should handle authentication error', async () => {
    vi.mocked(isWebAuthnSupported).mockReturnValue(true);
    const mockAuthenticate = vi.mocked(passkeyService.authenticate);
    mockAuthenticate.mockRejectedValue(new Error('Authentication failed'));

    const onError = vi.fn();
    const user = userEvent.setup();

    renderWithProviders(<PasskeyAuthenticationForm email="test@example.com" onError={onError} />);

    const button = screen.getByRole('button', { name: /sign in with passkey/i });
    await user.click(button);

    await waitFor(() => {
      expect(screen.getByText(/authentication failed/i)).toBeInTheDocument();
      expect(onError).toHaveBeenCalled();
    });
  });

  it('should disable button while authenticating', async () => {
    vi.mocked(isWebAuthnSupported).mockReturnValue(true);
    const mockAuthenticate = vi.mocked(passkeyService.authenticate);
    mockAuthenticate.mockImplementation(() => new Promise(() => {})); // Never resolves

    const user = userEvent.setup();

    renderWithProviders(<PasskeyAuthenticationForm email="test@example.com" />);

    const button = screen.getByRole('button', { name: /sign in with passkey/i });
    await user.click(button);

    await waitFor(() => {
      expect(button).toBeDisabled();
    });
  });

  it('should show backup code button when enabled', () => {
    vi.mocked(isWebAuthnSupported).mockReturnValue(true);
    const onUseBackupCode = vi.fn();

    renderWithProviders(<PasskeyAuthenticationForm email="test@example.com" onUseBackupCode={onUseBackupCode} showBackupCodeOption={true} />);

    expect(screen.getByRole('button', { name: /use backup code/i })).toBeInTheDocument();
  });

  it('should hide backup code button when disabled', () => {
    vi.mocked(isWebAuthnSupported).mockReturnValue(true);

    renderWithProviders(<PasskeyAuthenticationForm email="test@example.com" showBackupCodeOption={false} />);

    expect(screen.queryByRole('button', { name: /use backup code/i })).not.toBeInTheDocument();
  });

  it('should call onUseBackupCode when backup code button is clicked', async () => {
    vi.mocked(isWebAuthnSupported).mockReturnValue(true);
    const onUseBackupCode = vi.fn();
    const user = userEvent.setup();

    renderWithProviders(<PasskeyAuthenticationForm email="test@example.com" onUseBackupCode={onUseBackupCode} />);

    const button = screen.getByRole('button', { name: /use backup code/i });
    await user.click(button);

    expect(onUseBackupCode).toHaveBeenCalled();
  });
});
