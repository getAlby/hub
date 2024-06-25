#!/bin/bash

ALBYHUB_URL="https://nightly.link/getAlby/nostr-wallet-connect-next/workflows/http/build-executables/albyhub-Linux-x86_64.tar.gz.zip"
PHOENIX_VERSION="0.1.5"
PHOENIX_URL="https://github.com/ACINQ/phoenixd/releases/download/v$PHOENIX_VERSION/phoenix-$PHOENIX_VERSION-linux-x64.zip"

echo ""
echo ""
echo "⚡️ Welcome to AlbyHub"
echo "-----------------------------------------"
echo "Installing AlbyHub with phoenixd"
echo ""
read -p "Absolute install directory path (default: $HOME/albyhub-phoenixd): " USER_INSTALL_DIR

INSTALL_DIR="${USER_INSTALL_DIR:-$HOME/albyhub-phoenixd}"

echo "Installing phoenixd $PHOENIX_VERSION into $INSTALL_DIR"

mkdir -p $INSTALL_DIR
mkdir -p "$INSTALL_DIR/phoenixd"

cd $INSTALL_DIR

wget $PHOENIX_URL
unzip -j phoenix-$PHOENIX_VERSION-linux-x64.zip -d phoenixd

mkdir -p "$INSTALL_DIR/albyhub"
wget $ALBYHUB_URL
unzip albyhub-Linux-x86_64.tar.gz.zip
tar xf albyhub-Linux-x86_64.tar.gz --directory=albyhub

rm albyhub-Linux-x86_64.tar.gz
rm albyhub-Linux-x86_64.tar.gz.zip
rm phoenix-$PHOENIX_VERSION-linux-x64.zip

### Create start scripts
tee -a $INSTALL_DIR/phoenixd/start.sh > /dev/null << EOF
#!/bin/bash

echo "Starting phoenixd"
echo "Make sure to backup your phoenixd data in $INSTALL_DIR/phoenixd/data"
PHOENIX_DATADIR="$INSTALL_DIR/phoenixd/data" $INSTALL_DIR/phoenixd/phoenixd --agree-to-terms-of-service --http-bind-ip=0.0.0.0
EOF

tee -a $INSTALL_DIR/albyhub/start.sh > /dev/null << EOF
#!/bin/bash

echo "Starting Alby Hub"
phoenix_config_file=$INSTALL_DIR/phoenixd/data/phoenix.conf
PHOENIXD_AUTHORIZATION=\$(awk -F'=' '/^http-password/{print \$2}' "\$phoenix_config_file")
WORK_DIR="$INSTALL_DIR/albyhub/data" LN_BACKEND_TYPE=PHOENIX PHOENIXD_ADDRESS="http://localhost:9740" PHOENIXD_AUTHORIZATION=\$PHOENIXD_AUTHORIZATION LOG_EVENTS=true LDK_GOSSIP_SOURCE="" $INSTALL_DIR/albyhub/bin/albyhub-x86_64
EOF

tee -a $INSTALL_DIR/start.sh > /dev/null << EOF
#!/bin/bash

$INSTALL_DIR/phoenixd/start.sh &
# wait a but until phoenixd is started
# especially on the first run to make sure the config is there
sleep 7
$INSTALL_DIR/albyhub/start.sh &
echo "Started..."
EOF

chmod +x $INSTALL_DIR/start.sh
chmod +x $INSTALL_DIR/phoenixd/start.sh
chmod +x $INSTALL_DIR/albyhub/start.sh

echo ""
echo ""
echo "Installation done."
echo ""

read -p "Do you want to setup a systemd service? " -n 1 -r
if [[ ! $REPLY =~ ^[Yy]$ ]]
then
  echo "Run $INSTALL_DIR/start.sh to start phoenixd and Alby Hub"
  echo "DONE"
  exit
fi

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
ExecStart=$INSTALL_DIR/albyhub/start.sh

[Install]
WantedBy=multi-user.target
EOF

sudo tee -a /etc/systemd/system/phoenixd.service > /dev/null << EOF
[Unit]
Description=Phoenixd
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
Restart=always
RestartSec=1
User=root
ExecStart=$INSTALL_DIR/phoenixd/start.sh

[Install]
WantedBy=multi-user.target
EOF

echo ""
echo ""

echo "Run 'sudo systemctl enable albyhub' to enable the Albyhub service"
echo "Run 'sudo systemctl enable phoenixd' to enable the phoenixd service"
echo "Run 'sudo systemctl start albyhub' to start Albyhub"
echo "Run 'sudo systemctl start phoenixd' to start phoenixd"
echo ""
echo "DONE."
