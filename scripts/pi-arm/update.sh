#!/bin/sh

echo "🔃 Updating Alby Hub..."
sudo systemctl stop albyhub

# Download new artifacts
cd /opt/albyhub || exit 1
rm -rf albyhub-backup
mkdir albyhub-backup

if command -v rsync > /dev/null 2>&1; then
  rsync -av data bin lib albyhub-backup/
else
  mv bin albyhub-backup
  mv lib albyhub-backup
  cp -r data albyhub-backup
fi

wget https://getalby.com/install/hub/server-linux-armv6.tar.bz2

if ! ./verify.sh server-linux-armv6.tar.bz2 albyhub-Server-Linux-armv6.tar.bz2; then
  echo "❌ Verification failed, aborting installation"
  exit 1
fi

# Extract archives
tar -xvf server-linux-armv6.tar.bz2

# Cleanup
rm server-linux-armv6.tar.bz2

sudo systemctl start albyhub

echo "✅ Update finished! Please login again to start your wallet."
