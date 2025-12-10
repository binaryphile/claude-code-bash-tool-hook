# claude-code-bash-tool-hook

A PreToolUse hook for Claude Code that fixes bash preprocessing bugs by wrapping commands in base64-encoded `bash -c`.

## Problem

Claude Code has preprocessing bugs ([GitHub Issue #11225](https://github.com/anthropics/claude-code/issues/11225)) where bash commands fail silently:

```bash
echo $USER | cat          # Returns empty
for i in a b; do echo $i; done  # Variables stripped
```

## Solution

This hook wraps all commands in `bash -c "$(echo 'BASE64' | base64 -d)"` to bypass preprocessing entirely. Base64 encoding eliminates all quoting edge cases.

## Installation

```bash
# Build
go build -o claude-code-bash-tool-hook

# Install
mkdir -p ~/.claude/hooks
cp claude-code-bash-tool-hook ~/.claude/hooks/
```

Add to `~/.claude/settings.json`:

```json
{
  "hooks": {
    "PreToolUse": [{
      "matcher": "Bash",
      "hooks": [{
        "type": "command",
        "command": "~/.claude/hooks/claude-code-bash-tool-hook",
        "timeout": 5
      }]
    }]
  }
}
```

## How It Works

**Input:** `echo $USER | cat`

**Output:** `bash -c "$(echo 'ZWNobyAkVVNFUiB8IGNhdA==' | base64 -d)"`

All commands are wrapped except:
- Empty commands
- Already wrapped (`bash -c ...`)
- Escape markers (`# bypass-hook`, `# no-wrap`, `# skip-hook`)

## Configuration

Config file: `~/.claude/bash-hook-config.json`

```json
{
  "enabled": true,
  "debug_log": false,
  "additional_escape_markers": ["# my-marker"]
}
```

## CLI Testing

```bash
./claude-code-bash-tool-hook --test "echo \$USER | cat"
# Output: bash -c "$(echo 'ZWNobyAkVVNFUiB8IGNhdA==' | base64 -d)"
```

## Development

```bash
go test -v      # Run tests
go build        # Build
```

## References

- [GitHub Issue #11225](https://github.com/anthropics/claude-code/issues/11225)
- [Claude Code Hooks Documentation](https://docs.anthropic.com/en/docs/claude-code/hooks)
