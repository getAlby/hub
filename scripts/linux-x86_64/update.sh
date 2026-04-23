#!/bin/sh

ALBYHUB_URL="https://getalby.com/install/hub/server-linux-x86_64.tar.bz2"

# Default values
INSTALL_DIR=""
NON_INTERACTIVE=false
SKIP_VERIFY=false

# Parse command-line arguments
while [ $# -gt 0 ]; do
  case "$1" in
    -d|--install-dir)
      if [ -z "$2" ] || [ "${2#-}" != "$2" ]; then
        echo "Error: --install-dir requires a non-empty directory path"
        exit 1
      fi
      INSTALL_DIR="$2"
      shift 2
      ;;
    -y|--yes)
      NON_INTERACTIVE=true
      shift
      ;;
    --skip-verify)
      SKIP_VERIFY=true
      shift
      ;;
    -h|--help)
      echo "Usage: $0 [OPTIONS]"
      echo ""
      echo "Options:"
      echo "  -d, --install-dir DIR    Set installation directory (auto-detected by default)"
      echo "  -y, --yes                Non-interactive mode (skip confirmation prompt)"
      echo "      --skip-verify        Skip package signature verification"
      echo "  -h, --help               Show this help message"
      echo ""
      echo "Examples:"
      echo "  $0                      # Interactive mode"
      echo "  $0 -y                   # Skip confirmation prompt"
      echo "  $0 -y -d /opt/albyhub   # Non-interactive with specific directory"
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      echo "Use -h or --help for usage information"
      exit 1
      ;;
  esac
done

echo ""
echo ""
echo "⚡️ Updating Alby Hub"
echo "-----------------------------------------"
echo "This will download the latest version of Alby Hub."
echo "You will have to unlock Alby Hub after the update."
echo ""
echo "Make sure you have your unlock password available and a backup of your seed."

# Confirmation prompt
if [ "$NON_INTERACTIVE" = false ]; then
  printf "Do you want to continue? (y/n): "
  read REPLY
  case "$REPLY" in
    [Yy]*) ;;
    *) exit ;;
  esac
  echo ""
fi

if sudo systemctl list-units --type=service --all | grep -Fq albyhub.service; then
  echo "Stopping Alby Hub"
  sudo systemctl stop albyhub
fi

if pgrep -x "albyhub" > /dev/null; then
  echo "Alby Hub process is still running, stopping it now."
  pkill -f albyhub
fi

SCRIPT_DIR=$(dirname "$(readlink -f "$0")")

# Determine install directory
if [ -z "$INSTALL_DIR" ]; then
  if [ "$NON_INTERACTIVE" = true ]; then
    INSTALL_DIR="$SCRIPT_DIR"
  else
    printf "Absolute install directory path (default: %s): " "$SCRIPT_DIR"
    read USER_INSTALL_DIR
    echo ""
    INSTALL_DIR="${USER_INSTALL_DIR:-$SCRIPT_DIR}"
  fi
fi

if ! test -f "$INSTALL_DIR/data/nwc.db"; then
  echo "Could not find Alby Hub in this directory"
  exit 1
fi


echo "Running in $INSTALL_DIR"
# make sure we run this in the install directory
cd "$INSTALL_DIR" || exit 1

echo "Cleaning up old backup"
rm -rf albyhub-backup
mkdir albyhub-backup

echo "Creating current backup"
if command -v rsync > /dev/null 2>&1; then
  rsync -a data bin lib albyhub-backup/
else
  mv bin albyhub-backup
  mv lib albyhub-backup
  cp -r data albyhub-backup
fi


echo "Downloading latest version"
wget -q "$ALBYHUB_URL"

if [ "$SKIP_VERIFY" = false ]; then
  if ! ./verify.sh server-linux-x86_64.tar.bz2 albyhub-Server-Linux-x86_64.tar.bz2; then
    echo "❌ Verification failed, aborting installation"
    exit 1
  fi
fi

tar -xf server-linux-x86_64.tar.bz2
rm server-linux-x86_64.tar.bz2

if sudo systemctl list-units --type=service --all | grep -Fq albyhub.service; then
  echo "Starting Alby Hub"
  sudo systemctl start albyhub
fi

echo ""
echo ""
echo "✅ Update finished! Please unlock your wallet."
echo ""
