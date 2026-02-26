#!/bin/sh

ALBYHUB_URL="https://getalby.com/install/hub/server-linux-x86_64.tar.bz2"
PHOENIX_VERSION="0.1.5"
PHOENIX_URL="https://github.com/ACINQ/phoenixd/releases/download/v$PHOENIX_VERSION/phoenix-$PHOENIX_VERSION-linux-x64.zip"

echo ""
echo ""
echo "⚡️ Welcome to AlbyHub"
echo "-----------------------------------------"
echo "Installing AlbyHub with phoenixd"
echo ""
printf "Absolute install directory path (default: %s/albyhub-phoenixd): " "$HOME"
read USER_INSTALL_DIR

INSTALL_DIR="${USER_INSTALL_DIR:-$HOME/albyhub-phoenixd}"

echo "Installing phoenixd $PHOENIX_VERSION into $INSTALL_DIR"

mkdir -p "$INSTALL_DIR"
mkdir -p "$INSTALL_DIR/phoenixd"

cd "$INSTALL_DIR" || exit 1

wget "$PHOENIX_URL"
unzip -j "phoenix-$PHOENIX_VERSION-linux-x64.zip" -d phoenixd

mkdir -p "$INSTALL_DIR/albyhub"
wget "$ALBYHUB_URL"
if ! tar xvf server-linux-x86_64.tar.bz2 --directory=albyhub; then
  echo "Failed to unpack Alby Hub. Potentially bzip2 is missing"
  echo "Install it with sudo apt-get install bzip2"
  exit 1
fi

rm server-linux-x86_64.tar.bz2
rm "phoenix-$PHOENIX_VERSION-linux-x64.zip"

### Create start scripts
tee "$INSTALL_DIR/phoenixd/start.sh" > /dev/null << EOF
#!/bin/sh

echo "Starting phoenixd"
echo "Make sure to backup your phoenixd data in $INSTALL_DIR/phoenixd/data"
PHOENIX_DATADIR="$INSTALL_DIR/phoenixd/data" $INSTALL_DIR/phoenixd/phoenixd --agree-to-terms-of-service --http-bind-ip=0.0.0.0
EOF

tee "$INSTALL_DIR/albyhub/start.sh" > /dev/null << EOF
#!/bin/sh

echo "Starting Alby Hub"
phoenix_config_file=$INSTALL_DIR/phoenixd/data/phoenix.conf
PHOENIXD_AUTHORIZATION=\$(awk -F'=' '/^http-password/{print \$2}' "\$phoenix_config_file")
WORK_DIR="$INSTALL_DIR/albyhub/data" LN_BACKEND_TYPE=PHOENIX PHOENIXD_ADDRESS="http://localhost:9740" PHOENIXD_AUTHORIZATION=\$PHOENIXD_AUTHORIZATION LDK_GOSSIP_SOURCE="" $INSTALL_DIR/albyhub/bin/albyhub
EOF

tee "$INSTALL_DIR/start.sh" > /dev/null << EOF
#!/bin/sh

$INSTALL_DIR/phoenixd/start.sh &
# wait a bit until phoenixd is started
# especially on the first run to make sure the config is there
sleep 8
$INSTALL_DIR/albyhub/start.sh &
echo "Started..."
EOF

chmod +x "$INSTALL_DIR/start.sh"
chmod +x "$INSTALL_DIR/phoenixd/start.sh"
chmod +x "$INSTALL_DIR/albyhub/start.sh"

echo ""
echo ""
echo "Installation done."
echo ""

printf "Do you want to setup a systemd service? (y/n): "
read REPLY
case "$REPLY" in
  [Yy]*) ;;
  *)
    echo "Run $INSTALL_DIR/start.sh to start phoenixd and Alby Hub"
    echo "DONE"
    exit
    ;;
esac

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
ExecStart=$INSTALL_DIR/albyhub/start.sh

[Install]
WantedBy=multi-user.target
EOF

sudo tee /etc/systemd/system/phoenixd.service > /dev/null << EOF
[Unit]
Description=Phoenixd
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
Restart=always
RestartSec=1
User=$USER
ExecStart=$INSTALL_DIR/phoenixd/start.sh

[Install]
WantedBy=multi-user.target
EOF

echo ""
echo ""

sudo systemctl enable albyhub
sudo systemctl enable phoenixd
sudo systemctl start phoenixd
sudo systemctl start albyhub

echo "Run 'sudo systemctl start/stop albyhub' to start/stop AlbyHub"
echo "Run 'sudo systemctl start/stop phoenixd' to start/stop phoenixd"
echo ""
echo "✅ DONE."
echo "Alby Hub runs by default on localhost:8080"
