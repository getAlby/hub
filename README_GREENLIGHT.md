# Greenlight

To enable the GREENLIGHT LNClient some additional steps are required.

## Required Software

- Python 3 + pip
- [Greenlight Client](https://github.com/Blockstream/greenlight/tree/main?tab=readme-ov-file#install-and-updating-glcli-and-python-api) with [fix commit](https://github.com/Blockstream/greenlight/commit/2dc5a94668d41baef7275dae860c09b4a5dba198)

## Setup

1. Wallet can be registered or recovered through NWC UI
2. Get Liquidity
   1. peer with blocktank (TBC): `glcli connect 0296b2db342fcf87ea94d981757fdf4d3e545bd5cef4919f58b5d38dfdd73bf5c9 130.211.95.29 9735`
   2. run `glcli scheduler schedule` to get your node ID and GRPC uri
   3. go to blocktank and pay for a channel. After paying invoice, claim manually. Format is without the https: `pubkey@domain_name:9735`
   4. Unless you paid for some outgoing liquidity, you'll need to deposit and reserve at least 1% of the channel balance to send outgoing payments.
