import type { Metric } from 'web-vitals';

/**
 * Umami global object injected by the Umami tracking script.
 * Used to send custom events for Web Vitals metrics.
 */
declare global {
  interface Window {
    umami?: {
      track: (eventName: string, data?: Record<string, string | number>) => void;
    };
  }
}

/**
 * Reports a Web Vitals metric to Umami as a custom event.
 *
 * Each metric is sent with its name, value, rating, delta, and navigation type.
 * Events are fire-and-forget and will not block page interactions.
 *
 * @param metric - The Web Vitals metric to report
 */
function sendToUmami(metric: Metric): void {
  window.umami?.track(`web-vitals-${metric.name}`, {
    value: Math.round(metric.name === 'CLS' ? metric.delta * 1000 : metric.delta),
    rating: metric.rating,
    navigationType: metric.navigationType,
  });
}

/**
 * Initializes Web Vitals reporting to Umami.
 *
 * Captures Core Web Vitals (LCP, CLS, INP) and additional metrics (FCP, TTFB)
 * from real users and reports them as Umami custom events.
 *
 * Note: FID was deprecated by Chrome in favor of INP and is not available
 * in web-vitals v5+. FCP (First Contentful Paint) is captured instead.
 *
 * This function is non-blocking and uses dynamic import to avoid
 * impacting initial page load performance.
 */
export function initWebVitals(): void {
  import('web-vitals').then(({ onCLS, onFCP, onINP, onLCP, onTTFB }) => {
    onCLS(sendToUmami);
    onFCP(sendToUmami);
    onINP(sendToUmami);
    onLCP(sendToUmami);
    onTTFB(sendToUmami);
  }).catch(() => {
    // Silently ignore — Web Vitals are non-critical telemetry
  });
}
