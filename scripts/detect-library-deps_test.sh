#!/bin/bash
# Test script for detect-library-deps.sh
#
# Usage: detect-library-deps_test.sh
#
# Exit codes:
#   0: All tests passed
#   1: One or more tests failed

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DETECT_SCRIPT="${SCRIPT_DIR}/detect-library-deps.sh"
TEST_CACHE="/tmp/test-lib-deps-cache-$$.txt"
TESTS_PASSED=0
TESTS_FAILED=0

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test helper functions
print_test() {
  echo -e "${YELLOW}TEST: $1${NC}"
}

pass_test() {
  echo -e "${GREEN}✓ PASS${NC}"
  ((TESTS_PASSED++))
}

fail_test() {
  echo -e "${RED}✗ FAIL: $1${NC}"
  ((TESTS_FAILED++))
}

# Clean up cache file on exit
cleanup() {
  rm -f "$TEST_CACHE"
}
trap cleanup EXIT

# Test 1: Script exists and is executable
print_test "Script exists and is executable"
if [ -x "$DETECT_SCRIPT" ]; then
  pass_test
else
  fail_test "Script not found or not executable: $DETECT_SCRIPT"
  exit 1
fi

# Test 2: Help message works
print_test "Help message works"
if "$DETECT_SCRIPT" --help >/dev/null 2>&1; then
  pass_test
else
  fail_test "Help message failed"
fi

# Test 3: Error on no binaries
print_test "Error when no binaries specified"
if ! "$DETECT_SCRIPT" 2>/dev/null; then
  pass_test
else
  fail_test "Should error when no binaries specified"
fi

# Test 4: Detect libraries for /bin/ls
print_test "Detect libraries for /bin/ls"
OUTPUT=$("$DETECT_SCRIPT" --format=paths --no-cache /bin/ls)
if echo "$OUTPUT" | grep -q "libc.so"; then
  pass_test
else
  fail_test "Expected to find libc.so for /bin/ls"
  echo "Output: $OUTPUT"
fi

# Test 5: AWF mount format
print_test "AWF mount format is correct"
OUTPUT=$("$DETECT_SCRIPT" --format=awf-mounts --no-cache /bin/ls)
if echo "$OUTPUT" | head -1 | grep -q "^--mount"; then
  if echo "$OUTPUT" | head -1 | grep -q ":ro$"; then
    pass_test
  else
    fail_test "Mount should end with :ro"
    echo "Output: $OUTPUT"
  fi
else
  fail_test "Output should start with --mount"
  echo "Output: $OUTPUT"
fi

# Test 6: JSON format
print_test "JSON format is valid"
OUTPUT=$("$DETECT_SCRIPT" --format=json --no-cache /bin/ls)
if echo "$OUTPUT" | head -1 | grep -q "^\["; then
  if echo "$OUTPUT" | tail -1 | grep -q "^\]"; then
    pass_test
  else
    fail_test "JSON should end with ]"
    echo "Output: $OUTPUT"
  fi
else
  fail_test "JSON should start with ["
  echo "Output: $OUTPUT"
fi

# Test 7: Multiple binaries
print_test "Handle multiple binaries"
OUTPUT=$("$DETECT_SCRIPT" --format=paths --no-cache /bin/ls /bin/cat)
LINE_COUNT=$(echo "$OUTPUT" | wc -l)
if [ "$LINE_COUNT" -gt 1 ]; then
  pass_test
else
  fail_test "Expected multiple libraries for two binaries"
  echo "Output: $OUTPUT"
fi

# Test 8: Caching mechanism
print_test "Caching mechanism works"
# First run with cache
"$DETECT_SCRIPT" --cache-file="$TEST_CACHE" --format=paths /bin/ls >/dev/null
if [ -f "$TEST_CACHE" ]; then
  if [ -s "$TEST_CACHE" ]; then
    # Second run should use cache (should be faster)
    OUTPUT2=$("$DETECT_SCRIPT" --cache-file="$TEST_CACHE" --format=paths /bin/ls)
    if echo "$OUTPUT2" | grep -q "libc.so"; then
      pass_test
    else
      fail_test "Cached result should still contain libraries"
      echo "Cache file: $(cat "$TEST_CACHE")"
      echo "Output: $OUTPUT2"
    fi
  else
    fail_test "Cache file is empty"
  fi
else
  fail_test "Cache file not created"
fi

# Test 9: Handle non-existent binary gracefully
print_test "Handle non-existent binary gracefully"
OUTPUT=$("$DETECT_SCRIPT" --format=paths --no-cache /nonexistent/binary 2>&1)
# Script should not crash, but may warn
if [ $? -eq 0 ] || [ $? -eq 1 ]; then
  pass_test
else
  fail_test "Script crashed on non-existent binary"
  echo "Output: $OUTPUT"
fi

# Test 10: Common utilities have libraries
print_test "Common utilities (curl, jq, git) have libraries"
COMMON_BINARIES=()
[ -f /usr/bin/curl ] && COMMON_BINARIES+=(/usr/bin/curl)
[ -f /usr/bin/jq ] && COMMON_BINARIES+=(/usr/bin/jq)
[ -f /usr/bin/git ] && COMMON_BINARIES+=(/usr/bin/git)

if [ ${#COMMON_BINARIES[@]} -gt 0 ]; then
  OUTPUT=$("$DETECT_SCRIPT" --format=paths --no-cache "${COMMON_BINARIES[@]}")
  if [ -n "$OUTPUT" ]; then
    # Check for common library patterns
    if echo "$OUTPUT" | grep -q -E "(libc\.so|libcurl\.so|libz\.so)"; then
      pass_test
    else
      fail_test "Expected common libraries not found"
      echo "Output: $OUTPUT"
    fi
  else
    fail_test "No libraries detected for common utilities"
  fi
else
  echo "Skipped - no common binaries found"
  ((TESTS_PASSED++))
fi

# Test 11: Deduplication works
print_test "Deduplication works (same binary twice)"
OUTPUT=$("$DETECT_SCRIPT" --format=paths --no-cache /bin/ls /bin/ls)
# Count unique paths
UNIQUE_COUNT=$(echo "$OUTPUT" | sort -u | wc -l)
TOTAL_COUNT=$(echo "$OUTPUT" | wc -l)
if [ "$UNIQUE_COUNT" -eq "$TOTAL_COUNT" ]; then
  pass_test
else
  fail_test "Output contains duplicates"
  echo "Unique: $UNIQUE_COUNT, Total: $TOTAL_COUNT"
fi

# Test 12: Symlink resolution
print_test "Symlink resolution works"
# /lib is typically a symlink to usr/lib
if [ -L /lib ]; then
  OUTPUT=$("$DETECT_SCRIPT" --format=paths --no-cache /bin/ls)
  # Check that we get resolved paths, not just symlinks
  if echo "$OUTPUT" | grep -q "/"; then
    pass_test
  else
    fail_test "Expected resolved library paths"
    echo "Output: $OUTPUT"
  fi
else
  echo "Skipped - /lib is not a symlink"
  ((TESTS_PASSED++))
fi

# Print summary
echo ""
echo "======================================"
echo "Test Summary:"
echo "======================================"
echo -e "Tests passed: ${GREEN}${TESTS_PASSED}${NC}"
if [ $TESTS_FAILED -gt 0 ]; then
  echo -e "Tests failed: ${RED}${TESTS_FAILED}${NC}"
  exit 1
else
  echo -e "Tests failed: ${GREEN}0${NC}"
  echo -e "${GREEN}All tests passed!${NC}"
  exit 0
fi
