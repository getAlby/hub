#!/bin/bash

echo "🔃 Updating Alby Hub..."
sudo systemctl stop albyhub

# Download new artifacts
cd /opt/albyhub
rm -rf albyhub-backup
mkdir albyhub-backup

if command -v rsync > /dev/null 2>&1; then
  rsync -av data bin lib albyhub-backup/
else
  mv bin albyhub-backup
  mv lib albyhub-backup
  cp -r data albyhub-backup
fi

wget https://getalby.com/install/hub/server-linux-aarch64.tar.bz2

./verify.sh server-linux-aarch64.tar.bz2 albyhub-Server-Linux-aarch64.tar.bz2
if [[ $? -ne 0 ]]; then
  echo "❌ Verification failed, aborting installation"
  exit 1
fi

# Extract archives
tar -xvf server-linux-aarch64.tar.bz2

# Cleanup
rm server-linux-aarch64.tar.bz2

sudo systemctl start albyhub

echo "✅ Update finished! Please login again to start your wallet."
