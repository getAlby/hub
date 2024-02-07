# Nostr Wallet Connect (Next)

This is a self-sovereign, self-custodial, single-user rewrite of NWC currently in an experimental phase development. **❗This version is not backwards compatible with NWC - it requires a fresh database and connections to be re-added**

This application allows you to control your Lightning node or wallet over Nostr.
Connect applications like [Damus](https://damus.io/) or [Amethyst](https://linktr.ee/amethyst.social) to your node.

**Specification**: [NIP-47](https://github.com/nostr-protocol/nips/blob/master/47.md)

The application can run in two modes:

- Wails (Desktop app): Mac (arm64), Windows (amd64), Linux (amd64)
- HTTP (Web app): Docker, Linux (amd64)

Ideally the app runs 24/7 (on a node, VPS or always-online desktop/laptop machine) so it can be connected to a lightning address and receive online payments.

## Supported Backends

- LND (see: lnd.go)
- Breez (see: breez.go)
- want more? please open an issue.

## Installation

### Requirements

The application has no runtime dependencies. (simple Go executable).

As data storage SQLite is used.

    $ cp .env.example .env
    # edit the config for your needs
    vim .env

To get a new random Nostr key use `openssl rand -hex 32` or similar.

## Development

### Server (HTTP mode)

1. Create a Lightning Polar setup with two LND nodes and uncomment the Polar LND section in your `.env` file.

2. Compile the frontend or run `touch frontend/dist/tmp` to ensure there are embeddable files available.

3. `go run .`

### React Frontend (HTTP mode)

Go to `/frontend`

1. `yarn install`
2. `yarn dev`

### Wails (Backend + Frontend)

_Make sure to have [wails](https://wails.io/docs/gettingstarted/installation) installed and all platform-specific dependencies installed (see wails doctor)_

`unset GTK_PATH && wails dev -tags "wails"`

_If you get a blank screen the first load, close the window and start the app again_

#### Wails Production build

`wails build -tags "wails"`

### Build and run locally (HTTP mode)

`mkdir tmp`
`go build -o main`
`cp main tmp`
`cp .env tmp`
`cd tmp`
`./main`

### Run dockerfile locally (HTTP mode)

`docker build . -t nwc-local`

`docker run --env-file .env -p 8080:8080 nwc-local`

### Testing

`go test`

### Windows

Breez SDK requires gcc to build the Breez bindings. Run `choco install mingw` and copy the breez SDK bindings file into the root of this directory (from your go installation directory) as per the [Breez SDK instructions](https://github.com/breez/breez-sdk-go?tab=readme-ov-file#windows). ALSO copy the bindings file into the output directory alongside the .exe in order to run it.

## Configuration parameters

- `NOSTR_PRIVKEY`: the private key of this service. Should be a securely randomly generated 32 byte hex string.
- `CLIENT_NOSTR_PUBKEY`: if set, this service will only listen to events authored by this public key. You can set this to your own nostr public key.
- `RELAY`: default: "wss://relay.getalby.com/v1"
- `COOKIE_SECRET`: a randomly generated secret string. (only needed in http mode)
- `DATABASE_URI`: a sqlite filename. Default: .data/nwc.db
- `PORT`: the port on which the app should listen on (default: 8080)
- `LN_BACKEND_TYPE`: LND or BREEZ
- `WORK_DIR`: directory to store NWC data files. Default: .data
-

### LND Backend parameters

_For cert and macaroon, either hex or file options can be used._

- `LND_ADDRESS`: the LND gRPC address, eg. `localhost:10009` (used with the LND backend)
- `LND_CERT_FILE`: the location where LND's `tls.cert` file can be found (used with the LND backend)
- `LND_MACAROON_FILE`: the location where LND's `admin.macaroon` file can be found (used with the LND backend)

### BREEZ Backend parameters

- `BREEZ_MNEMONIC`: your bip39 mnemonic key phrase e.g. "define limit soccer guilt trim mechanic beyond outside best give south shine"
- `BREEZ_API_KEY`: contact breez for more info
- `GREENLIGHT_INVITE_CODE`: contact blockstream for more info

## Application deeplink options

### `/apps/new` deeplink options

Clients can use a deeplink to allow the user to add a new connection. Depending on the client this URL has different query options:

#### NWC created secret

The default option is that the NWC app creates a secret and the user uses the nostr wallet connect URL string to enable the client application.

##### Query parameter options

- `name`: the name of the client app

Example:

`/apps/new?name=myapp`

#### Client created secret

If the client creates the secret the client only needs to share the public key of that secret for authorization. The user authorized that pubkey and no sensitivate data needs to be shared.

##### Query parameter options for /new

- `name`: the name of the client app
- `pubkey`: the public key of the client's secret for the user to authorize
- `return_to`: (optional) if a `return_to` URL is provided the user will be redirected to that URL after authorization. The `lud16`, `relay` and `pubkey` query parameters will be added to the URL.
- `expires_at` (optional) connection cannot be used after this date. Unix timestamp in seconds.
- `max_amount` (optional) maximum amount in sats that can be sent per renewal period
- `budget_renewal` (optional) reset the budget at the end of the given budget renewal. Can be `never` (default), `daily`, `weekly`, `monthly`, `yearly`
- `request_methods` (optional) url encoded, space separated list of request types that you need permission for: `pay_invoice` (default), `get_balance` (see NIP47). For example: `..&request_methods=pay_invoice%20get_balance`

Example:

`/apps/new?name=myapp&pubkey=47c5a21...&return_to=https://example.com`

#### Web-flow: client created secret

Web clients can open a new prompt popup to load the authorization page.
Once the user has authorized the app connection a `nwc:success` message is sent to the opening page (using `postMessage`) to indicate that the connection is authorized. See the `initNWC()` function in the [alby-js-sdk](https://github.com/getAlby/alby-js-sdk#nostr-wallet-connect-documentation)

Example:

```js
import { webln } from "alby-js-sdk";
const nwc = new webln.NWC();
// initNWC opens a prompt with /apps/new?c=myapp&pubkey=xxxx
// the promise resolves once the user has authorized the connection (when the `nwc:success` message is received) and the popup is closed automatically
// the promise rejects if the user cancels by closing the prompt popup
await nwc.initNWC({ name: "myapp" });
```

## Help

If you need help contact support@getalby.com or reach out on Nostr: npub1getal6ykt05fsz5nqu4uld09nfj3y3qxmv8crys4aeut53unfvlqr80nfm
You can also visit the chat of our Community on [Telegram](https://t.me/getalby).

## ⚡️Donations

Want to support the work on Alby?

Support the Alby team ⚡️hello@getalby.com
You can also contribute to our [bounty program](https://github.com/getAlby/lightning-browser-extension/wiki/Bounties): ⚡️bounties@getalby.com

## NIP-47 Supported Methods

✅ NIP-47 info event

❌ `expiration` tag in requests

### LND

✅ `get_info`

✅ `get_balance`

✅ `pay_invoice`

- ⚠️ amount not supported (for amountless invoices)
- ⚠️ PAYMENT_FAILED error code not supported

✅ `pay_keysend`

- ⚠️ PAYMENT_FAILED error code not supported

✅ `make_invoice`

✅ `lookup_invoice`

- ⚠️ NOT_FOUND error code not supported

✅ `list_transactions`

- ⚠️ from and until in request not supported
- ⚠️ failed payments will not be returned

✅ `multi_pay_invoice`

- ⚠️ amount not supported (for amountless invoices)
- ⚠️ PAYMENT_FAILED error code not supported

✅ `multi_pay_keysend`

- ⚠️ PAYMENT_FAILED error code not supported

### Breez

(Supported methods coming soon)

## Node Distributions

Run NWC on your own node!

**NOTE: the below links are for the original version of NWC**

- [https://github.com/getAlby/umbrel-community-app-store](Umbrel)
- [https://github.com/horologger/nostr-wallet-connect-startos](Start9)

## Deploy it yourself

### Fly

- [install fly](https://fly.io/docs/hands-on/install-flyctl/)
- update `app = 'nwc'` on **line 6** to a unique name in fly.toml e.g. `app = 'nwc-john-doe-1234'`
- run `fly launch`
  - press 'y' to copy configuration to the new app and then hit enter
  - press 'n' to tweak the settings and then hit enter
  - wait for the deployment to succeed, it should give you a URL like `https://nwc-john-doe-1234.fly.dev`

### Custom Ubuntu VM

- install go (using snap)
- install build-essential
- install nvm (curl script)
- with nvm, choose node lts
- install yarn (via npm)
- then run yarn build
- go run .

### Docker

(TBC)
