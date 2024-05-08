
sudo mkdir -p /opt/albyhub
sudo chown -R $USER:$USER /opt/albyhub
cd /opt/albyhub
wget https://nightly.link/getalby/nostr-wallet-connect-next/workflows/package-raspberry-pi/master/nostr-wallet-connect.zip

unzip nostr-wallet-connect.zip -d app
rm nostr-wallet-connect.zip

### Create systemd service
cat > /etc/systemd/system/albyhub.service << EOF
[Unit]
Description=Alby Hub
After=network.target

[Service]
Type=simple
Restart=always
RestartSec=1
User=root
ExecStart=/opt/albyhub/app/nostr-wallet-connect

Environment="PORT=80"
Environment="WORK_DIR=/opt/albyhub/data"
Environment="LDK_ESPLORA_SERVER=https://electrs.albylabs.com"
Environment="LOG_EVENTS=true"

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl enable albyhub
sudo systemctl start albyhub
