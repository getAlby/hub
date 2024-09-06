echo ""
echo ""
echo "⚡️ Welcome to Alby Hub"
echo "-----------------------------------------"
echo "Installing..."


sudo mkdir -p /opt/albyhub
sudo chown -R $USER:$USER /opt/albyhub
cd /opt/albyhub
wget https://getalby.com/install/hub/server-linux-aarch64.tar.bz2

# Extract archives
tar -xvf server-linux-aarch64.tar.bz2
if [[ $? -ne 0 ]]; then
  echo "Failed to unpack Alby Hub. Potentially bzip2 is missing"
  echo "Install it with sudo apt-get install bzip2"
fi

# Cleanup
rm server-linux-aarch64.tar.bz2

# allow albyhub to bind on port 80
sudo setcap CAP_NET_BIND_SERVICE=+eip /opt/albyhub/bin/albyhub

# make libs available
echo "/opt/albyhub/lib" | sudo tee /etc/ld.so.conf.d/albyhub.conf
sudo ldconfig

### Create systemd service
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
ExecStart=/opt/albyhub/bin/albyhub
# Hack to ensure Alby Hub never uses more than 90% CPU
CPUQuota=90%

Environment="PORT=$PORT"
Environment="WORK_DIR=/opt/albyhub/data"
Environment="LDK_ESPLORA_SERVER=https://electrs.getalbypro.com"
Environment="LOG_EVENTS=true"
Environment="LDK_GOSSIP_SOURCE="

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl enable albyhub
sudo systemctl start albyhub

echo ""
echo ""
echo "✅ Installation finished! Please visit $URL to configure your new Alby Hub."
echo ""
