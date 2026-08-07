# Three-way comparison matrix: Greenlight standalone vs Hub standalone vs our GL backend

Status: skeleton. Populate as each flow is tested.

Scope: default flows on testnet (or local harness where testnet is blocked), then compare against our implementation.

| Flow | Upstream GL standalone | Upstream Hub standalone | Our GL backend | Verdict |
|---|---|---|---|---|
| Register node (dev cert) | | | | |
| Recovery from 12-word seed | | | | |
| Node getinfo / identity | | | | |
| Receive invoice (bolt11) | | | | |
| Send payment | | | | |
| Keysend | | | | |
| Balance / on-chain addr | | | | |
| Channel open/close | | | | |
| Signer lifecycle | | | | |
| Backup / restore | | | | |
| NWC app pairing | | | | |
| Isolated sub-wallets | | | | |
| Lightning address | | | | |
| UI setup wizard | | | | |

## Notes

- Upstream GL: glcli/glsdk-cli against Blockstream testnet with dev Nobody certs
- Upstream Hub: getAlby/hub default LDK flow on testnet
- Ours: feat/greenlight-backend, this repo
- Mark each cell with observed evidence (command, response, log line), not assumptions
