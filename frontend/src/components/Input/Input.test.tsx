import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { userEvent } from '@testing-library/user-event'
import { MantineProvider } from '@mantine/core'
import { Input } from './Input'

const TestWrapper = ({ children }: { children: React.ReactNode }) => {
  return <MantineProvider>{children}</MantineProvider>
}

describe('Input', () => {
  it('should render with label', () => {
    render(
      <TestWrapper>
        <Input label="Name" />
      </TestWrapper>
    )
    expect(screen.getByLabelText('Name')).toBeInTheDocument()
  })

  it('should render as text input by default', () => {
    render(
      <TestWrapper>
        <Input label="Email" />
      </TestWrapper>
    )
    const input = screen.getByLabelText('Email')
    expect(input).toBeInTheDocument()
    expect(input.tagName).toBe('INPUT')
  })

  it('should accept user input', async () => {
    const user = userEvent.setup()
    render(
      <TestWrapper>
        <Input label="Username" />
      </TestWrapper>
    )
    const input = screen.getByLabelText('Username')
    await user.type(input, 'test-user')
    expect(input).toHaveValue('test-user')
  })

  it('should display placeholder', () => {
    render(
      <TestWrapper>
        <Input label="Email" placeholder="Enter your email" />
      </TestWrapper>
    )
    expect(screen.getByPlaceholderText('Enter your email')).toBeInTheDocument()
  })

  it('should show error message', () => {
    render(
      <TestWrapper>
        <Input label="Email" error="Invalid email" />
      </TestWrapper>
    )
    expect(screen.getByText('Invalid email')).toBeInTheDocument()
  })

  it('should be disabled when disabled prop is true', () => {
    render(
      <TestWrapper>
        <Input label="Name" disabled />
      </TestWrapper>
    )
    expect(screen.getByLabelText('Name')).toBeDisabled()
  })

  it('should forward ref to input element', () => {
    const ref = { current: null }
    render(
      <TestWrapper>
        <Input label="Test" ref={ref} />
      </TestWrapper>
    )
    expect(ref.current).toBeInstanceOf(HTMLInputElement)
  })

  it('should apply required attribute', () => {
    render(
      <TestWrapper>
        <Input label="Email" required />
      </TestWrapper>
    )
    // Mantine adds an asterisk to the label when required, so we use a regex
    expect(screen.getByLabelText(/Email/)).toBeRequired()
  })

  it('should apply type attribute', () => {
    render(
      <TestWrapper>
        <Input label="Password" type="password" />
      </TestWrapper>
    )
    const input = screen.getByLabelText('Password')
    expect(input).toHaveAttribute('type', 'password')
  })
})
