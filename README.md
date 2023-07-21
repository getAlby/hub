# Nostr Wallet Connect

This application allows you to control your Lightning node or wallet over Nostr.
Connect applications like [Damus](https://damus.io/) or [Amethyst](https://linktr.ee/amethyst.social) to your node.



**Specification**: [NIP-47](https://github.com/nostr-protocol/nips/blob/master/47.md)

## Supported Backends

* [Alby](https://getalby.com) (see: alby.go)
* LND (see: lnd.go)
* want more? please open an issue.

## Installation

### Requirements

The application has no runtime dependencies. (simple Go executable).

As data storage SQLite or PostgreSQL (recommended) can be used.

    $ cp .env.example .env
    # edit the config for your needs
    vim .env

  To get a new random Nostr key use `openssl rand -hex 32` or similar.

## Development

`go run .`

To build the CSS run:

`npx tailwindcss -i ./views/application.css -o ./public/css/application.css --watch`

## Configuration parameters

- `NOSTR_PRIVKEY`: the private key of this service. Should be a securely randomly generated 32 byte hex string.
- `CLIENT_NOSTR_PUBKEY`: if set, this service will only listen to events authored by this public key. You can set this to your own nostr public key.
- `RELAY`: default: "wss://relay.getalby.com/v1"
- `LN_BACKEND_TYPE`: ALBY or LND
- `ALBY_CLIENT_SECRET`= Alby OAuth client secret (used with the Alby backend)
- `ALBY_CLIENT_ID`= Alby OAuth client ID (used with the Alby backend)
- `OAUTH_REDIRECT_URL`= OAuth redirect URL (e.g. http://localhost:8080/alby/callback) (used with the Alby backend)
- `LND_ADDRESS`: the LND gRPC address, eg. `localhost:10009` (used with the LND backend)
- `LND_CERT_FILE`: the location where LND's `tls.cert` file can be found (used with the LND backend)
- `LND_MACAROON_FILE`: the location where LND's `admin.macaroon` file can be found (used with the LND backend)
- `COOKIE_SECRET`: a randomly generated secret string.
- `DATABASE_URI`: a postgres connection string or sqlite filename. Default: nostr-wallet-connect.db (sqlite)
- `PORT`: the port on which the app should listen on (default: 8080)

## Application deeplink options

### `/apps/new` deeplink options

Clients can use a deeplink to allow the user to add a new connection. Depending on the client this URL has different query options:

#### NWC created secret
The default option is that the NWC app creates a secret and the user uses the nostr wallet connect URL string to enable the client application.

##### Query parameter options

- `c`: the name of the client app

Example:

`/apps/new?c=myapp`

#### Client created secret
If the client creates the secret the client only needs to share the public key of that secret for authorization. The user authorized that pubkey and no sensitivate data needs to be shared.

##### Query parameter options
- `c`: the name of the client app
- `pubkey`: the public key of the client's secret for the user to authorize
- `return_to`: (optional) if a `return_to` URL is provided the user will be redirected to that URL after authorization. The `lud16`, `relay` and `pubkey` query parameters will be added to the URL.
- `expires_at` (optional) connection cannot be used after this date. Unix timestamp in seconds.
- `max_amount` (optional) maximum amount in sats that can be sent per renewal period
- `budget_renewal` (optional) reset the budget at the end of the given budget renewal. Can be `never` (default), `daily`, `weekly`, `monthly`, `yearly`
- `editable` (optional) set to `false` to disable form editing by the user

Example:

`/apps/new?c=myapp&pubkey=47c5a21...&return_to=https://example.com`

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
await nwc.initNWC({name: 'myapp'});
````

## Help

If you need help contact hello@getalby.com or reach out on Nostr: npub1getal6ykt05fsz5nqu4uld09nfj3y3qxmv8crys4aeut53unfvlqr80nfm


## ⚡️Donations

Want to support the work on Alby?

Support the Alby team ⚡️hello@getalby.com
You can also contribute to our [bounty program](https://github.com/getAlby/lightning-browser-extension/wiki/Bounties): ⚡️bounties@getalby.com
