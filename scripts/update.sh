#!/bin/bash

echo "🔃 Updating Alby Hub..."
sudo systemctl stop albyhub

# Download new artifacts
cd /opt/albyhub
rm -rf albyhub-backup
mkdir albyhub-backup
mv bin albyhub-backup
cp -r data albyhub-backup
wget https://github.com/getAlby/nostr-wallet-connect-next/releases/latest/download/albyhub-Server-Linux-armv6.tar.bz2

# Extract archives
tar -xvf albyhub-Server-Linux-armv6.tar.bz2

# Cleanup
rm albyhub-Server-Linux-armv6.tar.bz2

sudo systemctl start albyhub

echo "✅ Update finished! Please login again to start your wallet."