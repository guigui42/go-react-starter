import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { userEvent } from '@testing-library/user-event';
import { PasskeyRegistrationForm } from './PasskeyRegistrationForm';
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
        'passkey.registration.description': 'Create a passkey for your device',
        'passkey.registration.friendly_name_label': 'Device Name',
        'passkey.registration.friendly_name_placeholder': 'My iPhone',
        'passkey.registration.friendly_name_description': 'Optional name for this passkey',
        'passkey.registration.create_button': 'Create Passkey',
        'passkey.errors.not_supported': 'Passkeys are not supported in this browser',
        'passkey.errors.registration_failed': 'Registration failed',
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

describe('PasskeyRegistrationForm', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render registration form when browser is supported', () => {
    vi.mocked(isWebAuthnSupported).mockReturnValue(true);

    renderWithProviders(<PasskeyRegistrationForm />);

    expect(screen.getByLabelText(/device name/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /create passkey/i })).toBeInTheDocument();
  });

  it('should show not supported message when browser is not supported', () => {
    vi.mocked(isWebAuthnSupported).mockReturnValue(false);

    renderWithProviders(<PasskeyRegistrationForm />);

    expect(screen.getByText(/not supported/i)).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /create passkey/i })).not.toBeInTheDocument();
  });

  it('should allow user to enter friendly name', async () => {
    vi.mocked(isWebAuthnSupported).mockReturnValue(true);
    const user = userEvent.setup();

    renderWithProviders(<PasskeyRegistrationForm />);

    const input = screen.getByLabelText(/device name/i);
    await user.type(input, 'iPhone 15 Pro');

    expect(input).toHaveValue('iPhone 15 Pro');
  });

  it('should call register when form is submitted', async () => {
    vi.mocked(isWebAuthnSupported).mockReturnValue(true);
    const mockRegister = vi.mocked(passkeyService.register);
    mockRegister.mockResolvedValue({
      credentialId: 'cred-123',
      friendlyName: 'iPhone 15 Pro',
      createdAt: new Date().toISOString(),
    });

    const onSuccess = vi.fn();
    const user = userEvent.setup();

    renderWithProviders(<PasskeyRegistrationForm onSuccess={onSuccess} />);

    const input = screen.getByLabelText(/device name/i);
    await user.type(input, 'iPhone 15 Pro');

    const button = screen.getByRole('button', { name: /create passkey/i });
    await user.click(button);

    await waitFor(() => {
      expect(mockRegister).toHaveBeenCalledWith('iPhone 15 Pro');
      expect(onSuccess).toHaveBeenCalledWith('cred-123');
    });
  });

  it('should handle registration error', async () => {
    vi.mocked(isWebAuthnSupported).mockReturnValue(true);
    const mockRegister = vi.mocked(passkeyService.register);
    mockRegister.mockRejectedValue(new Error('Registration failed'));

    const onError = vi.fn();
    const user = userEvent.setup();

    renderWithProviders(<PasskeyRegistrationForm onError={onError} />);

    const button = screen.getByRole('button', { name: /create passkey/i });
    await user.click(button);

    await waitFor(() => {
      expect(screen.getByText(/registration failed/i)).toBeInTheDocument();
      expect(onError).toHaveBeenCalled();
    });
  });

  it('should disable button while registering', async () => {
    vi.mocked(isWebAuthnSupported).mockReturnValue(true);
    const mockRegister = vi.mocked(passkeyService.register);
    mockRegister.mockImplementation(() => new Promise(() => {})); // Never resolves

    const user = userEvent.setup();

    renderWithProviders(<PasskeyRegistrationForm />);

    const button = screen.getByRole('button', { name: /create passkey/i });
    await user.click(button);

    await waitFor(() => {
      expect(button).toBeDisabled();
    });
  });

  it('should work without friendly name', async () => {
    vi.mocked(isWebAuthnSupported).mockReturnValue(true);
    const mockRegister = vi.mocked(passkeyService.register);
    mockRegister.mockResolvedValue({
      credentialId: 'cred-123',
      friendlyName: '',
      createdAt: new Date().toISOString(),
    });

    const onSuccess = vi.fn();
    const user = userEvent.setup();

    renderWithProviders(<PasskeyRegistrationForm onSuccess={onSuccess} />);

    const button = screen.getByRole('button', { name: /create passkey/i });
    await user.click(button);

    await waitFor(() => {
      expect(mockRegister).toHaveBeenCalledWith(undefined);
      expect(onSuccess).toHaveBeenCalledWith('cred-123');
    });
  });
});
