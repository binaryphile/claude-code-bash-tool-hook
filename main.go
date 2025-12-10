package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// HookInput represents the JSON input from Claude Code PreToolUse hook
type HookInput struct {
	ToolName  string                 `json:"tool_name"`
	ToolInput map[string]interface{} `json:"tool_input"`
}

// HookSpecificOutput is the inner output for PreToolUse hooks
type HookSpecificOutput struct {
	HookEventName      string                  `json:"hookEventName"`
	PermissionDecision string                  `json:"permissionDecision"`
	UpdatedInput       *map[string]interface{} `json:"updatedInput,omitempty"`
}

// HookOutput represents the JSON output to Claude Code
type HookOutput struct {
	HookSpecificOutput *HookSpecificOutput `json:"hookSpecificOutput,omitempty"`
}

func main() {
	// Check for CLI mode (for testing)
	if len(os.Args) > 1 {
		if os.Args[1] == "--test" && len(os.Args) > 2 {
			testCommand(os.Args[2])
			return
		} else if os.Args[1] == "--version" {
			fmt.Println("claude-code-bash-tool-hook v1.0.0")
			return
		} else if os.Args[1] == "--help" {
			printHelp()
			return
		}
	}

	// Hook mode: Read JSON from stdin
	inputBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		outputError(fmt.Sprintf("Failed to read stdin: %v", err))
		return
	}

	var input HookInput
	if err := json.Unmarshal(inputBytes, &input); err != nil {
		outputError(fmt.Sprintf("Failed to parse JSON: %v", err))
		return
	}

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		// Log error but don't block - use defaults
		LogDebug("Failed to load config, using defaults: %v", err)
		config = &Config{Enabled: true}
	}

	// Check if hook is enabled
	if !config.Enabled {
		LogDebug("Hook disabled via config")
		outputPassthrough()
		return
	}

	// Only process Bash tool calls
	if input.ToolName != "Bash" {
		LogDebug("Ignoring non-Bash tool: %s", input.ToolName)
		outputPassthrough()
		return
	}

	// Extract command parameter
	commandRaw, ok := input.ToolInput["command"]
	if !ok {
		LogDebug("No command parameter found")
		outputPassthrough()
		return
	}

	command, ok := commandRaw.(string)
	if !ok {
		LogDebug("Command parameter is not a string")
		outputPassthrough()
		return
	}

	LogDebug("Processing command: %s", command)

	if !ShouldWrapCommand(command, config) {
		LogDebug("Skipping wrap")
		outputPassthrough()
		return
	}

	wrappedCommand := WrapCommand(command)
	LogDebug("Wrapped: %s", wrappedCommand)

	// Output modified parameters using hookSpecificOutput format
	updatedParams := make(map[string]interface{})
	for k, v := range input.ToolInput {
		updatedParams[k] = v
	}
	updatedParams["command"] = wrappedCommand

	output := HookOutput{
		HookSpecificOutput: &HookSpecificOutput{
			HookEventName:      "PreToolUse",
			PermissionDecision: "allow",
			UpdatedInput:       &updatedParams,
		},
	}

	outputJSON, err := json.Marshal(output)
	if err != nil {
		outputError(fmt.Sprintf("Failed to marshal output: %v", err))
		return
	}

	LogDebug("Output JSON: %s", string(outputJSON))
	fmt.Println(string(outputJSON))
}

// outputPassthrough outputs empty JSON (no modifications)
func outputPassthrough() {
	fmt.Println("{}")
}

// outputError logs an error and outputs empty JSON (passthrough)
func outputError(msg string) {
	LogDebug("ERROR: %s", msg)
	// On error, passthrough (don't block) by outputting empty JSON
	fmt.Println("{}")
}

// testCommand is a CLI mode for testing command wrapping
func testCommand(command string) {
	config, _ := LoadConfig()
	if config == nil {
		config = &Config{Enabled: true}
	}

	if ShouldWrapCommand(command, config) {
		fmt.Printf("%s\n", WrapCommand(command))
	} else {
		fmt.Printf("%s\n", command)
	}
}

// printHelp displays usage information
func printHelp() {
	help := `claude-code-bash-tool-hook - PreToolUse hook for Claude Code

USAGE:
    claude-code-bash-tool-hook                    Hook mode (read JSON from stdin)
    claude-code-bash-tool-hook --test "command"   Test mode (check if command would be wrapped)
    claude-code-bash-tool-hook --version          Show version
    claude-code-bash-tool-hook --help             Show this help

HOOK MODE:
    Reads PreToolUse hook JSON from stdin, wraps bash commands that need it,
    and outputs modified JSON to stdout.

TEST MODE:
    Tests whether a command would be wrapped and shows the result.
    Example: claude-code-bash-tool-hook --test "ls | grep foo"

CONFIGURATION:
    Config file: ~/.claude/bash-hook-config.json
    See README.md for configuration options.

DOCUMENTATION:
    See README.md for full documentation.
`
	fmt.Print(help)
}
