echo ""
echo ""
echo "⚡️ Welcome to Alby Hub"
echo "-----------------------------------------"
echo "Installing..."


sudo mkdir -p /opt/albyhub
sudo chown -R $USER:$USER /opt/albyhub
cd /opt/albyhub
wget https://github.com/getAlby/nostr-wallet-connect-next/releases/latest/download/albyhub-Server-Linux-armv6.tar.bz2

# Extract archives
tar -xvf albyhub-Server-Linux-armv6.tar.bz2

# Cleanup
rm albyhub-Server-Linux-armv6.tar.bz2

### Create systemd service
sudo tee -a /etc/systemd/system/albyhub.service > /dev/null << EOF
[Unit]
Description=Alby Hub
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
Restart=always
RestartSec=1
User=root
ExecStart=/opt/albyhub/bin/albyhub

Environment="PORT=80"
Environment="WORK_DIR=/opt/albyhub/data"
Environment="LDK_ESPLORA_SERVER=https://electrs.albylabs.com"
Environment="LOG_EVENTS=true"
Environment="LDK_GOSSIP_SOURCE="

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl enable albyhub
sudo systemctl start albyhub

echo ""
echo ""
echo "✅ Installation finished! Please visit http://$HOSTNAME.local to configure your new Alby Hub."
echo ""
