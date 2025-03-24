#!/bin/bash

ALBYHUB_URL="https://getalby.com/install/hub/server-linux-aarch64.tar.bz2"
MANIFEST_URL="https://getalby.com/install/hub/manifest.txt"
SIGNATURE_URL="https://getalby.com/install/hub/manifest.txt.asc"
echo ""
echo ""
echo "⚡️ Welcome to Alby Hub"
echo "-----------------------------------------"
echo "Installing Alby Hub"
echo ""
read -p "Absolute install directory path (default: $HOME/albyhub): " USER_INSTALL_DIR

INSTALL_DIR="${USER_INSTALL_DIR:-$HOME/albyhub}"

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

# create installation directory
mkdir -p $INSTALL_DIR
cd $INSTALL_DIR

# download and extract the Alby Hub executable
wget $ALBYHUB_URL

if ! verify_package "server-linux-aarch64.tar.bz2" "albyhub-Server-Linux-aarch64.tar.bz2"; then
  echo "❌ Verification failed, aborting installation"
  exit 1
fi

tar xvf server-linux-aarch64.tar.bz2
if [[ $? -ne 0 ]]; then
  echo "Failed to unpack Alby Hub. Potentially bzip2 is missing"
  echo "Install it with sudo apt-get install bzip2"
  exit
fi

rm server-linux-aarch64.tar.bz2

# prepare the data directory. this is pesistent and will hold all important data
mkdir -p $INSTALL_DIR/data

# create a simple start script that sets the default configuration variables
tee $INSTALL_DIR/start.sh > /dev/null << EOF
#!/bin/bash

echo "Starting Alby Hub"
WORK_DIR="$INSTALL_DIR/data" LOG_EVENTS=true LDK_GOSSIP_SOURCE="" $INSTALL_DIR/bin/albyhub
EOF
chmod +x $INSTALL_DIR/start.sh

# add an update script to keep the Hub up to date
# run this to update the hub
wget https://raw.githubusercontent.com/getAlby/hub/master/scripts/linux-aarch64/update.sh
chmod +x $INSTALL_DIR/update.sh

echo ""
echo ""
echo "✅ Installation done."
echo ""

# optionally create a systemd service to start alby hub
read -p "Do you want to setup a systemd service (requires sudo permission)? (y/n): " -n 1 -r
if [[ ! $REPLY =~ ^[Yy]$ ]]
then
  echo ""
  echo ""
  echo "Run $INSTALL_DIR/start.sh to start Alby Hub"
  echo "✅ DONE"
  exit
fi

sudo tee /etc/systemd/system/albyhub.service > /dev/null << EOF
[Unit]
Description=Alby Hub
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
Restart=always
RestartSec=1
User=$USER
ExecStart=$INSTALL_DIR/start.sh
Environment="PORT=8029"

[Install]
WantedBy=multi-user.target
EOF

echo ""
echo ""

sudo systemctl enable albyhub
sudo systemctl start albyhub

echo "Run 'sudo systemctl start/stop albyhub' to start/stop AlbyHub"
echo ""
echo ""
echo " ✅ DONE. Open Alby Hub to get started"
echo "Alby Hub runs by default on localhost:8029"
