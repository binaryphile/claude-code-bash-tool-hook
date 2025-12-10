package main

import (
	"strings"
)

// EscapeMarkers are comment patterns that indicate a command should not be wrapped
// Users can add these to commands to opt-out of wrapping when needed
var EscapeMarkers = []string{
	"# bypass-hook",
	"# no-wrap",
	"# skip-hook",
}

// ShouldWrapCommand determines if a command needs to be wrapped
// Returns false only for: empty, already wrapped, escape markers, or disabled
func ShouldWrapCommand(command string, config *Config) bool {
	cmd := strings.TrimSpace(command)

	// Skip: empty, already wrapped, disabled
	if cmd == "" || strings.HasPrefix(cmd, "bash -c ") {
		return false
	}
	if config != nil && !config.Enabled {
		return false
	}

	// Skip: escape markers
	for _, marker := range EscapeMarkers {
		if strings.Contains(cmd, marker) {
			return false
		}
	}
	if config != nil {
		for _, marker := range config.AdditionalEscapeMarkers {
			if strings.Contains(cmd, marker) {
				return false
			}
		}
	}

	return true
}

// WrapCommand wraps a command in bash -c with base64 encoding to bypass preprocessing
// Base64 encoding eliminates all quoting edge cases - any byte sequence works
func WrapCommand(command string) string {
	encoded := base64Encode(command)
	return `bash -c "$(echo '` + encoded + `' | base64 -d)"`
}

// base64Encode encodes a string to base64 without newlines
func base64Encode(s string) string {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	var result strings.Builder
	input := []byte(s)

	for i := 0; i < len(input); i += 3 {
		var chunk uint32
		remaining := len(input) - i

		chunk = uint32(input[i]) << 16
		if remaining > 1 {
			chunk |= uint32(input[i+1]) << 8
		}
		if remaining > 2 {
			chunk |= uint32(input[i+2])
		}

		result.WriteByte(alphabet[(chunk>>18)&0x3F])
		result.WriteByte(alphabet[(chunk>>12)&0x3F])
		if remaining > 1 {
			result.WriteByte(alphabet[(chunk>>6)&0x3F])
		} else {
			result.WriteByte('=')
		}
		if remaining > 2 {
			result.WriteByte(alphabet[chunk&0x3F])
		} else {
			result.WriteByte('=')
		}
	}
	return result.String()
}
