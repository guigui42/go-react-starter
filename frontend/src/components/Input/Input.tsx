import { TextInput, type TextInputProps } from '@mantine/core'
import { forwardRef } from 'react'

// eslint-disable-next-line @typescript-eslint/no-empty-object-type
export interface InputProps extends TextInputProps {
  // Add custom validation or behavior in the future if needed
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  (props, ref) => {
    return <TextInput ref={ref} {...props} />
  }
)

Input.displayName = 'Input'
