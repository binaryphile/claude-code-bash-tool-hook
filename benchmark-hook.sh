#!/bin/bash

# Benchmark script for claude-code-bash-tool-hook
# Measures hook startup time and processing overhead

set -e

BINARY="./bin/claude-code-bash-tool-hook"
ITERATIONS=100

# Colors
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
NC='\033[0m'

echo "========================================="
echo "claude-code-bash-tool-hook Performance Benchmark"
echo "========================================="
echo ""

# Check if binary exists
if [ ! -f "$BINARY" ]; then
    echo "ERROR: Binary not found at $BINARY"
    echo "Run 'make build' first"
    exit 1
fi

# Prepare test input
INPUT='{"tool":"Bash","parameters":{"command":"ls | grep foo"}}'

echo "Running $ITERATIONS iterations..."
echo ""

# Measure total time
start_time=$(date +%s%N)

for i in $(seq 1 $ITERATIONS); do
    echo "$INPUT" | "$BINARY" > /dev/null
done

end_time=$(date +%s%N)

# Calculate statistics
total_ns=$((end_time - start_time))
total_ms=$((total_ns / 1000000))
avg_ns=$((total_ns / ITERATIONS))
avg_ms=$((avg_ns / 1000000))
avg_us=$((avg_ns / 1000))

echo "Results:"
echo "--------"
echo "Total time:    ${total_ms} ms"
echo "Iterations:    $ITERATIONS"
echo "Average time:  ${avg_ms} ms (${avg_us} μs)"
echo ""

# Check performance target
if [ $avg_ms -lt 10 ]; then
    echo -e "${GREEN}✓ Performance target met (< 10ms startup)${NC}"
    echo ""
else
    echo -e "${YELLOW}⚠ Performance target not met (>= 10ms startup)${NC}"
    echo "  Target: < 10ms"
    echo "  Actual: ${avg_ms}ms"
    echo ""
fi

# Single iteration detailed measurement
echo "Single iteration timing:"
echo "------------------------"
/usr/bin/time -f "Real: %E\nUser: %U\nSys:  %S" bash -c "echo '$INPUT' | $BINARY > /dev/null" 2>&1
echo ""

echo "Binary size:"
echo "------------"
ls -lh "$BINARY" | awk '{print $5 " (" $9 ")"}'
echo ""

echo "Benchmark complete."
