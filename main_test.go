package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestShouldWrapCommand_WrapsEverything(t *testing.T) {
	config := &Config{Enabled: true}

	commands := []string{
		"cd /home/user", "pwd", "ls -la", "git status", "echo hello",
		"ls | grep test", "echo $(pwd)", "git status && git diff",
		"for i in 1 2 3; do echo $i; done",
	}

	for _, cmd := range commands {
		if !ShouldWrapCommand(cmd, config) {
			t.Errorf("should wrap: %s", cmd)
		}
	}
}

func TestShouldWrapCommand_EmptyCommand(t *testing.T) {
	config := &Config{Enabled: true}

	for _, cmd := range []string{"", "   ", "\t", "\n"} {
		if ShouldWrapCommand(cmd, config) {
			t.Errorf("should not wrap empty: %q", cmd)
		}
	}
}

func TestShouldWrapCommand_AlreadyWrapped(t *testing.T) {
	config := &Config{Enabled: true}

	for _, cmd := range []string{
		"bash -c 'echo hello'",
		"bash -c \"ls | grep test\"",
		`bash -c "$(echo 'dGVzdA==' | base64 -d)"`,
	} {
		if ShouldWrapCommand(cmd, config) {
			t.Errorf("should not re-wrap: %s", cmd)
		}
	}
}

func TestShouldWrapCommand_EscapeMarkers(t *testing.T) {
	config := &Config{Enabled: true}

	for _, cmd := range []string{
		"ls | grep foo # bypass-hook",
		"echo $(pwd) # no-wrap",
		"cat file.txt # skip-hook",
	} {
		if ShouldWrapCommand(cmd, config) {
			t.Errorf("should not wrap with escape marker: %s", cmd)
		}
	}
}

func TestShouldWrapCommand_CustomEscapeMarkers(t *testing.T) {
	config := &Config{
		Enabled:                 true,
		AdditionalEscapeMarkers: []string{"# custom-skip"},
	}

	if ShouldWrapCommand("ls | grep foo # custom-skip", config) {
		t.Error("should not wrap with custom escape marker")
	}
}

func TestWrapCommand_Base64Format(t *testing.T) {
	tests := []string{
		`echo "hello"`,
		`echo $HOME`,
		"echo `pwd`",
		`echo \n`,
		`echo 'hello'`,
		"echo \"line1\nline2\"", // newline
		`echo "quotes" && echo 'more'`,
	}

	for _, input := range tests {
		wrapped := WrapCommand(input)
		// Verify format: bash -c "$(echo 'BASE64' | base64 -d)"
		if !strings.HasPrefix(wrapped, `bash -c "$(echo '`) {
			t.Errorf("WrapCommand(%q) should start with base64 format: %s", input, wrapped)
		}
		if !strings.HasSuffix(wrapped, `' | base64 -d)"`) {
			t.Errorf("WrapCommand(%q) should end with base64 decode: %s", input, wrapped)
		}
	}
}

func TestBase64Encode_Roundtrip(t *testing.T) {
	tests := []string{
		"hello",
		"echo $USER | cat",
		"line1\nline2\nline3",
		"single'quote",
		"double\"quote",
		"back`tick",
		"special!@#$%^&*()",
		"",
	}

	for _, input := range tests {
		encoded := base64Encode(input)
		// Verify only safe characters
		for _, c := range encoded {
			if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '+' || c == '/' || c == '=') {
				t.Errorf("base64Encode(%q) contains unsafe char %q: %s", input, c, encoded)
			}
		}
	}
}

func TestBase64Encode_MatchesStdlib(t *testing.T) {
	// Verify our encoder produces same output as encoding/base64
	import64 := "encoding/base64"
	_ = import64 // suppress unused warning in comment

	tests := []string{
		"hello",
		"test string",
		"with\nnewlines\n",
		"special chars: !@#$%",
	}

	for _, input := range tests {
		got := base64Encode(input)
		// Expected values computed from encoding/base64.StdEncoding
		expected := map[string]string{
			"hello":                 "aGVsbG8=",
			"test string":           "dGVzdCBzdHJpbmc=",
			"with\nnewlines\n":      "d2l0aApuZXdsaW5lcwo=",
			"special chars: !@#$%": "c3BlY2lhbCBjaGFyczogIUAjJCU=",
		}[input]
		if got != expected {
			t.Errorf("base64Encode(%q) = %q, want %q", input, got, expected)
		}
	}
}

func TestWrapCommand_ComplexCommands(t *testing.T) {
	tests := []string{
		`ls | grep "test"`,
		`for i in $(seq 1 10); do echo $i; done`,
		`git log --format="%h %s" | head -5`,
		`echo "value: ${HOME}"`,
	}

	for _, input := range tests {
		wrapped := WrapCommand(input)
		if !strings.HasPrefix(wrapped, `bash -c "$(echo '`) {
			t.Errorf("Wrapped command should start with base64 format: %s", wrapped)
		}
		if !strings.HasSuffix(wrapped, `' | base64 -d)"`) {
			t.Errorf("Wrapped command should end with base64 decode: %s", wrapped)
		}
	}
}

// TestIssue11225Patterns - regression tests for GitHub issue #11225
func TestIssue11225Patterns(t *testing.T) {
	config := &Config{Enabled: true}

	for _, cmd := range []string{
		`for i in one two three; do echo "Item: $i" | cat; done`,
		`result=$(echo "test" | tr a-z A-Z); echo "Result: $result"`,
		`export TEST_VAR=alpha && echo "$TEST_VAR" | hexdump -C`,
		`echo $USER | cat`,
	} {
		if !ShouldWrapCommand(cmd, config) {
			t.Errorf("should wrap: %s", cmd)
		}
	}
}

// TestHookOutputFormat verifies the hookSpecificOutput JSON format
func TestHookOutputFormat(t *testing.T) {
	output := HookOutput{
		HookSpecificOutput: &HookSpecificOutput{
			HookEventName:      "PreToolUse",
			PermissionDecision: "allow",
			UpdatedInput:       &map[string]interface{}{"command": "test"},
		},
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	jsonStr := string(data)
	if !strings.Contains(jsonStr, `"hookSpecificOutput"`) {
		t.Error("Output should contain hookSpecificOutput")
	}
	if !strings.Contains(jsonStr, `"hookEventName":"PreToolUse"`) {
		t.Error("Output should contain hookEventName")
	}
	if !strings.Contains(jsonStr, `"permissionDecision":"allow"`) {
		t.Error("Output should contain permissionDecision")
	}
	if !strings.Contains(jsonStr, `"updatedInput"`) {
		t.Error("Output should contain updatedInput")
	}
}

func TestRedactSecrets(t *testing.T) {
	tests := []struct {
		input    string
		contains string
	}{
		{"api_key=sk-1234567890abcdef", "[REDACTED]"},
		{"token: ghp_1234567890abcdef", "[REDACTED]"},
		{"password=mysecretpassword123", "[REDACTED]"},
	}

	for _, test := range tests {
		result := redactSecrets(test.input)
		if !strings.Contains(result, test.contains) {
			t.Errorf("redactSecrets(%q) should contain %q, got: %q", test.input, test.contains, result)
		}
	}
}

func TestDefaultConfigPath(t *testing.T) {
	path := DefaultConfigPath()
	if path == "" {
		t.Skip("Could not determine home directory")
	}
	if !strings.Contains(path, ".claude") {
		t.Errorf("Config path should contain .claude: %s", path)
	}
	if !strings.HasSuffix(path, "bash-hook-config.json") {
		t.Errorf("Config path should end with bash-hook-config.json: %s", path)
	}
}

func TestLoadConfig_DefaultWhenMissing(t *testing.T) {
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() should not error on missing file: %v", err)
	}
	if config == nil {
		t.Fatal("LoadConfig() should return default config when file missing")
	}
	if !config.Enabled {
		t.Error("Default config should have Enabled=true")
	}
}
