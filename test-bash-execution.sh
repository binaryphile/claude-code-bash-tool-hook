#!/bin/bash

# Integration test script for claude-code-bash-tool-hook
# Tests the hook with various bash commands to ensure correct wrapping behavior

set -e

BINARY="./bin/claude-code-bash-tool-hook"
PASSED=0
FAILED=0

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "========================================="
echo "claude-code-bash-tool-hook Integration Tests"
echo "========================================="
echo ""

# Check if binary exists
if [ ! -f "$BINARY" ]; then
    echo -e "${RED}ERROR: Binary not found at $BINARY${NC}"
    echo "Run 'make build' first"
    exit 1
fi

# Test function
test_command() {
    local description="$1"
    local command="$2"
    local should_wrap="$3"

    echo -n "Testing: $description ... "

    # Create JSON input (matching Claude Code's actual format)
    input=$(cat <<EOF
{
  "tool_name": "Bash",
  "tool_input": {
    "command": "$command"
  }
}
EOF
)

    # Run hook
    output=$(echo "$input" | "$BINARY")

    # Check if output contains updatedInput (indicating wrapping)
    if echo "$output" | grep -q "updatedInput"; then
        is_wrapped=true
    else
        is_wrapped=false
    fi

    # Verify result
    if [ "$should_wrap" = "true" ] && [ "$is_wrapped" = true ]; then
        echo -e "${GREEN}PASS${NC}"
        PASSED=$((PASSED + 1))
    elif [ "$should_wrap" = "false" ] && [ "$is_wrapped" = false ]; then
        echo -e "${GREEN}PASS${NC}"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}FAIL${NC}"
        echo "  Expected wrap=$should_wrap, got wrapped=$is_wrapped"
        echo "  Command: $command"
        echo "  Output: $output"
        FAILED=$((FAILED + 1))
    fi
}

# Test CLI mode
echo "=== CLI Mode Tests ==="
echo ""

echo "Testing --version flag..."
$BINARY --version
echo ""

echo "Testing --test mode with safe command..."
$BINARY --test "ls -la"
echo ""

echo "Testing --test mode with unsafe command..."
$BINARY --test "ls | grep foo"
echo ""

echo "=== Hook Mode Tests ==="
echo ""

# Safe commands (should NOT be wrapped)
test_command "Simple ls" "ls" false
test_command "ls with flags" "ls -la" false
test_command "cd command" "cd /tmp" false
test_command "git status" "git status" false
test_command "git log" "git log --oneline" false
test_command "echo simple" "echo hello" false
test_command "pwd" "pwd" false
test_command "plan-log" "plan-log" false

# Unsafe commands (SHOULD be wrapped)
test_command "Pipe" "ls | grep foo" true
test_command "Command substitution" "echo \$(pwd)" true
test_command "AND operator" "git status && git diff" true
test_command "OR operator" "npm install || npm update" true
test_command "Output redirect" "cat file.txt > output.txt" true
test_command "While loop" "while read line; do echo \$line; done" true
test_command "For loop" "for i in 1 2 3; do echo \$i; done" true
test_command "If statement" "if [ -f file ]; then cat file; fi" true

# Edge cases
echo ""
echo "=== Edge Case Tests ==="
echo ""

# Test escape marker
test_command "Escape marker bypass" "ls | grep foo # bypass-hook" false
test_command "No-wrap marker" "echo \$(pwd) # no-wrap" false

# Test non-Bash tool (should pass through)
echo -n "Testing non-Bash tool passthrough ... "
input='{"tool_name":"Read","tool_input":{"file_path":"/tmp/test.txt"}}'
output=$(echo "$input" | "$BINARY")
if [ "$output" = "{}" ]; then
    echo -e "${GREEN}PASS${NC}"
    PASSED=$((PASSED + 1))
else
    echo -e "${RED}FAIL${NC}"
    echo "  Expected empty JSON, got: $output"
    FAILED=$((FAILED + 1))
fi

# Test empty command
echo -n "Testing empty command ... "
input='{"tool_name":"Bash","tool_input":{"command":""}}'
output=$(echo "$input" | "$BINARY")
if [ "$output" = "{}" ]; then
    echo -e "${GREEN}PASS${NC}"
    PASSED=$((PASSED + 1))
else
    echo -e "${RED}FAIL${NC}"
    echo "  Expected passthrough, got: $output"
    FAILED=$((FAILED + 1))
fi

# Summary
echo ""
echo "========================================="
echo "Test Results"
echo "========================================="
echo -e "Passed: ${GREEN}$PASSED${NC}"
if [ $FAILED -gt 0 ]; then
    echo -e "Failed: ${RED}$FAILED${NC}"
else
    echo -e "Failed: $FAILED"
fi
echo "Total:  $((PASSED + FAILED))"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed.${NC}"
    exit 1
fi
