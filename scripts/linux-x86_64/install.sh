#!/bin/sh

ALBYHUB_URL="https://getalby.com/install/hub/server-linux-x86_64.tar.bz2"
VERIFIER_URL="https://getalby.com/install/hub/verify.sh"

# Default values
INSTALL_DIR=""
SYSTEMD=""
NON_INTERACTIVE=false
SKIP_VERIFY=false

# Parse command-line arguments
while [ $# -gt 0 ]; do
  case "$1" in
    -d|--install-dir)
      if [ -z "$2" ] || [ "${2#-}" != "$2" ]; then
        echo "Error: --install-dir requires a non-empty directory path"
        exit 1
      fi
      if printf '%s' "$2" | grep -q '[[:space:]]'; then
        echo "Error: --install-dir must not contain whitespace"
        exit 1
      fi
      INSTALL_DIR="$2"
      shift 2
      ;;
    -s|--systemd)
      SYSTEMD="yes"
      shift
      ;;
    --no-systemd)
      SYSTEMD="no"
      shift
      ;;
    -y|--yes)
      NON_INTERACTIVE=true
      shift
      ;;
    --skip-verify)
      SKIP_VERIFY=true
      shift
      ;;
    -h|--help)
      echo "Usage: $0 [OPTIONS]"
      echo ""
      echo "Options:"
      echo "  -d, --install-dir DIR    Set installation directory (default: \$HOME/albyhub)"
      echo "  -s, --systemd            Setup systemd service (auto-yes)"
      echo "      --no-systemd         Skip systemd service setup (auto-no)"
      echo "  -y, --yes                Non-interactive mode (auto-confirm all prompts)"
      echo "      --skip-verify        Skip package signature verification"
      echo "  -h, --help               Show this help message"
      echo ""
      echo "Examples:"
      echo "  $0                                    # Interactive mode"
      echo "  $0 -y -d /opt/albyhub -s              # Non-interactive with systemd"
      echo "  $0 --yes --install-dir /app/albyhub   # Non-interactive, no systemd"
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      echo "Use -h or --help for usage information"
      exit 1
      ;;
  esac
done

echo ""
echo ""
echo "⚡️ Welcome to Alby Hub"
echo "-----------------------------------------"
echo "Installing Alby Hub"
echo ""

# Determine install directory
if [ -z "$INSTALL_DIR" ]; then
  if [ "$NON_INTERACTIVE" = true ]; then
    INSTALL_DIR="$HOME/albyhub"
  else
    printf "Absolute install directory path (default: %s/albyhub): " "$HOME"
    read USER_INSTALL_DIR
    INSTALL_DIR="${USER_INSTALL_DIR:-$HOME/albyhub}"
  fi
fi

echo "Installing to: $INSTALL_DIR"

# create installation directory
mkdir -p "$INSTALL_DIR"
cd "$INSTALL_DIR" || exit 1

# check bzip2 is available before downloading
if ! command -v bzip2 > /dev/null 2>&1; then
  echo "❌ bzip2 is required but not installed. Run: sudo apt-get install -y bzip2" >&2
  exit 1
fi

# download and extract the Alby Hub executable
echo "Downloading Alby Hub..."
if ! wget -q "$ALBYHUB_URL"; then
  echo "❌ Failed to download Alby Hub" >&2
  exit 1
fi

if [ "$SKIP_VERIFY" = false ]; then
  if [ ! -f "verify.sh" ]; then
    echo "Downloading the verification script..."
    if ! wget -q "$VERIFIER_URL"; then
      echo "❌ Failed to download the verification script." >&2
      exit 1
    fi
    chmod +x verify.sh
  fi

  if ! ./verify.sh server-linux-x86_64.tar.bz2 albyhub-Server-Linux-x86_64.tar.bz2; then
    echo "❌ Verification failed, aborting installation"
    exit 1
  fi
fi

if ! tar xf server-linux-x86_64.tar.bz2; then
  echo "Failed to unpack Alby Hub. Potentially bzip2 is missing"
  echo "Install it with sudo apt-get install bzip2"
  exit 1
fi

rm server-linux-x86_64.tar.bz2

# prepare the data directory. this is pesistent and will hold all important data
mkdir -p "$INSTALL_DIR/data"

# create a simple start script that sets the default configuration variables
tee "$INSTALL_DIR/start.sh" > /dev/null << EOF
#!/bin/sh

echo "Starting Alby Hub"
WORK_DIR="$INSTALL_DIR/data" LDK_GOSSIP_SOURCE="" $INSTALL_DIR/bin/albyhub
EOF
chmod +x "$INSTALL_DIR/start.sh"

# add an update script to keep the Hub up to date
# run this to update the hub
wget -q https://raw.githubusercontent.com/getAlby/hub/master/scripts/linux-x86_64/update.sh
chmod +x "$INSTALL_DIR/update.sh"

echo ""
echo ""
echo "✅ Installation done."
echo ""

# optionally create a systemd service to start alby hub
SETUP_SYSTEMD=""
if [ -n "$SYSTEMD" ]; then
  if [ "$SYSTEMD" = "yes" ]; then
    SETUP_SYSTEMD="y"
  else
    SETUP_SYSTEMD="n"
  fi
elif [ "$NON_INTERACTIVE" = true ]; then
  # Default to no systemd in non-interactive mode unless explicitly requested
  SETUP_SYSTEMD="n"
else
  printf "Do you want to setup a systemd service (requires sudo permission)? (y/n): "
  read REPLY
  case "$REPLY" in
    [Yy]*) SETUP_SYSTEMD="y" ;;
    *) SETUP_SYSTEMD="n" ;;
  esac
fi

if [ "$SETUP_SYSTEMD" = "n" ]; then
  echo ""
  echo ""
  echo "Run $INSTALL_DIR/start.sh to start Alby Hub"
  echo "✅ DONE"
  exit 0
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
