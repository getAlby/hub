#!/bin/bash

echo "🔃 Updating Alby Hub..."
sudo systemctl stop albyhub

cd /opt/albyhub
sudo rm -rf albyhub-backup
mkdir albyhub-backup
mv app albyhub-backup
sudo cp -r data albyhub-backup
wget https://nightly.link/getalby/nostr-wallet-connect-next/workflows/package-raspberry-pi/master/nostr-wallet-connect.zip

unzip nostr-wallet-connect.zip -d app
chmod +x app/nostr-wallet-connect
rm nostr-wallet-connect.zip

sudo systemctl start albyhub

echo "✅ Update finished! Please visit http://$HOSTNAME.local to unlock your wallet."
