#!/bin/bash

ALBYHUB_URL="https://getalby.com/install/hub/server-linux-aarch64.tar.bz2"
echo ""
echo ""
echo "⚡️ Updating Alby Hub"
echo "-----------------------------------------"
echo "This will download the latest version of Alby Hub."
echo "You will have to unlock Alby Hub after the update."
echo ""
echo "Make sure you have your unlock password available and a backup of your seed."

read -p "Do you want continue? (y/n):" -n 1 -r
if [[ ! $REPLY =~ ^[Yy]$ ]]
then
  exit
fi
echo ""

sudo systemctl list-units --type=service --all | grep -Fq albyhub.service
if [[ $? -eq 0 ]]; then
  echo "Stopping Alby Hub"
  sudo systemctl stop albyhub
fi

if pgrep -x "albyhub" > /dev/null
then
  echo "Alby Hub process is still running, stopping it now."
  pkill -f albyhub
fi

SCRIPT_DIR=$(dirname "$(readlink -f "$0")")
read -p "Absolute install directory path (default: $SCRIPT_DIR): " USER_INSTALL_DIR
echo ""

INSTALL_DIR="${USER_INSTALL_DIR:-$SCRIPT_DIR}"

if ! test -f $INSTALL_DIR/data/nwc.db; then
  echo "Could not find Alby Hub in this directory"
  exit 1
fi


echo "Running in $INSTALL_DIR"
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
wget $ALBYHUB_URL

./verify.sh server-linux-aarch64.tar.bz2 albyhub-Server-Linux-aarch64.tar.bz2
if [[ $? -ne 0 ]]; then
  echo "❌ Verification failed, aborting installation"
  exit 1
fi

tar -xvf server-linux-aarch64.tar.bz2
rm server-linux-aarch64.tar.bz2

sudo systemctl list-units --type=service --all | grep -Fq albyhub.service
if [[ $? -eq 0 ]]; then
  echo "Starting Alby Hub"
  sudo systemctl start albyhub
fi

echo ""
echo ""
echo "✅ Update finished! Please unlock your wallet."
echo ""
