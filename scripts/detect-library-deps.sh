#!/bin/bash
# Script to dynamically detect library dependencies for mounted binaries
# and generate Docker mount arguments.
#
# Usage: detect-library-deps.sh [OPTIONS] <binary1> [binary2 ...]
#
# Options:
#   --cache-file FILE    Use FILE for caching library dependencies (default: /tmp/lib-deps-cache.txt)
#   --no-cache           Disable caching (always run ldd)
#   --format FORMAT      Output format: awf-mounts (default), paths, or json
#   --help               Show this help message
#
# Output formats:
#   awf-mounts: Generates --mount arguments for AWF (e.g., --mount /lib/x86_64-linux-gnu/libc.so.6:/lib/x86_64-linux-gnu/libc.so.6:ro)
#   paths:      Just the library paths, one per line
#   json:       JSON array of library paths
#
# Examples:
#   detect-library-deps.sh /usr/bin/curl /usr/bin/jq
#   detect-library-deps.sh --format=paths /usr/bin/gh
#   detect-library-deps.sh --no-cache /usr/bin/git
#
# Exit codes:
#   0: Success
#   1: Error (missing binary, ldd failed, etc.)

set -euo pipefail

# Default options
CACHE_FILE="/tmp/lib-deps-cache.txt"
USE_CACHE=true
OUTPUT_FORMAT="awf-mounts"
BINARIES=()

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case "$1" in
    --cache-file)
      CACHE_FILE="$2"
      shift 2
      ;;
    --cache-file=*)
      CACHE_FILE="${1#*=}"
      shift
      ;;
    --no-cache)
      USE_CACHE=false
      shift
      ;;
    --format)
      OUTPUT_FORMAT="$2"
      shift 2
      ;;
    --format=*)
      OUTPUT_FORMAT="${1#*=}"
      shift
      ;;
    --help|-h)
      grep '^#' "$0" | grep -v '^#!/' | sed 's/^# \?//'
      exit 0
      ;;
    -*)
      echo "Error: Unknown option: $1" >&2
      echo "Use --help for usage information" >&2
      exit 1
      ;;
    *)
      BINARIES+=("$1")
      shift
      ;;
  esac
done

# Validate inputs
if [ ${#BINARIES[@]} -eq 0 ]; then
  echo "Error: No binaries specified" >&2
  echo "Use --help for usage information" >&2
  exit 1
fi

# Validate output format
case "$OUTPUT_FORMAT" in
  awf-mounts|paths|json) ;;
  *)
    echo "Error: Invalid output format: $OUTPUT_FORMAT" >&2
    echo "Valid formats: awf-mounts, paths, json" >&2
    exit 1
    ;;
esac

# Function to resolve symlinks to their final target
resolve_symlinks() {
  local path="$1"
  local resolved
  
  # Use readlink -f to fully resolve symlinks
  # If readlink fails (e.g., file doesn't exist), return the original path
  if resolved=$(readlink -f "$path" 2>/dev/null); then
    echo "$resolved"
  else
    echo "$path"
  fi
}

# Function to extract library dependencies for a single binary
extract_lib_deps() {
  local binary="$1"
  local cache_key
  local cached_result
  
  # Check if binary exists
  if [ ! -f "$binary" ]; then
    echo "Warning: Binary not found: $binary" >&2
    return 1
  fi
  
  # Generate cache key (binary path + modification time)
  cache_key="${binary}:$(stat -c '%Y' "$binary" 2>/dev/null || echo 0)"
  
  # Check cache if enabled
  if [ "$USE_CACHE" = true ] && [ -f "$CACHE_FILE" ]; then
    cached_result=$(grep "^${cache_key}=" "$CACHE_FILE" 2>/dev/null | cut -d= -f2- || echo "")
    if [ -n "$cached_result" ]; then
      echo "$cached_result"
      return 0
    fi
  fi
  
  # Run ldd to get library dependencies
  local ldd_output
  if ! ldd_output=$(ldd "$binary" 2>&1); then
    # Some binaries are statically linked or may not work with ldd
    # This is not necessarily an error, just return empty
    return 0
  fi
  
  # Parse ldd output to extract library paths
  # Format examples:
  #   linux-vdso.so.1 (0x00007fff...) -> skip (virtual)
  #   libcurl.so.4 => /lib/x86_64-linux-gnu/libcurl.so.4 (0x00007f...)
  #   /lib64/ld-linux-x86-64.so.2 (0x00007f...)
  local libs=()
  while IFS= read -r line; do
    # Skip virtual libraries (vdso, vsyscall)
    if echo "$line" | grep -q "vdso\|vsyscall"; then
      continue
    fi
    
    # Extract library path
    local lib_path=""
    if echo "$line" | grep -q " => "; then
      # Format: libname => /path/to/lib (address)
      lib_path=$(echo "$line" | sed -n 's/.*=> \([^ ]*\) .*/\1/p')
    elif echo "$line" | grep -q "^[[:space:]]*\/"; then
      # Format: /path/to/lib (address)
      lib_path=$(echo "$line" | sed -n 's/^[[:space:]]*\([^ ]*\) .*/\1/p')
    fi
    
    # Validate library path
    if [ -n "$lib_path" ] && [ -f "$lib_path" ]; then
      # Resolve symlinks to get the actual library file
      local resolved_path
      resolved_path=$(resolve_symlinks "$lib_path")
      
      # Add both the symlink and the resolved path
      # This ensures compatibility with binaries that reference either
      if [ "$resolved_path" != "$lib_path" ]; then
        libs+=("$lib_path")
      fi
      libs+=("$resolved_path")
    fi
  done <<< "$ldd_output"
  
  # Sort and deduplicate libraries
  local unique_libs
  unique_libs=$(printf '%s\n' "${libs[@]}" | sort -u)
  
  # Cache the result if caching is enabled
  if [ "$USE_CACHE" = true ]; then
    mkdir -p "$(dirname "$CACHE_FILE")"
    echo "${cache_key}=${unique_libs}" >> "$CACHE_FILE"
  fi
  
  echo "$unique_libs"
}

# Collect all library dependencies
ALL_LIBS=()
for binary in "${BINARIES[@]}"; do
  # Extract dependencies for this binary
  deps=$(extract_lib_deps "$binary")
  
  # Add to global list
  while IFS= read -r lib; do
    if [ -n "$lib" ]; then
      ALL_LIBS+=("$lib")
    fi
  done <<< "$deps"
done

# Sort and deduplicate all libraries
UNIQUE_LIBS=($(printf '%s\n' "${ALL_LIBS[@]}" | sort -u))

# Output results in the requested format
case "$OUTPUT_FORMAT" in
  awf-mounts)
    # Generate --mount arguments for AWF
    for lib in "${UNIQUE_LIBS[@]}"; do
      echo "--mount" "${lib}:${lib}:ro"
    done
    ;;
  paths)
    # Just output the paths
    printf '%s\n' "${UNIQUE_LIBS[@]}"
    ;;
  json)
    # Output as JSON array
    echo "["
    first=true
    for lib in "${UNIQUE_LIBS[@]}"; do
      if [ "$first" = true ]; then
        first=false
      else
        echo ","
      fi
      echo -n "  \"$lib\""
    done
    echo
    echo "]"
    ;;
esac

exit 0
