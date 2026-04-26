import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MantineProvider } from '@mantine/core'
import { Button } from './Button'

const TestWrapper = ({ children }: { children: React.ReactNode }) => {
  return <MantineProvider>{children}</MantineProvider>
}

describe('Button', () => {
  it('should render children', () => {
    render(
      <TestWrapper>
        <Button>Click me</Button>
      </TestWrapper>
    )
    expect(screen.getByText('Click me')).toBeInTheDocument()
  })

  it('should render as button element', () => {
    render(
      <TestWrapper>
        <Button>Submit</Button>
      </TestWrapper>
    )
    const button = screen.getByRole('button')
    expect(button).toBeInTheDocument()
    expect(button).toHaveTextContent('Submit')
  })

  it('should apply variant prop', () => {
    render(
      <TestWrapper>
        <Button variant="filled">Submit</Button>
      </TestWrapper>
    )
    const button = screen.getByRole('button')
    expect(button).toHaveClass('mantine-Button-root')
  })

  it('should be clickable when not disabled', () => {
    render(
      <TestWrapper>
        <Button>Click</Button>
      </TestWrapper>
    )
    const button = screen.getByRole('button')
    expect(button).not.toBeDisabled()
  })

  it('should be disabled when disabled prop is true', () => {
    render(
      <TestWrapper>
        <Button disabled>Click</Button>
      </TestWrapper>
    )
    const button = screen.getByRole('button')
    expect(button).toBeDisabled()
  })

  it('should forward ref to button element', () => {
    const ref = { current: null }
    render(
      <TestWrapper>
        <Button ref={ref}>Click</Button>
      </TestWrapper>
    )
    expect(ref.current).toBeInstanceOf(HTMLButtonElement)
  })

  it('should apply custom className', () => {
    render(
      <TestWrapper>
        <Button className="custom-class">Click</Button>
      </TestWrapper>
    )
    const button = screen.getByRole('button')
    expect(button).toHaveClass('custom-class')
  })
})
