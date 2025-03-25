#!/bin/bash

ALBYHUB_URL="https://getalby.com/install/hub/server-linux-aarch64.tar.bz2"
MANIFEST_URL="https://getalby.com/install/hub/manifest.txt"
SIGNATURE_URL="https://getalby.com/install/hub/manifest.txt.asc"
echo ""
echo ""
echo "⚡️ Updating Alby Hub"
echo "-----------------------------------------"
echo "This will download the latest version of Alby Hub."
echo "You will have to unlock Alby Hub after the update."
echo ""
echo "Make sure you have your unlock password available and a backup of your seed."

verify_package() {
  local archive_file="${1}"
  local overridden_manifest_name="${2}"
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
  if ! wget -q "$MANIFEST_URL"; then
    echo "❌ Failed to download manifest file." >&2
    return 1
  fi

  echo "Downloading manifest signature file..."
  if ! wget -q "$SIGNATURE_URL"; then
    echo "❌ Failed to download manifest signature file." >&2
    return 1
  fi

  if ! gpg --batch --verify "manifest.txt.asc" "manifest.txt"; then
    echo "❌ GPG signature verification failed!" >&2
    return 1
  fi

  local expected_hash
  expected_hash=$(grep "${overridden_manifest_name}" "manifest.txt" | awk '{print $1}') || true
  if [[ -z "$expected_hash" ]]; then
    echo "❌ No hash entry found for ${overridden_manifest_name} in the manifest." >&2
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

read -p "Do you want continue? (y/n):" -n 1 -r
if [[ ! $REPLY =~ ^[Yy]$ ]]
then
  exit
fi
echo ""

sudo systemctl list-units --type=service --all | grep -Fq albyhub.service
if [[ $? -eq 0 ]]; then
  echo "Stopping Alby Hub"
  sudo systemctl stop albyhub
fi

if pgrep -x "albyhub" > /dev/null
then
  echo "Alby Hub process is still running, stopping it now."
  pkill -f albyhub
fi

SCRIPT_DIR=$(dirname "$(readlink -f "$0")")
read -p "Absolute install directory path (default: $SCRIPT_DIR): " USER_INSTALL_DIR
echo ""

INSTALL_DIR="${USER_INSTALL_DIR:-$SCRIPT_DIR}"

if ! test -f $INSTALL_DIR/data/nwc.db; then
  echo "Could not find Alby Hub in this directory"
  exit 1
fi


echo "Running in $INSTALL_DIR"
# make sure we run this in the install directory
cd $INSTALL_DIR

echo "Cleaning up old backup"
rm -rf albyhub-backup
mkdir albyhub-backup

echo "Creating current backup"
mv bin albyhub-backup
mv lib albyhub-backup
cp -r data albyhub-backup


echo "Downloading latest version"
wget $ALBYHUB_URL

if ! verify_package "server-linux-aarch64.tar.bz2" "albyhub-Server-Linux-aarch64.tar.bz2"; then
  echo "❌ Verification failed, aborting installation"
  exit 1
fi

tar -xvf server-linux-aarch64.tar.bz2
rm server-linux-aarch64.tar.bz2

sudo systemctl list-units --type=service --all | grep -Fq albyhub.service
if [[ $? -eq 0 ]]; then
  echo "Starting Alby Hub"
  sudo systemctl start albyhub
fi

echo ""
echo ""
echo "✅ Update finished! Please unlock your wallet."
echo ""
