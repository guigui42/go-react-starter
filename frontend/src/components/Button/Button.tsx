import { Button as MantineButton, type ButtonProps } from '@mantine/core'
import { forwardRef } from 'react'

// eslint-disable-next-line @typescript-eslint/no-empty-object-type
export interface CustomButtonProps extends ButtonProps {
  // Add custom props if needed in the future
}

export const Button = forwardRef<HTMLButtonElement, CustomButtonProps>(
  (props, ref) => {
    return <MantineButton ref={ref} {...props} />
  }
)

Button.displayName = 'Button'
