#!/bin/bash

ALBYHUB_URL="https://getalby.com/install/hub/server-linux-x86_64.tar.bz2"
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
tar xvf server-linux-x86_64.tar.bz2
if [[ $? -eq 0 ]]; then
  echo "Failed to unpack Alby Hub. Potentially bzip2 is missing"
  echo "Install it with sudo apt-get install bzip2"
fi

rm server-linux-x86_64.tar.bz2

# prepare the data directory. this is pesistent and will hold all important data
mkdir -p $INSTALL_DIR/data

# create a simple start script that sets the default configuration variables
tee -a $INSTALL_DIR/start.sh > /dev/null << EOF
#!/bin/bash

echo "Starting Alby Hub"
WORK_DIR="$INSTALL_DIR/data" LOG_EVENTS=true LDK_GOSSIP_SOURCE="" $INSTALL_DIR/bin/albyhub
EOF
chmod +x $INSTALL_DIR/start.sh

# add an update script to keep the Hub up to date
# run this to update the hub
tee -a $INSTALL_DIR/update.sh > /dev/null << EOF
#!/bin/bash

echo ""
echo ""
echo "⚡️ Updating Alby Hub"
echo "-----------------------------------------"
echo "This will download the latest version of Alby Hub."
echo "You will have to unlock Alby Hub after the update."
echo ""
echo "Make sure you have your unlock password available and a backup of your seed."

read -p "Do you want continue? (y/n): " -n 1 -r
if [[ ! \$REPLY =~ ^[Yy]$ ]]
then
  exit
fi

sudo systemctl list-units --type=service --all | grep -Fq albyhub.service
if [[ \$? -eq 0 ]]; then
  echo "Stopping Alby Hub"
  sudo systemctl stop albyhub
fi

if pgrep -x "albyhub" > /dev/null
then
  echo "Alby Hub process is still running, stopping it now."
  pkill -f albyhub
fi

# make sure we run this in the install directory
cd $INSTALL_DIR

echo "Cleaning up old backup"
rm -rf albyhub-backup
mkdir albyhub-backup

echo "Creating current backup"
mv bin albyhub-backup
mv lib albyhub-backup
cp -r data albyhub-backup


echo "Downloading latest version"
ALBYHUB_URL="$ALBYHUB_URL"
wget \$ALBYHUB_URL
tar -xvf server-linux-x86_64.tar.bz2
rm server-linux-x86_64.tar.bz2

sudo systemctl list-units --type=service --all | grep -Fq albyhub.service
if [[ \$? -eq 0 ]]; then
  echo "Starting Alby Hub"
  sudo systemctl start albyhub
fi

echo ""
echo ""
echo "✅ Update finished! Please unlock your wallet."
echo ""
EOF
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

sudo tee -a /etc/systemd/system/albyhub.service > /dev/null << EOF
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
echo "Alby Hub runs by default on localhost:8080"
