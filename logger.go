package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	maxLogSize  = 1024 * 1024 // 1MB
	maxLogLines = 1000
)

// SecretPatterns are regex patterns for redacting sensitive information
var SecretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(api[_-]?key|apikey)[\s=:]+['"']?([a-zA-Z0-9_-]{16,})['"']?`),
	regexp.MustCompile(`(token|bearer)[\s=:]+['"']?([a-zA-Z0-9_.-]{16,})['"']?`),
	regexp.MustCompile(`(password|passwd|pwd)[\s=:]+['"']?([^\s'"]{8,})['"']?`),
	regexp.MustCompile(`(secret|auth)[\s=:]+['"']?([a-zA-Z0-9_-]{16,})['"']?`),
	regexp.MustCompile(`sk-[a-zA-Z0-9]{32,}`), // OpenAI API keys
	regexp.MustCompile(`ghp_[a-zA-Z0-9]{36,}`), // GitHub personal access tokens
}

// LogDebug writes a debug message to the log file if debug logging is enabled
func LogDebug(format string, args ...interface{}) {
	config, err := LoadConfig()
	if err != nil || config == nil || !config.DebugLog {
		return
	}

	logFile := config.LogFile
	if logFile == "" {
		homeDir, _ := os.UserHomeDir()
		logFile = filepath.Join(homeDir, ".claude", "bash-hook-debug.log")
	}

	// Ensure log directory exists
	logDir := filepath.Dir(logFile)
	os.MkdirAll(logDir, 0755)

	// Format message
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	message := fmt.Sprintf(format, args...)

	// Redact secrets
	message = redactSecrets(message)

	logLine := fmt.Sprintf("[%s] %s\n", timestamp, message)

	// Check if log rotation is needed
	rotateLogIfNeeded(logFile)

	// Append to log file
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	defer f.Close()

	f.WriteString(logLine)
}

// redactSecrets replaces sensitive information with [REDACTED]
func redactSecrets(message string) string {
	result := message
	for _, pattern := range SecretPatterns {
		result = pattern.ReplaceAllStringFunc(result, func(match string) string {
			// Keep the key name but redact the value
			parts := pattern.FindStringSubmatch(match)
			if len(parts) >= 2 {
				keyName := parts[1]
				return fmt.Sprintf("%s=[REDACTED]", keyName)
			}
			return "[REDACTED]"
		})
	}
	return result
}

// rotateLogIfNeeded checks if log file exceeds limits and rotates if necessary
func rotateLogIfNeeded(logFile string) {
	// Check file size
	info, err := os.Stat(logFile)
	if err != nil {
		return // File doesn't exist yet
	}

	if info.Size() < maxLogSize {
		return // No rotation needed
	}

	// Read all lines
	f, err := os.Open(logFile)
	if err != nil {
		return
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Keep only the last maxLogLines
	if len(lines) > maxLogLines {
		lines = lines[len(lines)-maxLogLines:]
	}

	// Write back truncated log
	content := strings.Join(lines, "\n") + "\n"
	os.WriteFile(logFile, []byte(content), 0600)
}
