#!/bin/bash

MANIFEST_URL="https://getalby.com/install/hub/manifest.txt"
SIGNATURE_URL="https://getalby.com/install/hub/manifest.txt.asc"

verify_package() {
  local archive_file="${1}"
  local filename_in_manifest="${2}"
  local response=""

  while true; do
    read -r -p "Verify package signature and integrity? (Y/N): " response
    case "$response" in
      [Yy]) break ;;
      [Nn]) echo "Verification skipped." ; return 0 ;;
      *) echo "Invalid input. Please enter Y or N." ;;
    esac
  done

  for cmd in gpg sha256sum; do
    if ! command -v "$cmd" &>/dev/null; then
      echo "❌ Required command '$cmd' is not available." >&2
      return 1
    fi
  done

  echo "Downloading manifest file..."
  if ! wget -q -O manifest.txt "$MANIFEST_URL"; then
    echo "❌ Failed to download manifest file." >&2
    return 1
  fi

  echo "Downloading manifest signature file..."
  if ! wget -q -O manifest.txt.asc "$SIGNATURE_URL"; then
    echo "❌ Failed to download manifest signature file." >&2
    return 1
  fi

  if ! gpg --batch --verify "manifest.txt.asc" "manifest.txt"; then
    echo "❌ GPG signature verification failed!" >&2
    echo "Visit https://github.com/getAlby/hub/releases for more information on how to verify the release" >&2
    return 1
  fi

  local expected_hash
  expected_hash=$(grep "${filename_in_manifest}" "manifest.txt" | awk '{print $1}') || true
  if [[ -z "$expected_hash" ]]; then
    echo "❌ No hash entry found for ${filename_in_manifest} in the manifest." >&2
    return 1
  fi

  local actual_hash
  actual_hash=$(sha256sum "$archive_file" | awk '{print $1}')

  if [[ "$expected_hash" != "$actual_hash" ]]; then
    echo "❌ SHA256 hash mismatch! The file may be corrupted or tampered with." >&2
    return 1
  fi

  echo "✅ Verification successful. The package is authentic and intact."
  return 0
}

if [[ $# -ne 2 ]]; then
  echo "Usage: $0 <archive_file> <filename_in_manifest>"
  exit 1
fi

verify_package "$1" "$2"
if [[ $? -ne 0 ]]; then
  exit 1
fi
