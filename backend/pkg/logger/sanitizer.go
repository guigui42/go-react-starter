// Package logger provides structured logging utilities for the the API.
// This file provides sanitization utilities for redacting sensitive data from logs.
package logger

import (
	"regexp"
)

// sensitivePatterns contains compiled regex patterns for sensitive data that should be redacted.
// These patterns match common sensitive values in URLs and headers.
// Patterns use (?:^|[?&]) to ensure we match parameter names at boundaries.
//
// Note: Package-level compiled regex patterns are a standard Go pattern for performance.
// They are initialized once at package load and reused, which is more efficient than
// compiling on each call. This follows Go best practices (see regexp.MustCompile docs).
var sensitivePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?:^|[?&])api_token=[^&\s]+`),
	regexp.MustCompile(`(?:^|[?&])api_key=[^&\s]+`),
	regexp.MustCompile(`(?:^|[?&])apikey=[^&\s]+`),
	regexp.MustCompile(`(?:^|[?&])password=[^&\s]+`),
	regexp.MustCompile(`(?:^|[?&])secret=[^&\s]+`),
	regexp.MustCompile(`(?:^|[?&])access_token=[^&\s]+`),
	regexp.MustCompile(`(?i)Authorization:\s*Bearer\s+[^\s]+`),
	regexp.MustCompile(`(?i)X-API-Key:\s*[^\s]+`),
}

// Sanitize removes sensitive data from a string by replacing it with [REDACTED].
// It handles common patterns like API tokens in URLs and authorization headers.
// This function is safe to use on URLs, log messages, and other strings that
// may contain sensitive information.
//
// Example:
//
//	url := "https://api.example.com?api_token=secret123&fmt=json"
//	safe := Sanitize(url)
//	// safe = "https://api.example.com?api_token=[REDACTED]&fmt=json"
func Sanitize(s string) string {
	for _, pattern := range sensitivePatterns {
		s = pattern.ReplaceAllStringFunc(s, func(match string) string {
			// Find the position of '=' to preserve the key name and any prefix (&, ?)
			for i, ch := range match {
				if ch == '=' {
					return match[:i+1] + "[REDACTED]"
				}
				// Handle Authorization and X-API-Key headers (colon separator)
				if ch == ':' {
					return match[:i+1] + " [REDACTED]"
				}
			}
			return "[REDACTED]"
		})
	}
	return s
}
