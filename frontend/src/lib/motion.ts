/**
 * Shared Framer Motion utilities and animation variants
 *
 * This file provides reusable animation patterns for consistent
 * motion design across the Go React Starter application.
 */

import {
    Box,
    Card,
    Container,
    Group,
    Paper,
    SimpleGrid,
    Stack,
    Text,
    Title,
} from '@mantine/core';
import { motion, type Variants } from 'framer-motion';

// ============================================================================
// Animation Variants
// ============================================================================

/**
 * Fade in from below with slight upward movement
 */
export const fadeInUp: Variants = {
  hidden: { opacity: 0, y: 20 },
  visible: {
    opacity: 1,
    y: 0,
    transition: { duration: 0.3, ease: 'easeOut' },
  },
};

/**
 * Fade in from above with slight downward movement
 */
export const fadeInDown: Variants = {
  hidden: { opacity: 0, y: -20 },
  visible: {
    opacity: 1,
    y: 0,
    transition: { duration: 0.3, ease: 'easeOut' },
  },
};

/**
 * Fade in from left
 */
export const fadeInLeft: Variants = {
  hidden: { opacity: 0, x: -20 },
  visible: {
    opacity: 1,
    x: 0,
    transition: { duration: 0.3, ease: 'easeOut' },
  },
};

/**
 * Fade in from right
 */
export const fadeInRight: Variants = {
  hidden: { opacity: 0, x: 20 },
  visible: {
    opacity: 1,
    x: 0,
    transition: { duration: 0.3, ease: 'easeOut' },
  },
};

/**
 * Simple fade without movement
 */
export const fadeIn: Variants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: { duration: 0.25, ease: 'easeOut' },
  },
};

/**
 * Scale up from slightly smaller
 */
export const scaleIn: Variants = {
  hidden: { opacity: 0, scale: 0.95 },
  visible: {
    opacity: 1,
    scale: 1,
    transition: { duration: 0.25, ease: 'easeOut' },
  },
};

/**
 * Stagger container for orchestrating child animations
 */
export const staggerContainer: Variants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: {
      staggerChildren: 0.1,
      delayChildren: 0.1,
    },
  },
};

/**
 * Fast stagger for many items
 */
export const staggerContainerFast: Variants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: {
      staggerChildren: 0.05,
      delayChildren: 0.05,
    },
  },
};

// Alias for shorter import
export const staggerFast = staggerContainerFast;

/**
 * Slow stagger for emphasis
 */
export const staggerContainerSlow: Variants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: {
      staggerChildren: 0.15,
      delayChildren: 0.2,
    },
  },
};

/**
 * Page transition variant
 */
export const pageTransition: Variants = {
  hidden: { opacity: 0, y: 10 },
  visible: {
    opacity: 1,
    y: 0,
    transition: { duration: 0.3, ease: 'easeOut' },
  },
  exit: {
    opacity: 0,
    y: -10,
    transition: { duration: 0.2, ease: 'easeIn' },
  },
};

/**
 * Card hover animation props
 */
export const cardHover = {
  whileHover: { y: -4, transition: { duration: 0.2 } },
  whileTap: { scale: 0.98 },
};

/**
 * Subtle card hover (less movement)
 */
export const cardHoverSubtle = {
  whileHover: { y: -2, transition: { duration: 0.2 } },
};

/**
 * Button press animation
 */
export const buttonPress = {
  whileTap: { scale: 0.97 },
};

// ============================================================================
// Motion Components (Mantine + Framer Motion)
// ============================================================================

// Note: Using type assertion for Mantine polymorphic components
// because motion.create() has stricter type requirements than runtime needs.
// These work correctly at runtime - the type is overly restrictive.
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const createMotionComponent = <T>(component: T) => motion.create(component as any);

export const MotionBox = createMotionComponent(Box);
export const MotionCard = createMotionComponent(Card);
export const MotionContainer = createMotionComponent(Container);
export const MotionGroup = createMotionComponent(Group);
export const MotionPaper = createMotionComponent(Paper);
export const MotionSimpleGrid = createMotionComponent(SimpleGrid);
export const MotionStack = createMotionComponent(Stack);
export const MotionTitle = createMotionComponent(Title);
export const MotionText = createMotionComponent(Text);

// Native HTML elements with motion
export const MotionDiv = motion.div;
export const MotionSection = motion.section;

// ============================================================================
// Animation Presets (ready-to-use props)
// ============================================================================

/**
 * Animate on scroll into view
 */
export const animateOnScroll = {
  initial: 'hidden',
  whileInView: 'visible',
  viewport: { once: true, margin: '-50px' },
};

/**
 * Animate immediately on mount
 */
export const animateOnMount = {
  initial: 'hidden',
  animate: 'visible',
};

/**
 * Page wrapper animation props
 */
export const pageAnimationProps = {
  initial: 'hidden',
  animate: 'visible',
  exit: 'exit',
  variants: pageTransition,
};

// ============================================================================
// Custom Hooks & Utilities
// ============================================================================

/**
 * Get stagger delay for an item in a list
 */
export function getStaggerDelay(index: number, baseDelay = 0.1): number {
  return index * baseDelay;
}

/**
 * Create custom item variants with specific delay
 */
export function createItemVariants(delay = 0): Variants {
  return {
    hidden: { opacity: 0, y: 20 },
    visible: {
      opacity: 1,
      y: 0,
      transition: { duration: 0.3, ease: 'easeOut', delay },
    },
  };
}
