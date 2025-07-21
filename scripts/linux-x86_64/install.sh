#!/bin/bash

ALBYHUB_URL="https://getalby.com/install/hub/server-linux-x86_64.tar.bz2"
VERIFIER_URL="https://getalby.com/install/hub/verify.sh"
echo ""
echo ""
echo "⚡️ Welcome to Alby Hub"
echo "-----------------------------------------"
echo "Installing Alby Hub"
echo ""
read -p "Absolute install directory path (default: $HOME/albyhub): " USER_INSTALL_DIR

INSTALL_DIR="${USER_INSTALL_DIR:-$HOME/albyhub}"

# create installation directory
mkdir -p $INSTALL_DIR
cd $INSTALL_DIR

# download and extract the Alby Hub executable
wget $ALBYHUB_URL

if [[ ! -f "verify.sh" ]]; then
  echo "Downloading the verification script..."
  if ! wget -q "$VERIFIER_URL"; then
    echo "❌ Failed to download the verification script." >&2
    exit 1
  fi
  chmod +x verify.sh
fi

./verify.sh server-linux-x86_64.tar.bz2 albyhub-Server-Linux-x86_64.tar.bz2
if [[ $? -ne 0 ]]; then
  echo "❌ Verification failed, aborting installation"
  exit 1
fi

tar xvf server-linux-x86_64.tar.bz2
if [[ $? -ne 0 ]]; then
  echo "Failed to unpack Alby Hub. Potentially bzip2 is missing"
  echo "Install it with sudo apt-get install bzip2"
  exit
fi

rm server-linux-x86_64.tar.bz2

# prepare the data directory. this is pesistent and will hold all important data
mkdir -p $INSTALL_DIR/data

# create a simple start script that sets the default configuration variables
tee $INSTALL_DIR/start.sh > /dev/null << EOF
#!/bin/bash

echo "Starting Alby Hub"
WORK_DIR="$INSTALL_DIR/data" LDK_GOSSIP_SOURCE="" $INSTALL_DIR/bin/albyhub
EOF
chmod +x $INSTALL_DIR/start.sh

# add an update script to keep the Hub up to date
# run this to update the hub
wget https://raw.githubusercontent.com/getAlby/hub/master/scripts/linux-x86_64/update.sh
chmod +x $INSTALL_DIR/update.sh

echo ""
echo ""
echo "✅ Installation done."
echo ""

# optionally create a systemd service to start alby hub
read -p "Do you want to setup a systemd service (requires sudo permission)? (y/n): " -n 1 -r
if [[ ! $REPLY =~ ^[Yy]$ ]]
then
  echo ""
  echo ""
  echo "Run $INSTALL_DIR/start.sh to start Alby Hub"
  echo "✅ DONE"
  exit
fi

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
ExecStart=$INSTALL_DIR/start.sh
Environment="PORT=8029"

[Install]
WantedBy=multi-user.target
EOF

echo ""
echo ""

sudo systemctl enable albyhub
sudo systemctl start albyhub

echo "Run 'sudo systemctl start/stop albyhub' to start/stop AlbyHub"
echo ""
echo ""
echo " ✅ DONE. Open Alby Hub to get started"
echo "Alby Hub runs by default on localhost:8029"
