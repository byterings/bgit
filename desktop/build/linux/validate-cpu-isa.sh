#!/bin/sh
# Fail release builds when the Linux AMD64 desktop executable advertises an
# ISA requirement newer than the x86-64 baseline.
set -eu

if [ "$#" -ne 1 ]; then
  echo "usage: $0 <linux-amd64-binary>" >&2
  exit 2
fi

BINARY="$1"
if [ ! -f "$BINARY" ]; then
  echo "binary not found: $BINARY" >&2
  exit 1
fi

if ! command -v readelf >/dev/null 2>&1; then
  echo "readelf is required to validate the Linux desktop binary" >&2
  exit 1
fi

ELF_HEADER="$(readelf -h "$BINARY")"
ELF_CLASS="$(printf '%s\n' "$ELF_HEADER" | awk -F: '/Class:/ {gsub(/^[[:space:]]+/, "", $2); print $2}')"
ELF_MACHINE="$(printf '%s\n' "$ELF_HEADER" | awk -F: '/Machine:/ {gsub(/^[[:space:]]+/, "", $2); print $2}')"

if [ "$ELF_CLASS" != "ELF64" ] || [ "$ELF_MACHINE" != "Advanced Micro Devices X86-64" ]; then
  echo "expected a 64-bit x86-64 ELF executable; found class='$ELF_CLASS' machine='$ELF_MACHINE'" >&2
  exit 1
fi

ISA_REQUIREMENTS="$(readelf -n "$BINARY" | sed -n 's/.*x86 ISA needed: //p')"
if [ "$ISA_REQUIREMENTS" != "x86-64-baseline" ]; then
  echo "expected only x86-64-baseline ISA requirement; found '${ISA_REQUIREMENTS:-no GNU x86 ISA requirement note}'" >&2
  exit 1
fi

echo "CPU ISA validation passed: $BINARY is x86-64 and requires only x86-64-baseline"
