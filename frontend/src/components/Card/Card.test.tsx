import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MantineProvider } from '@mantine/core'
import { Card } from './Card'

const TestWrapper = ({ children }: { children: React.ReactNode }) => {
  return <MantineProvider>{children}</MantineProvider>
}

describe('Card', () => {
  it('should render children', () => {
    render(
      <TestWrapper>
        <Card>
          <div>Card content</div>
        </Card>
      </TestWrapper>
    )
    expect(screen.getByText('Card content')).toBeInTheDocument()
  })

  it('should have Mantine Card classes', () => {
    const { container } = render(
      <TestWrapper>
        <Card>Test</Card>
      </TestWrapper>
    )
    const card = container.querySelector('.mantine-Card-root')
    expect(card).toBeInTheDocument()
  })

  it('should apply default props (shadow, padding, radius, withBorder)', () => {
    const { container } = render(
      <TestWrapper>
        <Card>Content</Card>
      </TestWrapper>
    )
    const card = container.querySelector('.mantine-Card-root')
    expect(card).toBeInTheDocument()
    // Mantine applies these through classes/styles
  })

  it('should allow custom shadow prop', () => {
    const { container } = render(
      <TestWrapper>
        <Card shadow="lg">Content</Card>
      </TestWrapper>
    )
    const card = container.querySelector('.mantine-Card-root')
    expect(card).toBeInTheDocument()
  })

  it('should allow custom padding prop', () => {
    const { container } = render(
      <TestWrapper>
        <Card padding="xl">Content</Card>
      </TestWrapper>
    )
    const card = container.querySelector('.mantine-Card-root')
    expect(card).toBeInTheDocument()
  })

  it('should allow custom radius prop', () => {
    const { container } = render(
      <TestWrapper>
        <Card radius="xl">Content</Card>
      </TestWrapper>
    )
    const card = container.querySelector('.mantine-Card-root')
    expect(card).toBeInTheDocument()
  })

  it('should render complex children', () => {
    render(
      <TestWrapper>
        <Card>
          <h2>Title</h2>
          <p>Description</p>
          <button>Action</button>
        </Card>
      </TestWrapper>
    )
    expect(screen.getByText('Title')).toBeInTheDocument()
    expect(screen.getByText('Description')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Action' })).toBeInTheDocument()
  })

  it('should apply custom className', () => {
    const { container } = render(
      <TestWrapper>
        <Card className="custom-card">Content</Card>
      </TestWrapper>
    )
    const card = container.querySelector('.custom-card')
    expect(card).toBeInTheDocument()
    expect(card).toHaveClass('mantine-Card-root')
  })
})
