import js from '@eslint/js'
import globals from 'globals'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import react from 'eslint-plugin-react'
import tseslint from 'typescript-eslint'
import importPlugin from 'eslint-plugin-import'

const sharedLanguageOptions = {
  ecmaVersion: 2024,
  globals: globals.browser,
  parserOptions: {
    ecmaFeatures: { jsx: true },
  },
}

const sharedPlugins = {
  react,
  'react-hooks': reactHooks,
  'react-refresh': reactRefresh,
  'import': importPlugin,
}

const sharedSettings = {
  react: { version: 'detect' },
  'import/resolver': {
    typescript: {
      alwaysTryTypes: true,
      project: './tsconfig.json',
    },
    node: true,
  },
}

const sharedRules = {
  // React core rules
  ...react.configs.recommended.rules,
  'react/react-in-jsx-scope': 'off', // Not needed in React 19
  'react/prop-types': 'off', // Using TypeScript for prop validation
  
  // React Hooks rules
  ...reactHooks.configs.recommended.rules,
  
  // Console usage - use proper error handling instead
  'no-console': ['warn', { allow: ['warn'] }],
  
  // TypeScript rules
  '@typescript-eslint/explicit-module-boundary-types': 'off',
  '@typescript-eslint/no-unused-vars': [
    'error',
    { argsIgnorePattern: '^_', varsIgnorePattern: '^_' },
  ],
  
  // Import rules - enforce @ alias usage for src/ imports
  'no-restricted-imports': [
    'error',
    {
      patterns: [
        {
          group: ['../*', '../../*', '../../../*', '../../../../*'],
          message: 'Relative parent imports are not allowed. Use @ alias instead (e.g., @/components/...)',
        },
      ],
    },
  ],
}

export default tseslint.config(
  // Ignore patterns
  {
    ignores: ['dist', 'node_modules', 'coverage', 'playwright-report'],
  },
  
  // Base configs
  js.configs.recommended,
  ...tseslint.configs.recommended,
  
  // All TypeScript/React files (excluding route files)
  {
    files: ['**/*.{ts,tsx}'],
    ignores: ['src/routes/**', 'src/main.tsx'],
    languageOptions: sharedLanguageOptions,
    plugins: sharedPlugins,
    settings: sharedSettings,
    rules: {
      ...sharedRules,
      'react-refresh/only-export-components': [
        'warn',
        { allowConstantExport: true },
      ],
    },
  },

  // Route files and main.tsx: disable react-refresh (TanStack Router file-based routing)
  {
    files: ['src/routes/**/*.{ts,tsx}', 'src/main.tsx'],
    languageOptions: sharedLanguageOptions,
    plugins: sharedPlugins,
    settings: sharedSettings,
    rules: {
      ...sharedRules,
      'react-refresh/only-export-components': 'off',
    },
  }
)
