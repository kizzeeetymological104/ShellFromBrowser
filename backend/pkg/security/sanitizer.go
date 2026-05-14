package security

import (
	"regexp"
	"strings"
	"unicode"
)

// ============================================
// Input Sanitizer — Command Injection Prevention
// ============================================
// Protects against:
// - ANSI escape sequences (terminal control hijacking)
// - Unicode right-to-left override (visual spoofing)
// - Shell metacharacters (command injection)

var (
	// Dangerous ANSI escape sequences
	ansiEscapeRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

	// Unicode right-to-left override (U+202E)
	rloChar = '‮'

	// Shell metacharacters (basic list for Phase 1)
	shellMetachars = []string{";", "&", "|", "`", "$", "(", ")", "<", ">", "\n", "\r"}
)

// SanitizeInput sanitizes user input for terminal commands
func SanitizeInput(input string, maxLength int) (string, bool) {
	// 1. Length check
	if len(input) > maxLength {
		return "", false
	}

	// 2. Remove ANSI escape sequences
	sanitized := ansiEscapeRegex.ReplaceAllString(input, "")

	// 3. Check for Unicode RLO (right-to-left override)
	if strings.ContainsRune(sanitized, rloChar) {
		return "", false
	}

	// 4. Check for shell metacharacters
	for _, meta := range shellMetachars {
		if strings.Contains(sanitized, meta) {
			return "", false
		}
	}

	// 5. Check for null bytes
	if strings.ContainsRune(sanitized, '\x00') {
		return "", false
	}

	// 6. Trim whitespace
	sanitized = strings.TrimSpace(sanitized)

	return sanitized, true
}

// IsAllowedCommand checks if a command is in the allowlist (Phase 2 feature)
func IsAllowedCommand(cmd string, allowlist []string) bool {
	// Extract command name (first word)
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return false
	}

	cmdName := parts[0]

	// Check allowlist
	for _, allowed := range allowlist {
		if cmdName == allowed {
			return true
		}
	}

	return false
}

// ContainsDangerousPatterns checks for known dangerous patterns
func ContainsDangerousPatterns(input string) bool {
	// Patterns that indicate potential attacks
	dangerousPatterns := []string{
		"rm -rf",
		":(){ :|:& };:", // Fork bomb
		"/dev/tcp",      // Network redirection
		"chmod +s",      // SUID escalation
		"curl | sh",     // Remote code execution
		"wget | sh",
	}

	lowered := strings.ToLower(input)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowered, pattern) {
			return true
		}
	}

	return false
}

// ValidateFilePath validates file paths to prevent directory traversal
func ValidateFilePath(path string) bool {
	// Block directory traversal attempts
	if strings.Contains(path, "..") {
		return false
	}

	// Block absolute paths outside allowed directories (Phase 1: block all absolute)
	if strings.HasPrefix(path, "/") {
		return false
	}

	// Block null bytes
	if strings.ContainsRune(path, '\x00') {
		return false
	}

	return true
}

// StripNonPrintable removes non-printable characters
func StripNonPrintable(input string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) || r == '\n' || r == '\t' {
			return r
		}
		return -1 // Remove character
	}, input)
}
