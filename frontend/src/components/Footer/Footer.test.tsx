import { MantineProvider } from '@mantine/core'
import { render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { Footer } from './Footer'

// Mock TanStack Router's Link
vi.mock('@tanstack/react-router', () => ({
  Link: ({ children, to, ...props }: { children: React.ReactNode; to: string }) => (
    <a href={to} {...props}>{children}</a>
  ),
}))

const renderWithProviders = (ui: React.ReactElement) => {
  return render(
    <MantineProvider>
      {ui}
    </MantineProvider>
  )
}

describe('Footer', () => {
  it('should render the footer element', () => {
    renderWithProviders(<Footer />)
    
    expect(screen.getByRole('contentinfo')).toBeInTheDocument()
  })

  it('should display the copyright text', () => {
    renderWithProviders(<Footer />)
    
    const year = new Date().getFullYear()
    expect(screen.getByText(new RegExp(`© ${year}`))).toBeInTheDocument()
  })

  it('should display the company name', () => {
    renderWithProviders(<Footer />)
    
    expect(screen.getByText(/Go React Starter/)).toBeInTheDocument()
  })

  it('should have a link to privacy policy', () => {
    renderWithProviders(<Footer />)
    
    const privacyLink = screen.getByRole('link', { name: /privacy/i })
    expect(privacyLink).toBeInTheDocument()
    expect(privacyLink).toHaveAttribute('href', '/privacy')
  })

  it('should have a link to terms of service', () => {
    renderWithProviders(<Footer />)
    
    const termsLink = screen.getByRole('link', { name: /terms/i })
    expect(termsLink).toBeInTheDocument()
    expect(termsLink).toHaveAttribute('href', '/terms')
  })

  it('should have a link to legal notice', () => {
    renderWithProviders(<Footer />)
    
    const legalLink = screen.getByRole('link', { name: /legal/i })
    expect(legalLink).toBeInTheDocument()
    expect(legalLink).toHaveAttribute('href', '/legal')
  })

  it('should be accessible via keyboard', async () => {
    renderWithProviders(<Footer />)
    
    const privacyLink = screen.getByRole('link', { name: /privacy/i })
    const termsLink = screen.getByRole('link', { name: /terms/i })
    const legalLink = screen.getByRole('link', { name: /legal/i })
    
    expect(privacyLink).not.toHaveAttribute('tabindex', '-1')
    expect(termsLink).not.toHaveAttribute('tabindex', '-1')
    expect(legalLink).not.toHaveAttribute('tabindex', '-1')
  })

  it('should have proper semantic structure', () => {
    renderWithProviders(<Footer />)
    
    const footer = screen.getByRole('contentinfo')
    expect(footer).toBeInTheDocument()
    
    const nav = screen.getByRole('navigation', { name: /footer/i })
    expect(nav).toBeInTheDocument()
  })
})
