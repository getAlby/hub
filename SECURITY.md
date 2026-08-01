# Security

## Key facts

- API keys are native to the daemon that minted them. They are not a security boundary: the spend gate on the completions path is the Cashu wallet balance, not the key.
- The daemon binds `*:8008` with unauthenticated admin endpoints and no key check on completions.

## Required hardening

**Block port 8008 in the firewall before the wallet holds sats.** Anyone who can reach it can manage clients and spend a funded wallet. See [docs/security.md](docs/security.md) for details.

## Reporting

Open an issue, or contact the maintainers directly for sensitive findings.
