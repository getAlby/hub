#!/usr/bin/env bash
set -e

echo "=== Alby Hub Channel Funding from Coinbase ==="
echo ""
echo "Step 1: Get your on-chain Bitcoin address"
echo "  1. Open http://localhost:8087"
echo "  2. Navigate to Channels → Deposit"
echo "  3. Copy your on-chain Bitcoin address"
echo ""
read -p "Paste your Bitcoin address here: " BTC_ADDRESS

if [ -z "$BTC_ADDRESS" ]; then
    echo "Error: No address provided"
    exit 1
fi

echo ""
echo "Step 2: Withdraw from Coinbase"
echo "  1. Log into Coinbase"
echo "  2. Go to Assets → Bitcoin → Send"
echo "  3. Paste address: $BTC_ADDRESS"
echo "  4. Enter amount (minimum ~50,000 sats recommended)"
echo "  5. Complete withdrawal"
echo ""
echo "Step 3: Wait for confirmation"
echo "  - Check transaction: https://mempool.space/address/$BTC_ADDRESS"
echo "  - Wait for 1-3 confirmations (~10-30 minutes)"
echo "  - Your Alby Hub balance will update automatically"
echo ""
echo "Step 4: Open a Lightning channel"
echo "  1. Go to Channels → Open Channel"
echo "  2. Choose a node (or use Alby's recommended nodes)"
echo "  3. Set channel capacity"
echo "  4. Confirm and wait for channel to open"
echo ""
echo "Your address: $BTC_ADDRESS"
