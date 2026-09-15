#!/bin/sh

ALBYHUB_URL="https://getalby.com/install/hub/server-linux-x86_64.tar.bz2"
VERIFIER_URL="https://getalby.com/install/hub/verify.sh"

# Default values
INSTALL_DIR=""
SYSTEMD=""
NON_INTERACTIVE=false
SKIP_VERIFY=false

# Optional Caddy reverse proxy with an extra password gate, see
# https://github.com/getAlby/hub/tree/master/scripts/caddy-gate
CADDY_GATE=""
GATE_DOMAIN=""
GATE_USER=""
# Read from the environment so the password never shows up in `ps` output
GATE_PASSWORD="${ALBYHUB_GATE_PASSWORD:-}"

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
    --caddy-gate)
      CADDY_GATE="yes"
      shift
      ;;
    --no-caddy-gate)
      CADDY_GATE="no"
      shift
      ;;
    --domain)
      if [ -z "$2" ] || [ "${2#-}" != "$2" ]; then
        echo "Error: --domain requires a non-empty domain name"
        exit 1
      fi
      GATE_DOMAIN="$2"
      CADDY_GATE="yes"
      shift 2
      ;;
    --gate-user)
      if [ -z "$2" ] || [ "${2#-}" != "$2" ]; then
        echo "Error: --gate-user requires a non-empty username"
        exit 1
      fi
      GATE_USER="$2"
      shift 2
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
      echo "      --caddy-gate         Set up a Caddy reverse proxy with a password gate"
      echo "      --no-caddy-gate      Skip the Caddy password gate (auto-no)"
      echo "      --domain DOMAIN      Domain for the Caddy password gate (implies --caddy-gate)"
      echo "      --gate-user USER     Username for the Caddy password gate (default: albyhub)"
      echo "  -h, --help               Show this help message"
      echo ""
      echo "Environment variables:"
      echo "  ALBYHUB_GATE_PASSWORD    Password for the Caddy password gate, so it is"
      echo "                           not passed on the command line"
      echo ""
      echo "Examples:"
      echo "  $0                                    # Interactive mode"
      echo "  $0 -y -d /opt/albyhub -s              # Non-interactive with systemd"
      echo "  $0 --yes --install-dir /app/albyhub   # Non-interactive, no systemd"
      echo "  $0 -s --domain hub.example.com        # With a Caddy password gate"
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
  echo "❌ bzip2 is required but not installed." >&2
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

# Alby Hub listens on 8080 by default, the systemd service overrides it to 8029
ALBYHUB_PORT=8080

if [ "$SETUP_SYSTEMD" = "y" ]; then
  ALBYHUB_PORT=8029

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
fi

### Optional: Caddy reverse proxy with an extra password gate
#
# Alby Hub authenticates its own API with an `Authorization: Bearer <JWT>`
# header, so a plain `basic_auth` in front of it breaks the web UI: both want
# the same header. Instead, Basic Auth is used for the first request only,
# after which Caddy issues a random gate cookie and judges later requests on
# that cookie. See scripts/caddy-gate/README.md for the details.

GATE_ENV_FILE="/etc/caddy/gate.env"
CADDYFILE="/etc/caddy/Caddyfile"
ROTATE_SCRIPT="/usr/local/bin/albyhub-rotate-gate-secret.sh"
RAW_BASE_URL="https://raw.githubusercontent.com/getAlby/hub/master/scripts/caddy-gate"

# Read a secret without echoing it. Result is stored in SECRET_REPLY.
read_secret() {
  SECRET_REPLY=""
  printf '%s' "$1"
  if [ -t 0 ]; then
    STTY_ORIG=$(stty -g 2> /dev/null) || STTY_ORIG=""
    if [ -n "$STTY_ORIG" ]; then
      stty -echo 2> /dev/null
      read -r SECRET_REPLY
      stty "$STTY_ORIG" 2> /dev/null
      printf '\n'
      return 0
    fi
  fi
  read -r SECRET_REPLY
}

generate_gate_secret() {
  if command -v openssl > /dev/null 2>&1; then
    openssl rand -base64 48 | tr -d '=+/\n' | cut -c1-43
  else
    LC_ALL=C tr -dc 'A-Za-z0-9' < /dev/urandom | head -c 43
  fi
}

install_caddy() {
  if command -v caddy > /dev/null 2>&1; then
    return 0
  fi

  if ! command -v apt-get > /dev/null 2>&1; then
    echo "❌ Caddy is not installed, and this script can only install it automatically on Debian/Ubuntu."
    echo "   Install Caddy manually (https://caddyserver.com/docs/install), then follow"
    echo "   https://github.com/getAlby/hub/tree/master/scripts/caddy-gate"
    return 1
  fi

  if [ "$NON_INTERACTIVE" != true ]; then
    printf "Caddy is not installed. Install it from the official Caddy apt repository? (y/n): "
    read -r REPLY
    case "$REPLY" in
      [Yy]*) ;;
      *) echo "Skipping the password gate setup."; return 1 ;;
    esac
  fi

  echo "Installing Caddy..."
  sudo apt-get update || return 1
  sudo apt-get install -y debian-keyring debian-archive-keyring apt-transport-https curl gnupg || return 1
  curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' \
    | sudo gpg --batch --yes --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg || return 1
  curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' \
    | sudo tee /etc/apt/sources.list.d/caddy-stable.list > /dev/null || return 1
  sudo apt-get update || return 1
  sudo apt-get install -y caddy || return 1
}

setup_caddy_gate() {
  install_caddy || return 1

  # `basic_auth` was called `basicauth` before Caddy v2.8
  CADDY_VERSION=$(caddy version 2> /dev/null | head -n 1 | cut -d' ' -f1 | sed 's/^v//')
  CADDY_MAJOR=$(printf '%s' "$CADDY_VERSION" | cut -d. -f1)
  CADDY_MINOR=$(printf '%s' "$CADDY_VERSION" | cut -d. -f2)
  case "$CADDY_MAJOR.$CADDY_MINOR" in
    [0-9]*.[0-9]*)
      if [ "$CADDY_MAJOR" -lt 2 ] || { [ "$CADDY_MAJOR" -eq 2 ] && [ "$CADDY_MINOR" -lt 8 ]; }; then
        echo "❌ Caddy v2.8 or newer is required (found v$CADDY_VERSION)."
        return 1
      fi
      ;;
  esac

  # --- domain ---
  if [ -z "$GATE_DOMAIN" ] && [ "$NON_INTERACTIVE" != true ]; then
    printf "Domain name pointing to this server (e.g. hub.example.com): "
    read -r GATE_DOMAIN
  fi
  if [ -z "$GATE_DOMAIN" ]; then
    echo "❌ A domain name is required: the gate cookie is only accepted over HTTPS."
    return 1
  fi
  case "$GATE_DOMAIN" in
    *[!a-zA-Z0-9.-]* | -* | .* | *.)
      echo "❌ '$GATE_DOMAIN' does not look like a valid domain name."
      return 1
      ;;
    *.*) ;;
    *)
      echo "❌ '$GATE_DOMAIN' does not look like a valid domain name."
      return 1
      ;;
  esac

  # --- username ---
  if [ -z "$GATE_USER" ] && [ "$NON_INTERACTIVE" != true ]; then
    printf "Username for the password prompt (default: albyhub): "
    read -r GATE_USER
  fi
  GATE_USER="${GATE_USER:-albyhub}"
  case "$GATE_USER" in
    *[!a-zA-Z0-9._@-]*)
      echo "❌ The username may only contain letters, digits and . _ @ -"
      return 1
      ;;
  esac

  # --- password ---
  if [ -z "$GATE_PASSWORD" ]; then
    if [ "$NON_INTERACTIVE" = true ]; then
      echo "❌ Set ALBYHUB_GATE_PASSWORD when using --caddy-gate in non-interactive mode."
      return 1
    fi
    while [ -z "$GATE_PASSWORD" ]; do
      read_secret "Password for the password prompt (input hidden): "
      GATE_PASSWORD="$SECRET_REPLY"
      if [ -z "$GATE_PASSWORD" ]; then
        echo "The password must not be empty."
        continue
      fi
      read_secret "Repeat the password: "
      if [ "$GATE_PASSWORD" != "$SECRET_REPLY" ]; then
        echo "The passwords do not match, please try again."
        GATE_PASSWORD=""
      fi
    done
    SECRET_REPLY=""
  fi

  # Caddy reads the password from stdin so that it never appears in `ps`.
  # The trailing newline is required: without it caddy errors out on EOF.
  GATE_PASSWORD_HASH=$(printf '%s\n' "$GATE_PASSWORD" | caddy hash-password 2> /dev/null)
  GATE_PASSWORD=""
  if [ -z "$GATE_PASSWORD_HASH" ]; then
    echo "❌ Failed to hash the password with 'caddy hash-password'."
    return 1
  fi

  # --- gate secrets + systemd drop-in ---
  echo "Configuring the gate secrets..."
  sudo mkdir -p /etc/systemd/system/caddy.service.d
  sudo tee /etc/systemd/system/caddy.service.d/caddy-gate.conf > /dev/null << 'EOF'
# Installed by the Alby Hub install script.
# Makes the gate secrets available to Caddy as {env.GATE_SECRET} and
# {env.GATE_SECRET_OLD}. systemd reads the file as root, before Caddy drops
# privileges, so /etc/caddy/gate.env can stay root-only.
[Service]
EnvironmentFile=/etc/caddy/gate.env
EOF
  sudo systemctl daemon-reload

  sudo mkdir -p /etc/caddy
  if [ ! -f "$GATE_ENV_FILE" ]; then
    GATE_SECRET=$(generate_gate_secret)
    GATE_SECRET_OLD=$(generate_gate_secret)
    if [ -z "$GATE_SECRET" ] || [ -z "$GATE_SECRET_OLD" ]; then
      echo "❌ Failed to generate the gate secrets."
      return 1
    fi
    # create the file root-only *before* writing the secrets into it
    sudo install -m 600 -o root -g root /dev/null "$GATE_ENV_FILE"
    # GATE_SECRET_OLD gets a random value rather than an empty one so that it is
    # never guessable, even on a fresh install with nothing to keep alive yet
    printf '%s\n%s\n%s\n' \
      "# Managed by the Alby Hub install script and albyhub-rotate-gate-secret.sh." \
      "GATE_SECRET=$GATE_SECRET" \
      "GATE_SECRET_OLD=$GATE_SECRET_OLD" \
      | sudo tee "$GATE_ENV_FILE" > /dev/null
    GATE_SECRET=""
    GATE_SECRET_OLD=""
  else
    echo "Keeping the existing $GATE_ENV_FILE"
  fi

  # --- rotation script ---
  TMP_ROTATE=$(mktemp) || return 1
  if wget -q -O "$TMP_ROTATE" "$RAW_BASE_URL/rotate-gate-secret.sh"; then
    sudo install -m 755 -o root -g root "$TMP_ROTATE" "$ROTATE_SCRIPT"
  else
    echo "⚠️  Could not download the secret rotation script, continuing without it."
    ROTATE_SCRIPT=""
  fi
  rm -f "$TMP_ROTATE"

  # --- Caddyfile ---
  CADDYFILE_BACKUP=""
  if [ -f "$CADDYFILE" ]; then
    if [ "$NON_INTERACTIVE" != true ]; then
      printf "%s already exists. Replace it (a backup will be kept)? (y/n): " "$CADDYFILE"
      read -r REPLY
      case "$REPLY" in
        [Yy]*) ;;
        *) echo "Skipping the password gate setup."; return 1 ;;
      esac
    fi
    CADDYFILE_BACKUP="$CADDYFILE.albyhub-backup.$(date +%Y%m%d%H%M%S)"
    sudo cp "$CADDYFILE" "$CADDYFILE_BACKUP"
    echo "Backed up the previous Caddyfile to $CADDYFILE_BACKUP"
  fi

  TMP_CADDYFILE=$(mktemp) || return 1
  TMP_CADDYFILE_OUT=$(mktemp) || return 1
  # quoted heredoc: nothing in here is expanded by the shell, the placeholders
  # are substituted below (the bcrypt hash contains $ signs)
  cat > "$TMP_CADDYFILE" << 'CADDY_EOF'
# Alby Hub behind Caddy with an extra password gate.
# Generated by the Alby Hub install script.
# Documentation: https://github.com/getAlby/hub/tree/master/scripts/caddy-gate
#
# Alby Hub authenticates its own API with `Authorization: Bearer <JWT>`, which
# the frontend sets explicitly on every request. A plain `basic_auth` in front
# of it would therefore break the web UI, because both want the same header.
# Instead, Basic Auth is used for the first request only; after that Caddy
# issues a gate cookie and leaves the Authorization header to Alby Hub.

__DOMAIN__ {
	# A gate cookie holding the current *or* the previous secret is accepted, so
	# rotating the secret does not lock everyone out at once.
	#
	# The `!= ""` check is not cosmetic: if /etc/caddy/gate.env is missing or was
	# not loaded, both {env.*} placeholders expand to an empty string, and a
	# request carrying no cookie at all would compare "" == "" and get through.
	@gated expression `{http.request.cookie.__Host-albyhub_gate} != "" && ({http.request.cookie.__Host-albyhub_gate} == {env.GATE_SECRET} || {http.request.cookie.__Host-albyhub_gate} == {env.GATE_SECRET_OLD})`

	# Browsers keep attaching their cached `Authorization: Basic ...` header once
	# challenged. Drop it - but only when it really is a Basic credential, never
	# when it is Alby Hub's Bearer token.
	@basic_credentials header Authorization Basic*

	# --- Valid gate cookie: pass through, Authorization: Bearer untouched ---
	handle @gated {
		route {
			# Sliding expiry: every request extends the cookie, and a cookie still
			# holding the previous secret is upgraded to the current one.
			header +Set-Cookie "__Host-albyhub_gate={env.GATE_SECRET}; Path=/; HttpOnly; Secure; SameSite=Lax; Max-Age=604800"

			request_header @basic_credentials -Authorization
			# Alby Hub does not use cookies, its JWT lives in localStorage
			request_header -Cookie

			reverse_proxy 127.0.0.1:__ALBYHUB_PORT__
		}
	}

	# --- No or invalid gate cookie: Basic Auth challenge, then issue a cookie ---
	handle {
		route {
			basic_auth {
				# Change the password with: caddy hash-password
				__GATE_USER__ __GATE_HASH__
			}

			header +Set-Cookie "__Host-albyhub_gate={env.GATE_SECRET}; Path=/; HttpOnly; Secure; SameSite=Lax; Max-Age=604800"

			# Never forward the Basic credentials to Alby Hub
			request_header -Authorization
			request_header -Cookie

			reverse_proxy 127.0.0.1:__ALBYHUB_PORT__
		}
	}
}
CADDY_EOF

  sed -e "s|__DOMAIN__|$GATE_DOMAIN|g" \
      -e "s|__ALBYHUB_PORT__|$ALBYHUB_PORT|g" \
      -e "s|__GATE_USER__|$GATE_USER|g" \
      -e "s|__GATE_HASH__|$GATE_PASSWORD_HASH|g" \
      "$TMP_CADDYFILE" > "$TMP_CADDYFILE_OUT"
  sudo install -m 644 -o root -g root "$TMP_CADDYFILE_OUT" "$CADDYFILE"
  rm -f "$TMP_CADDYFILE" "$TMP_CADDYFILE_OUT"

  if ! sudo caddy validate --config "$CADDYFILE" --envfile "$GATE_ENV_FILE"; then
    echo "❌ The generated Caddy config failed validation."
    if [ -n "$CADDYFILE_BACKUP" ]; then
      sudo cp "$CADDYFILE_BACKUP" "$CADDYFILE"
      echo "Restored the previous Caddyfile from $CADDYFILE_BACKUP"
    fi
    return 1
  fi

  sudo systemctl enable caddy > /dev/null 2>&1
  # restart, not reload: systemd only reads EnvironmentFile when the service starts
  if ! sudo systemctl restart caddy; then
    echo "❌ Caddy failed to start. Check 'sudo journalctl -u caddy'."
    return 1
  fi

  CADDY_GATE_ENABLED="y"
  return 0
}

CADDY_GATE_ENABLED="n"

SETUP_CADDY_GATE=""
if [ -n "$CADDY_GATE" ]; then
  if [ "$CADDY_GATE" = "yes" ]; then
    SETUP_CADDY_GATE="y"
  else
    SETUP_CADDY_GATE="n"
  fi
elif [ "$NON_INTERACTIVE" = true ]; then
  SETUP_CADDY_GATE="n"
else
  echo ""
  echo ""
  echo "Alby Hub has its own unlock password. On top of that you can put a second"
  echo "password prompt on the web server itself, using Caddy as an HTTPS reverse"
  echo "proxy. Nobody reaches the Alby Hub login screen without it."
  echo "This requires a domain name pointing to this server."
  echo "Learn more: https://github.com/getAlby/hub/tree/master/scripts/caddy-gate"
  echo ""
  printf "Do you want to set this up (requires sudo permission)? (y/n): "
  read -r REPLY
  case "$REPLY" in
    [Yy]*) SETUP_CADDY_GATE="y" ;;
    *) SETUP_CADDY_GATE="n" ;;
  esac
fi

if [ "$SETUP_CADDY_GATE" = "y" ]; then
  if ! setup_caddy_gate; then
    echo ""
    echo "⚠️  The password gate was not set up. Alby Hub itself is installed and working."
    echo "   You can set it up later: https://github.com/getAlby/hub/tree/master/scripts/caddy-gate"
  fi
fi

echo ""
echo ""

if [ "$SETUP_SYSTEMD" = "y" ]; then
  echo "Run 'sudo systemctl start/stop albyhub' to start/stop AlbyHub"
else
  echo "Run $INSTALL_DIR/start.sh to start Alby Hub"
fi

echo ""
echo ""
echo " ✅ DONE. Open Alby Hub to get started"

if [ "$CADDY_GATE_ENABLED" = "y" ]; then
  echo "Alby Hub is available at https://$GATE_DOMAIN and asks for the '$GATE_USER' password first."
  if [ -n "$ROTATE_SCRIPT" ]; then
    echo "Rotate the gate secret any time with: sudo $ROTATE_SCRIPT"
  fi
  echo "⚠️  Block port $ALBYHUB_PORT in your firewall, otherwise Alby Hub can still be reached directly, bypassing the gate."
else
  echo "Alby Hub listens on port $ALBYHUB_PORT on all network interfaces, e.g. http://<server-ip>:$ALBYHUB_PORT"
  echo "⚠️  Alby Hub is intended for secure local networks. Make sure port $ALBYHUB_PORT is not reachable from the public internet (use a firewall)."
fi
