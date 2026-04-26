import '@testing-library/jest-dom';
import { cleanup } from '@testing-library/react';
import { afterEach, beforeEach, expect } from 'vitest';
import { configureAxe } from 'vitest-axe';
import * as matchers from 'vitest-axe/matchers';

// Import the app's i18n instance to configure it for tests
import i18n from '../src/i18n';

// Configure i18n for tests - set English as default before each test
// Using beforeEach ensures language is consistent even if a test changes it
beforeEach(async () => {
  // Ensure i18n is fully initialized
  if (!i18n.isInitialized) {
    await new Promise<void>((resolve) => {
      i18n.on('initialized', () => resolve());
    });
  }
  // Change language to English for consistent test results
  await i18n.changeLanguage('en');
});

// Mock localStorage for tests
const localStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => {
      store[key] = value.toString();
    },
    removeItem: (key: string) => {
      delete store[key];
    },
    clear: () => {
      store = {};
    },
    get length() {
      return Object.keys(store).length;
    },
    key: (index: number) => Object.keys(store)[index] || null,
  };
})();

Object.defineProperty(window, 'localStorage', {
  value: localStorageMock,
  writable: true,
});

// Extend matchers with accessibility matchers
expect.extend(matchers);

// Configure axe for WCAG 2.1 Level AA compliance
export const axe = configureAxe({
  rules: {
    // WCAG 2.1 Level AA rules
    'color-contrast': { enabled: true },
    'valid-lang': { enabled: true },
    'html-has-lang': { enabled: true },
    'aria-required-attr': { enabled: true },
    'button-name': { enabled: true },
    'image-alt': { enabled: true },
    'label': { enabled: true },
  },
});

// Mock window.matchMedia for Mantine components
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: (query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: () => {}, // Deprecated
    removeListener: () => {}, // Deprecated
    addEventListener: () => {},
    removeEventListener: () => {},
    dispatchEvent: () => true,
  }),
});

// Mock document.fonts (FontFaceSet API) for Mantine Textarea autosize (not available in jsdom 29+)
if (!document.fonts) {
  Object.defineProperty(document, 'fonts', {
    value: {
      addEventListener: () => {},
      removeEventListener: () => {},
      ready: Promise.resolve(),
    },
    writable: true,
  });
}

// Mock ResizeObserver for Mantine ScrollArea component
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
};

// Mock Element.scrollIntoView for Mantine Combobox (not available in JSDOM)
Element.prototype.scrollIntoView = Element.prototype.scrollIntoView || function () {};

// Mock IntersectionObserver for lazy loading components
global.IntersectionObserver = class IntersectionObserver {
  constructor() {}
  observe() {}
  unobserve() {}
  disconnect() {}
} as any;

// Set React testing environment flag
globalThis.IS_REACT_ACT_ENVIRONMENT = true;

// Polyfill for React.act in React 19 - the testing library expects this
// In React 19, act is only available in development/test builds
// We create a simple implementation that wraps the callback
const createActPolyfill = () => {
  return async function act(callback: () => void | Promise<void>) {
    const result = callback();
    if (result && typeof result.then === 'function') {
      await result;
    }
    // Allow microtasks to flush
    await new Promise((resolve) => setTimeout(resolve, 0));
    return result;
  };
};

// Polyfill react-dom/test-utils for @testing-library/react compatibility
const ReactDOMTestUtils = {
  act: createActPolyfill(),
};

// Make it available globally for react-dom/test-utils imports
(globalThis as any)['react-dom/test-utils'] = ReactDOMTestUtils;

// Cleanup after each test
afterEach(() => {
  cleanup();
});
