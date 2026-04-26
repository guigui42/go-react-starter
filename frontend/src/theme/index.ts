import {
  createTheme,
  type CSSVariablesResolver,
  type MantineColorsTuple,
  rem,
} from '@mantine/core';

/**
 * Mantine v8 theme configuration for Go React Starter.
 * Premium financial application design system with modern aesthetics.
 *
 * @see https://mantine.dev/theming/theme-object/
 */

// Dark palette with cool blue undertones for premium depth layering
// dark[7] = page body, dark[6] = cards/panels, dark[5] = modals/popovers
const AppDark: MantineColorsTuple = [
  '#C1C8D6',
  '#A6AEBF',
  '#8B94A9',
  '#6C7793',
  '#2e3347',
  '#242836',
  '#1a1d27',
  '#0f1117',
  '#0b0d12',
  '#070810',
];

// Primary brand color - Modern deep blue with confidence
const AppBlue: MantineColorsTuple = [
  '#eef3ff',
  '#dce4f5',
  '#b9c7e2',
  '#94a8d0',
  '#748dc1',
  '#5f7cb8',
  '#5474b4',
  '#44639f',
  '#39588f',
  '#2b4a80',
];

// Premium emerald for success/gains - Evokes growth and prosperity
const AppGreen: MantineColorsTuple = [
  '#e6fcf5',
  '#c3fae8',
  '#8de4c8',
  '#52d1a4',
  '#26c285',
  '#0db970',
  '#00b865',
  '#00a257',
  '#00904c',
  '#007d3e',
];

// Warm gold for accents - Premium feel
const AppGold: MantineColorsTuple = [
  '#fff9db',
  '#fff3bf',
  '#ffec99',
  '#ffe066',
  '#ffd43b',
  '#fcc419',
  '#fab005',
  '#f59f00',
  '#f08c00',
  '#e67700',
];

// Coral for losses/alerts
const AppCoral: MantineColorsTuple = [
  '#fff5f5',
  '#ffe3e3',
  '#ffc9c9',
  '#ffa8a8',
  '#ff8787',
  '#ff6b6b',
  '#fa5252',
  '#f03e3e',
  '#e03131',
  '#c92a2a',
];

// Sophisticated slate for text and backgrounds
const AppSlate: MantineColorsTuple = [
  '#f8f9fa',
  '#f1f3f5',
  '#e9ecef',
  '#dee2e6',
  '#ced4da',
  '#adb5bd',
  '#868e96',
  '#495057',
  '#343a40',
  '#212529',
];

export const theme = createTheme({
  /** DM Sans — characterful body font with excellent tabular-nums support */
  fontFamily:
    '"DM Sans", -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',

  headings: {
    /** DM Serif Display — editorial gravitas for headings */
    fontFamily:
      '"DM Serif Display", Georgia, "Times New Roman", serif',
    fontWeight: '400',
    sizes: {
      h1: { fontSize: 'clamp(1.75rem, 1.2rem + 2.5vw, 2.75rem)', lineHeight: '1.1', fontWeight: '400' },
      h2: { fontSize: 'clamp(1.5rem, 1.1rem + 2vw, 2.125rem)', lineHeight: '1.2', fontWeight: '400' },
      h3: { fontSize: 'clamp(1.25rem, 1rem + 1.25vw, 1.625rem)', lineHeight: '1.3', fontWeight: '400' },
      h4: { fontSize: 'clamp(1.125rem, 1rem + 0.5vw, 1.25rem)', lineHeight: '1.4', fontWeight: '400' },
    },
  },

  /** Primary color for branding */
  primaryColor: 'AppBlue',

  /** Extended color palette */
  colors: {
    AppBlue,
    AppGreen,
    AppGold,
    AppCoral,
    AppSlate,
    dark: AppDark,
  },

  /** Breakpoints */
  breakpoints: {
    xs: '36em', // 576px
    sm: '48em', // 768px
    md: '62em', // 992px
    lg: '75em', // 1200px
    xl: '88em', // 1408px
  },

  /** Refined spacing scale */
  spacing: {
    xs: rem(10),
    sm: rem(14),
    md: rem(20),
    lg: rem(28),
    xl: rem(40),
  },

  /** Radius */
  radius: {
    xs: '0.125rem',
    sm: '0.25rem',
    md: '0.5rem',
    lg: '1rem',
    xl: '2rem',
  },

  /** Modern rounded corners by default */
  defaultRadius: 'lg',

  /** Premium shadow system - using CSS variables for dark mode support */
  shadows: {
    xs: '0 1px 2px rgba(0, 0, 0, 0.05)',
    sm: '0 1px 3px rgba(0, 0, 0, 0.08), 0 1px 2px rgba(0, 0, 0, 0.06)',
    md: '0 4px 8px rgba(0, 0, 0, 0.08), 0 2px 4px rgba(0, 0, 0, 0.06)',
    lg: '0 10px 20px rgba(0, 0, 0, 0.1), 0 4px 8px rgba(0, 0, 0, 0.06)',
    xl: '0 20px 30px rgba(0, 0, 0, 0.12), 0 10px 12px rgba(0, 0, 0, 0.08)',
  },

  /** Component overrides for premium feel */
  components: {
    Button: {
      defaultProps: {
        fw: 600,
      },
      styles: {
        root: {
          transition: 'all 0.2s ease',
        },
      },
    },
    Card: {
      defaultProps: {
        shadow: 'sm',
        padding: 'lg',
        withBorder: true,
      },
      styles: {
        root: {
          transition: 'all 0.2s ease',
          borderColor: 'var(--app-border-subtle)',
        },
      },
    },
    Paper: {
      defaultProps: {
        shadow: 'sm',
      },
      styles: {
        root: {
          borderColor: 'var(--app-border-subtle)',
        },
      },
    },
    Modal: {
      styles: {
        content: {
          backgroundColor: 'var(--app-elevation-elevated)',
          borderColor: 'var(--app-border-subtle)',
        },
      },
    },
    Popover: {
      styles: {
        dropdown: {
          backgroundColor: 'var(--app-elevation-elevated)',
          borderColor: 'var(--app-border-subtle)',
        },
      },
    },
    Menu: {
      styles: {
        dropdown: {
          backgroundColor: 'var(--app-elevation-elevated)',
          borderColor: 'var(--app-border-subtle)',
        },
      },
    },
    Title: {
      styles: {
        root: {
          letterSpacing: '-0.01em',
        },
      },
    },
    Table: {
      styles: {
        td: {
          fontVariantNumeric: 'tabular-nums',
          lineHeight: 1.6,
        },
        th: {
          fontVariantNumeric: 'tabular-nums',
          lineHeight: 1.6,
        },
      },
    },
  },

  /** Other theme settings */
  cursorType: 'pointer',
});

/**
 * CSS variables resolver for dark/light elevation layering.
 * Provides semantic tokens: background → surface → elevated.
 */
export const cssVariablesResolver: CSSVariablesResolver = () => ({
  variables: {
    '--app-border-subtle': 'rgba(0, 0, 0, 0.08)',
    '--app-elevation-elevated': 'var(--mantine-color-body)',
  },
  light: {
    '--app-elevation-bg': '#ffffff',
    '--app-elevation-surface': '#ffffff',
    '--app-elevation-elevated': '#ffffff',
    '--app-border-subtle': 'rgba(0, 0, 0, 0.08)',
    '--app-accent-glow': 'none',
    '--app-shadow-sm': '0 1px 3px rgba(0,0,0,0.08), 0 1px 2px rgba(0,0,0,0.06)',
    '--app-shadow-md': '0 4px 8px rgba(0,0,0,0.08), 0 2px 4px rgba(0,0,0,0.06)',
  },
  dark: {
    '--app-elevation-bg': '#0f1117',
    '--app-elevation-surface': '#1a1d27',
    '--app-elevation-elevated': '#242836',
    '--app-border-subtle': 'rgba(255, 255, 255, 0.06)',
    '--app-accent-glow': '0 0 20px rgba(94, 124, 184, 0.3)',
    '--app-shadow-sm':
      '0 1px 3px rgba(0,0,0,0.4), 0 1px 2px rgba(0,0,0,0.3)',
    '--app-shadow-md':
      '0 4px 12px rgba(0,0,0,0.5), 0 2px 4px rgba(0,0,0,0.3)',
  },
});
