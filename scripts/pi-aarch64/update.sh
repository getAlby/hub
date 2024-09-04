#!/bin/bash

echo "🔃 Updating Alby Hub..."
sudo systemctl stop albyhub

# Download new artifacts
cd /opt/albyhub
rm -rf albyhub-backup
mkdir albyhub-backup
mv bin albyhub-backup
mv lib albyhub-backup
cp -r data albyhub-backup

wget https://getalby.com/install/hub/server-linux-aarch64.tar.bz2

# Extract archives
tar -xvf server-linux-aarch64.tar.bz2

# Cleanup
rm server-linux-aarch64.tar.bz2

sudo systemctl start albyhub

echo "✅ Update finished! Please login again to start your wallet."
