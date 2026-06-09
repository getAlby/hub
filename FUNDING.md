# Funding Alby Hub from Coinbase

## Quick Start

```bash
./fund-from-coinbase.sh
```

## Manual Steps

### 1. Get On-Chain Address

Open Alby Hub at http://localhost:8087 and navigate to:
- **Channels** → **Deposit**
- Copy your Bitcoin address

### 2. Withdraw from Coinbase

1. Log into [Coinbase](https://www.coinbase.com)
2. Go to **Assets** → **Bitcoin** → **Send**
3. Paste your Alby Hub address
4. Enter amount (minimum 50,000 sats / 0.0005 BTC recommended)
5. Complete 2FA and confirm withdrawal

### 3. Wait for Confirmation

- Track transaction: https://mempool.space
- Wait for 1-3 confirmations (~10-30 minutes)
- Your Alby Hub balance updates automatically

### 4. Open Lightning Channel

Once funds arrive:
1. Go to **Channels** → **Open Channel**
2. Choose a well-connected node or use Alby's recommendations
3. Set channel capacity (leave some for on-chain fees)
4. Confirm and wait for channel to open (~10 minutes)

## Recommended Nodes

- **Alby**: `031b301307574bbe9b9ac7b79cbe1700e31e544513eae0b5d7497483083f99e581@hub.getalby.com:9735`
- **ACINQ**: `03864ef025fde8fb587d989186ce6a4a186895ee44a926bfc370e2c366597a3f8f@3.33.236.230:9735`

## Fees

- Coinbase withdrawal: ~$1-5 (varies)
- Channel open: ~1000-5000 sats (depends on mempool)
- Keep 10-20% of funds for future on-chain operations
