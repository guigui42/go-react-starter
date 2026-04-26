import { Card as MantineCard, type CardProps } from '@mantine/core'
import { type ReactNode } from 'react'

export interface CustomCardProps extends CardProps {
  children: ReactNode
}

export function Card({ children, ...props }: CustomCardProps) {
  return (
    <MantineCard shadow="sm" padding="lg" radius="md" withBorder {...props}>
      {children}
    </MantineCard>
  )
}
