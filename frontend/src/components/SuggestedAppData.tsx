import { ZapIcon } from "lucide-react";
import { Link } from "react-router-dom";
import albyGo from "src/assets/suggested-apps/alby-go.png";
import alby from "src/assets/suggested-apps/alby.png";
import amethyst from "src/assets/suggested-apps/amethyst.png";
import bitrefill from "src/assets/suggested-apps/bitrefill.png";
import btcpay from "src/assets/suggested-apps/btcpay.png";
import buzzpay from "src/assets/suggested-apps/buzzpay.png";
import clams from "src/assets/suggested-apps/clams.png";
import coracle from "src/assets/suggested-apps/coracle.png";
import damus from "src/assets/suggested-apps/damus.png";
import goose from "src/assets/suggested-apps/goose.png";
import hablanews from "src/assets/suggested-apps/habla-news.png";
import lightningMessageboard from "src/assets/suggested-apps/lightning-messageboard.png";
import lnbits from "src/assets/suggested-apps/lnbits.png";
import lume from "src/assets/suggested-apps/lume.png";
import nakapay from "src/assets/suggested-apps/nakapay.png";
import nostrcheckserver from "src/assets/suggested-apps/nostrcheck-server.png";
import nostrudel from "src/assets/suggested-apps/nostrudel.png";
import nostter from "src/assets/suggested-apps/nostter.png";
import nostur from "src/assets/suggested-apps/nostur.png";
import paperScissorsHodl from "src/assets/suggested-apps/paper-scissors-hodl.png";
import primal from "src/assets/suggested-apps/primal.png";
import pullthatupjamie from "src/assets/suggested-apps/pullthatupjamie.png";
import simpleboost from "src/assets/suggested-apps/simple-boost.png";
import snort from "src/assets/suggested-apps/snort.png";
import stackernews from "src/assets/suggested-apps/stacker-news.png";
import wavlake from "src/assets/suggested-apps/wavlake.png";
import wherostr from "src/assets/suggested-apps/wherostr.png";
import yakihonne from "src/assets/suggested-apps/yakihonne.png";
import zapstream from "src/assets/suggested-apps/zap-stream.png";
import zapplanner from "src/assets/suggested-apps/zapplanner.png";
import zapplepay from "src/assets/suggested-apps/zapple-pay.png";
import zappybird from "src/assets/suggested-apps/zappy-bird.png";
import zapstore from "src/assets/suggested-apps/zapstore.png";
import zeus from "src/assets/suggested-apps/zeus.png";
import ExternalLink from "src/components/ExternalLink";

export type SuggestedApp = {
  id: string;
  title: string;
  description: string;
  logo?: string;

  // General links
  webLink?: string;

  // App store links
  playLink?: string;
  appleLink?: string;
  zapStoreLink?: string;

  // Extension store links
  chromeLink?: string;
  firefoxLink?: string;

  extendedDescription?: string;
  guide?: React.ReactNode;
  internal?: boolean;
};

export const suggestedApps: SuggestedApp[] = [
  {
    id: "alby-go",
    title: "Alby Go",
    description: "A simple mobile wallet that works great with Alby Hub",
    webLink: "https://albygo.com",
    playLink:
      "https://play.google.com/store/apps/details?id=com.getalby.mobile",
    appleLink: "https://apps.apple.com/us/app/alby-go/id6471335774",
    zapStoreLink: "https://zapstore.dev/download/",
    logo: albyGo,
    extendedDescription: "Sends and receives payments seamlessly from your Hub",
    internal: true,
  },
  {
    id: "buzzpay",
    title: "BuzzPay PoS",
    description: "Receive-only PoS you can safely share with your employees",
    internal: true,
    logo: buzzpay,
  },
  {
    id: "goose",
    title: "Goose",
    description: "Your local AI agent, automating engineering tasks seamlessly",
    internal: true,
    logo: goose,
  },
  {
    id: "simpleboost",
    title: "SimpleBoost",
    description: "Donation widget for your website",
    internal: true,
    logo: simpleboost,
  },
  {
    id: "lightning-messageboard",
    title: "Lightning Messageboard",
    description: "Paid messageboard widget for your website",
    internal: true,
    logo: lightningMessageboard,
  },
  {
    id: "alby-extension",
    title: "Alby Extension",
    description: "Wallet in your browser",
    webLink: "https://getalby.com/",
    chromeLink:
      "https://chromewebstore.google.com/detail/iokeahhehimjnekafflcihljlcjccdbe",
    firefoxLink: "https://addons.mozilla.org/en-US/firefox/addon/alby/",
    logo: alby,
    extendedDescription:
      "Connect your Hub to lightning-enabled websites and lets you pay seamlessly on the web",
    guide: (
      <>
        <div>
          <h3 className="font-medium">In Alby Browser Extension</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              1. Download and open{" "}
              <span className="font-medium text-foreground">
                Alby Extension
              </span>{" "}
              in your desktop browser or for Firefox mobile
            </li>
            <li>2. Set your Unlock Passcode</li>
            <li>
              3. Connect your wallet via the{" "}
              <span className="font-medium text-foreground">Alby Account</span>{" "}
              (recommended) or via{" "}
              <span className="font-medium text-foreground">
                Nostr Wallet Connect
              </span>{" "}
              (Find Your Wallet → Nostr Wallet Connect)
            </li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">
            In Alby Hub{" "}
            <span className="text-muted-foreground font-normal">
              (only needed if you connect via NWC)
            </span>
          </h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              4. Click{" "}
              <Link
                to="/apps/new?app=alby-extension"
                className="font-medium text-foreground underline"
              >
                Connect to Alby Extension
              </Link>
            </li>
            <li>5. Set app's wallet permissions (full access recommended)</li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Alby Browser Extension</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              6. Paste the connection secret from Alby Hub →{" "}
              <span className="font-medium text-foreground">Continue</span>
            </li>
          </ul>
        </div>
      </>
    ),
  },
  {
    id: "damus",
    title: "Damus",
    description: "iOS Nostr client",
    webLink: "https://damus.io/?utm_source=getalby",
    appleLink: "https://apps.apple.com/ca/app/damus/id1628663131",
    logo: damus,
    extendedDescription:
      "Tip nostr posts and profiles and receive zaps seamlessly with your Hub",
    guide: (
      <>
        <div>
          <h3 className="font-medium">In Damus</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              1. Download and open{" "}
              <span className="font-medium text-foreground">Damus</span> on your
              iOS device
            </li>
            <li>
              2. Go to{" "}
              <span className="font-medium text-foreground">Wallet</span> →{" "}
              <span className="font-medium text-foreground">Attach Wallet</span>
            </li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Alby Hub</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              3. Click{" "}
              <Link
                to="/apps/new?app=damus"
                className="font-medium text-foreground underline"
              >
                Connect to Damus
              </Link>
            </li>
            <li>4. Set app's wallet permissions (full access recommended)</li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Damus</h3>
          <ul className="list-inside text-muted-foreground">
            <li>5. Scan or paste the connection secret from Alby Hub</li>
          </ul>
        </div>
      </>
    ),
  },
  {
    id: "amethyst",
    title: "Amethyst",
    description: "Android Nostr client",
    webLink:
      "https://play.google.com/store/apps/details?id=com.vitorpamplona.amethyst",
    playLink:
      "https://play.google.com/store/apps/details?id=com.vitorpamplona.amethyst",
    logo: amethyst,
    extendedDescription:
      "Tip nostr posts and profiles and receive zaps seamlessly with your Hub",
    guide: (
      <>
        <div>
          <h3 className="font-medium">In Amethyst</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              1. Download and open{" "}
              <span className="font-medium text-foreground">Amethyst</span> on
              your Android device
            </li>
            <li>
              2. Long-press
              <ZapIcon className="inline text-foreground mx-2" />
              below a post in your feed
            </li>
            <li>
              3. Click on the
              <span className="font-medium text-foreground">
                QR code icon
              </span>{" "}
              to activate the QR code scanner
            </li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Alby Hub</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              4. Click{" "}
              <Link
                to="/apps/new?app=amthyst"
                className="font-medium text-foreground underline"
              >
                Connect to Amethyst
              </Link>
            </li>
            <li>5. Set app's wallet permissions (full access recommended)</li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Amethyst</h3>
          <ul className="list-inside text-muted-foreground">
            <li>6. Scan or paste the connection secret from Alby Hub</li>
          </ul>
        </div>
      </>
    ),
  },
  {
    id: "primal",
    title: "Primal",
    description: "Cross-platform social",
    webLink: "https://primal.net/",
    appleLink: "https://apps.apple.com/us/app/primal/id1673134518",
    playLink:
      "https://play.google.com/store/apps/details?id=net.primal.android",
    logo: primal,
    extendedDescription:
      "Tip nostr posts and profiles and receive zaps seamlessly with your Hub",
    guide: (
      <>
        <div>
          <h3 className="font-medium">In Primal</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              1. Download and open{" "}
              <span className="font-medium text-foreground">Primal</span> on
              your Android or iOS device
            </li>
            <li>
              2. Click on your{" "}
              <span className="font-medium text-foreground">profile image</span>{" "}
              in the top left corner →{" "}
              <span className="font-medium text-foreground">Settings</span> →{" "}
              <span className="font-medium text-foreground">Wallet</span> →{" "}
              <span className="font-medium text-foreground">
                Untoggle Primal wallet
              </span>
            </li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Alby Hub</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              3. Click{" "}
              <Link
                to="/apps/new?app=primal"
                className="font-medium text-foreground underline"
              >
                Connect to Primal
              </Link>
            </li>
            <li>
              4. Set app's wallet permissions (full access recommended) and
              click on "Next"
            </li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Primal</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              5. Click on "Paste NWC String" or "Scan NWC QR Code" to connect
              Alby Hub
            </li>
          </ul>
        </div>
      </>
    ),
  },
  {
    id: "zap-stream",
    title: "Zap Stream",
    description: "Stream and stack sats",
    webLink: "https://zap.stream/",
    logo: zapstream,
    extendedDescription:
      "Tip streamers, zap comments and pay or receive sats for streaming time with your Hub",
    guide: (
      <>
        <div>
          <h3 className="font-medium">In Zap Stream</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              1. Open{" "}
              <ExternalLink
                to="https://zap.stream"
                className="font-medium text-foreground underline"
              >
                Zap Stream
              </ExternalLink>{" "}
              in your browser and log in
            </li>
            <li>
              2. Click on your{" "}
              <span className="font-medium text-foreground">Profile Image</span>{" "}
              → <span className="font-medium text-foreground">Settings</span> →{" "}
              scroll to{" "}
              <span className="font-medium text-foreground">Wallet</span>
            </li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Alby Hub</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              3. Click{" "}
              <Link
                to="/apps/new?app=zap-stream"
                className="font-medium text-foreground underline"
              >
                Connect to Zap Stream
              </Link>
            </li>
            <li>4. Set app's wallet permissions (full access recommended)</li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Zap Stream</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              5. Paste connection secret from Alby Hub and click on{" "}
              <span className="font-medium text-foreground">Connect</span>
            </li>
          </ul>
        </div>
      </>
    ),
  },
  {
    id: "btcpay",
    title: "BTCPay Server",
    description: "Bitcoin payment processor",
    webLink: "https://btcpayserver.org/",
    logo: btcpay,
    extendedDescription:
      "Receive payments directly to your Hub for products you sell online",
    guide: (
      <>
        <div>
          <p>
            You can use your Alby Hub as a lightning wallet funding source for
            your{" "}
            <ExternalLink
              to="https://btcpayserver.org/"
              className="font-medium text-foreground underline"
            >
              BTCPay Server
            </ExternalLink>{" "}
            store, to accept and create payments. In order for this feature to
            work, your BTCPay Server instance needs to have the{" "}
            <span className="font-medium text-foreground">Nostr</span> plugin
            installed.
          </p>
        </div>
        <div>
          <h3 className="font-medium">In BTCPay Server</h3>
          <ul className="list-inside text-muted-foreground">
            <li>1. Log in to your BTCPay Server dashboard</li>
            <li>
              2. Find connection configuration for your Lightning node (
              <span className="font-medium text-foreground">Lightning</span> →
              <span className="font-medium text-foreground">Settings</span> →
              <span className="font-medium text-foreground">
                Change connection
              </span>
              )
            </li>
            <li>
              3. Choose{" "}
              <span className="font-medium text-foreground">
                Use custom node
              </span>
            </li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Alby Hub</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              4. Click{" "}
              <Link
                to="/apps/new?app=btcpay"
                className="font-medium text-foreground underline"
              >
                Connect to BTCPay Server
              </Link>
            </li>
            <li>
              5. Set wallet permissions as read-only unless payments are
              specifically needed.
            </li>
            <li>6. Copy generated NWC connection secret</li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In BTCPay Server</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              7. Paste the connection secret (nostr+walletconnect://....) in the
              configuration field
            </li>
            <li>
              8. Click <span className="font-medium text-foreground">Save</span>
            </li>
          </ul>
        </div>
      </>
    ),
  },
  {
    id: "lnbits",
    title: "LNbits",
    description: "Wallet accounts system with extensions",
    webLink: "https://lnbits.com/",
    logo: lnbits,
    extendedDescription:
      "Connect your Alby Hub to LNbits to give extra functionality through plugins such as BOLT cards and lightning vouchers",
    guide: (
      <>
        <div>
          <p>
            You can use your Alby Hub as a lightning wallet funding source for
            your{" "}
            <ExternalLink
              to="https://lnbits.com/"
              className="font-medium text-foreground underline"
            >
              LNbits
            </ExternalLink>{" "}
            instance, to accept and create payments.
          </p>
        </div>
        <div>
          <h3 className="font-medium">In LNbits</h3>
          <ul className="list-inside text-muted-foreground">
            <li>1. Log in to your LNbits admin dashboard</li>
            <li>
              2. Go to{" "}
              <span className="font-medium text-foreground">Manage</span> →{" "}
              <span className="font-medium text-foreground">Server</span> →{" "}
              <span className="font-medium text-foreground">Funding</span>, to
              configure funding wallet
            </li>
            <li>
              3. Under{" "}
              <span className="font-medium text-foreground">
                Active Funding
              </span>{" "}
              choose{" "}
              <span className="font-medium text-foreground">
                Nostr Wallet Connect
              </span>
            </li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Alby Hub</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              4. Click{" "}
              <Link
                to="/apps/new?app=lnbits"
                className="font-medium text-foreground underline"
              >
                Connect to LNbits
              </Link>
            </li>
            <li>5. Set wallet permissions, according to your preferences</li>
            <li>6. Copy generated NWC connection secret</li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In LNbits</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              7. Paste the connection secret (nostr+walletconnect://....) under{" "}
              <span className="font-medium text-foreground">Pairing URL</span>
            </li>
            <li>
              8. Click <span className="font-medium text-foreground">Save</span>{" "}
              and{" "}
              <span className="font-medium text-foreground">
                Restart Server
              </span>
            </li>
          </ul>
        </div>
      </>
    ),
  },
  {
    id: "coracle",
    title: "Coracle.social",
    description: "Desktop Nostr client",
    webLink: "https://coracle.social/",
    logo: coracle,
    extendedDescription:
      "Tip nostr posts and profiles and receive zaps seamlessly with your Hub",
    guide: (
      <>
        <p>
          You can connect your Alby Hub to Coracle to zap Nostr notes directly
          from your node.
        </p>
        <div>
          <h3 className="font-medium">In Coracle</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              1. Log in with your Nostr keys to{" "}
              <ExternalLink
                to="https://coracle.social/login"
                className="font-medium text-foreground underline"
              >
                Coracle
              </ExternalLink>{" "}
              (it is recommended to use the Alby Extension)
            </li>
            <li>
              2. Click on a zap icon ⚡ and{" "}
              <span className="font-medium text-foreground">Zap!</span> under
              any post, to configure wallet connection and make your first zap
            </li>
            <li>
              3. Click{" "}
              <span className="font-medium text-foreground">
                Connect Wallet to Pay
              </span>{" "}
              and choose{" "}
              <span className="font-medium text-foreground">
                Nostr Wallet Connect
              </span>
            </li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Alby Hub</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              4. Click{" "}
              <Link
                to="/apps/new?app=coracle"
                className="font-medium text-foreground underline"
              >
                Connect to Coracle
              </Link>
            </li>
            <li>
              5. Set wallet permissions (required:{" "}
              <span className="font-medium text-foreground">Send payments</span>{" "}
              and{" "}
              <span className="font-medium text-foreground">
                Lookup status of invoices
              </span>
              ) and maximum spendable budget
            </li>
            <li>
              6. Click <span className="font-medium text-foreground">Next</span>{" "}
              and copy generated NWC connection secret
            </li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Coracle</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              7. Paste the connection secret (nostr+walletconnect://....) and
              click <span className="font-medium text-foreground">Connect</span>
            </li>
          </ul>
        </div>
      </>
    ),
  },
  {
    id: "nostter",
    title: "Nostter",
    description: "Minimalistic, desktop Nostr client",
    webLink: "https://nostter.app/",
    logo: nostter,
    extendedDescription:
      "Tip nostr posts and profiles and receive zaps seamlessly with your Hub",
    guide: (
      <>
        <p>
          You can connect your Alby Hub to Nostter to zap Nostr notes directly
          from your node.
        </p>
        <div>
          <h3 className="font-medium">In Nostter</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              1. Log in with your Nostr keys to{" "}
              <ExternalLink
                to="https://nostter.app/"
                className="font-medium text-foreground underline"
              >
                Nostter
              </ExternalLink>{" "}
              (it is recommended to use the Alby Extension)
            </li>
            <li>
              2. Go to{" "}
              <span className="font-medium text-foreground">Preferences</span>
            </li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Alby Hub</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              3. Click{" "}
              <Link
                to="/apps/new?app=nostter"
                className="font-medium text-foreground underline"
              >
                Connect to Nostter
              </Link>
            </li>
            <li>
              4. Set wallet permissions (required:{" "}
              <span className="font-medium text-foreground">Send payments</span>{" "}
              and{" "}
              <span className="font-medium text-foreground">
                Lookup status of invoices
              </span>
              ) and maximum spendable budget
            </li>
            <li>
              5. Click <span className="font-medium text-foreground">Next</span>{" "}
              and copy generated NWC connection secret
            </li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Nostter</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              6. Paste the connection secret (nostr+walletconnect://....) under{" "}
              <span className="font-medium text-foreground">
                Nostr Wallet Connect
              </span>
            </li>
            <li>
              7. Click elsewhere for the connection to be tested and saved
            </li>
            <li>
              8. Go to <span className="font-medium text-foreground">Home</span>{" "}
              and click the zap icon (⚡) under any post to add a comment and
              send zap directly from your node
            </li>
          </ul>
        </div>
      </>
    ),
  },
  {
    id: "wavlake",
    title: "Wavlake",
    description: "Creators platform",
    webLink: "https://www.wavlake.com/",
    playLink:
      "https://play.google.com/store/apps/details?id=com.wavlake.mobile",
    appleLink: "https://testflight.apple.com/join/eWnqECG4",
    logo: wavlake,
    extendedDescription:
      "Support artists by paying to upvote music you enjoy with your Hub",
    guide: (
      <>
        <div>
          <h3 className="font-medium">In Wavlake</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              1. Download and open{" "}
              <span className="font-medium text-foreground">Wavlake</span> on
              your iOS or Android device
            </li>
            <li>
              2. Click on <span className="font-medium text-foreground">≡</span>{" "}
              → <span className="font-medium text-foreground">Settings</span> →{" "}
              <span className="font-medium text-foreground">
                Add a NWC compatible wallet
              </span>
            </li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Alby Hub</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              3. Click{" "}
              <Link
                to="/apps/new?app=wavlake"
                className="font-medium text-foreground underline"
              >
                Connect to Wavlake
              </Link>
            </li>
            <li>4. Set app's wallet permissions (full access recommended)</li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Wavlake</h3>
          <ul className="list-inside text-muted-foreground">
            <li>5. Scan or paste the connection secret from Alby Hub</li>
          </ul>
        </div>
      </>
    ),
  },
  {
    id: "snort",
    title: "Snort",
    description: "Web Nostr client",
    webLink: "https://snort.social/",
    logo: snort,
    extendedDescription:
      "Tip nostr posts and profiles and receive zaps seamlessly with your Hub",
    guide: (
      <>
        <div>
          <h3 className="font-medium">In Snort</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              1. Open{" "}
              <ExternalLink
                to="https://snort.social"
                className="font-medium text-foreground underline"
              >
                Snort
              </ExternalLink>{" "}
              in your browser
            </li>
            <li>
              2. Click on{" "}
              <span className="font-medium text-foreground">Settings</span> →{" "}
              <span className="font-medium text-foreground">Wallet</span>
            </li>
            <li>
              3. Choose{" "}
              <span className="font-medium text-foreground">
                Nostr Wallet Connect
              </span>
            </li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Alby Hub</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              4. Click{" "}
              <Link
                to="/apps/new?app=snort"
                className="font-medium text-foreground underline"
              >
                Connect to Snort
              </Link>
            </li>
            <li>5. Set app's wallet permissions (full access recommended)</li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Snort</h3>
          <ul className="list-inside text-muted-foreground">
            <li>6. Paste the connection secret from Alby Hub</li>
          </ul>
        </div>
      </>
    ),
  },
  {
    id: "habla-news",
    title: "Habla News",
    description: "Blogging platform",
    webLink: "https://habla.news/",
    logo: hablanews,
    extendedDescription:
      "Tip nostr posts and profiles and receive zaps seamlessly with your Hub",
    guide: (
      <>
        <div>
          <h3 className="font-medium">In Habla News</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              1. Open{" "}
              <ExternalLink
                to="https://habla.news"
                className="font-medium text-foreground underline"
              >
                habla.news
              </ExternalLink>{" "}
              in your browser
            </li>
            <li>
              2. Go to{" "}
              <span className="font-medium text-foreground">Settings</span> →{" "}
              <span className="font-medium text-foreground">Wallet</span> →{" "}
              Click{" "}
              <span className="font-medium text-foreground">
                Connect Wallet
              </span>
            </li>
            <li>
              3. Choose{" "}
              <span className="font-medium text-foreground">NWC Generic</span>{" "}
            </li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Alby Hub</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              3. Click{" "}
              <Link
                to="/apps/new?app=habla-news"
                className="font-medium text-foreground underline"
              >
                Connect to Habla News
              </Link>
            </li>
            <li>4. Set app's wallet permissions (full access recommended)</li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Habla News</h3>
          <ul className="list-inside text-muted-foreground">
            <li>5. Paste the connection secret from Alby Hub</li>
          </ul>
        </div>
      </>
    ),
  },
  {
    id: "clams",
    title: "Clams",
    description: "Multi wallet accounting tool",
    webLink: "https://clams.tech/",
    logo: clams,
    extendedDescription:
      "Get insights into your transaction history and accounting tools by connecting your Hub",
    guide: (
      <>
        <div>
          <h3 className="font-medium">In Clams</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              1. Download and open{" "}
              <ExternalLink
                to="https://clams.tech/"
                className="font-medium text-foreground underline"
              >
                Clams
              </ExternalLink>{" "}
              on your device
            </li>
            <li>2. Add a connection: "+ Add Connection" → NWC</li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Alby Hub</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              {" "}
              3. Click{" "}
              <Link
                to="/apps/new?app=clams"
                className="font-medium text-foreground underline"
              >
                Connect to Clams
              </Link>
            </li>
            <li>4. Set wallet permissions (Read Only)</li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Clams</h3>
          <ul className="list-inside text-muted-foreground">
            <li>5. Add label & paste connection secret</li>
            <li>6. Click Connect and Save</li>
          </ul>
        </div>
      </>
    ),
  },
  {
    id: "nostrcheck-server",
    title: "Nostrcheck Server",
    description: "Sovereign Nostr services",
    webLink: "https://github.com/quentintaranpino/nostrcheck-server",
    logo: nostrcheckserver,
    extendedDescription:
      "Enable payments to your Hub from users who register or upload and download files",
    guide: (
      <>
        <div>
          <h3 className="font-medium">In Alby Hub</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              1. Click{" "}
              <Link
                to="/apps/new?app=nostrcheck-server"
                className="font-medium text-foreground underline"
              >
                Connect to Nostrcheck Server
              </Link>
            </li>
            <li>2. Set app's wallet permissions (full access recommended)</li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Nostrcheck server</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              3. Go to{" "}
              <span className="font-medium text-foreground">Settings</span> and
              choose{" "}
              <span className="font-medium text-foreground">Payments</span> tab
            </li>
            <li>
              4. Scroll to Nostr wallet connect settings and paste the{" "}
              <span className="font-medium text-foreground">
                connection secret
              </span>{" "}
              from Alby Hub
            </li>
            <li>
              5. Press the{" "}
              <span className="font-medium text-foreground">Save</span> button
            </li>
          </ul>
        </div>
      </>
    ),
  },
  {
    id: "nostrudel",
    title: "noStrudel",
    description: "Web Nostr client",
    webLink: "https://nostrudel.ninja/",
    logo: nostrudel,
    extendedDescription:
      "Tip nostr posts and profiles and receive zaps seamlessly with your Hub",
    guide: (
      <>
        <div>
          <h3 className="font-medium">In Nostrudel</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              1. Open{" "}
              <ExternalLink
                to="https://nostrudel.ninja"
                className="font-medium text-foreground underline"
              >
                NoStrudel
              </ExternalLink>{" "}
              in your browser
            </li>
            <li>
              2. Click on{" "}
              <span className="font-medium text-foreground">Settings</span> →{" "}
              <span className="font-medium text-foreground">Lightning</span> →{" "}
              <span className="font-medium text-foreground">
                Connect Wallet
              </span>
            </li>
            <li>
              3. Choose{" "}
              <span className="font-medium text-foreground">
                Custom Nostr Wallet Connect
              </span>
            </li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Alby Hub</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              4. Click{" "}
              <Link
                to="/apps/new?app=nostrudel"
                className="font-medium text-foreground underline"
              >
                Connect to noStrudel
              </Link>
            </li>
            <li>5. Set app's wallet permissions (full access recommended)</li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Nostrudel</h3>
          <ul className="list-inside text-muted-foreground">
            <li>6. Paste the connection secret from Alby Hub</li>
          </ul>
        </div>
      </>
    ),
  },
  {
    id: "yakihonne",
    title: "YakiHonne",
    description: "Your all in one nostr client",
    webLink: "https://yakihonne.com/",
    playLink:
      "https://play.google.com/store/apps/details?id=com.yakihonne.yakihonne",
    appleLink: "https://apps.apple.com/us/app/yakihonne/id6472556189",
    logo: yakihonne,
    extendedDescription:
      "Tip nostr posts and profiles and receive zaps seamlessly with your Hub",
    guide: (
      <>
        <div>
          <h3 className="font-medium">In YakiHonne</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              1. Open{" "}
              <ExternalLink
                to="https://yakihonne.com/wallet"
                className="font-medium text-foreground underline"
              >
                YakiHonne
              </ExternalLink>{" "}
              in your browser
            </li>
            <li>
              2. Click on{" "}
              <span className="font-medium text-foreground">Add wallet</span>
            </li>
            <li>
              3. Choose{" "}
              <span className="font-medium text-foreground">
                Nostr Wallet Connect
              </span>
            </li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Alby Hub</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              4. Click{" "}
              <Link
                to="/apps/new?app=yakihonne"
                className="font-medium text-foreground underline"
              >
                Connect to YakiHonne
              </Link>
            </li>
            <li>5. Set app's wallet permissions (full access recommended)</li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In YakiHonne</h3>
          <ul className="list-inside text-muted-foreground">
            <li>6. Paste the connection secret from Alby Hub</li>
          </ul>
        </div>
      </>
    ),
  },
  {
    id: "zapplanner",
    title: "ZapPlanner",
    description: "Schedule payments",
    webLink: "https://zapplanner.albylabs.com/",
    logo: zapplanner,
    internal: true,
  },
  {
    id: "zapplepay",
    title: "Zapple Pay",
    description: "Zap from any client",
    webLink: "https://www.zapplepay.com/",
    logo: zapplepay,
    extendedDescription:
      "ZapplePay will make payments from your Hub to zap posts when you react to them",
    guide: (
      <>
        <div>
          <h3 className="font-medium">In Zapple Pay</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              1. Open{" "}
              <ExternalLink
                to="https://www.zapplepay.com"
                className="font-medium text-foreground"
              >
                Zapple Pay
              </ExternalLink>{" "}
              in your browser
            </li>
            <li>
              2. Add your details (e.g. you npub, etc.), then choose{" "}
              <span className="font-medium text-foreground">Wallet</span> →{" "}
              <span className="font-medium text-foreground">
                Manual Connect
              </span>
            </li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Alby Hub</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              3. Click{" "}
              <Link
                to="/apps/new?app=zapplepay"
                className="font-medium text-foreground underline"
              >
                Connect to Zapple Pay
              </Link>
            </li>
            <li>4. Set app's wallet permissions (full access recommended)</li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Zapple Pay</h3>
          <ul className="list-inside text-muted-foreground">
            <li>5. Paste the connection secret from Alby Hub</li>
          </ul>
        </div>
      </>
    ),
  },
  {
    id: "lume",
    title: "Lume",
    description: "macOS Nostr client",
    webLink: "https://lume.nu/",
    logo: lume,
    extendedDescription:
      "Tip nostr posts and profiles and receive zaps seamlessly with your Hub",
    guide: (
      <>
        <div>
          <h3 className="font-medium">In Lume</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              1. Download Lume from{" "}
              <ExternalLink
                to="https://github.com/lumehq/lume/releases"
                className="font-medium text-foreground underline"
              >
                GitHub
              </ExternalLink>{" "}
              and install it on your computer
            </li>
            <li>
              2. Click on your profile image →{" "}
              <span className="font-medium text-foreground">Settings</span> →{" "}
              <span className="font-medium text-foreground">Wallet</span> →{" "}
              <span className="font-medium text-foreground">
                Connect Wallet
              </span>
            </li>
            <li>
              3. Choose{" "}
              <span className="font-medium text-foreground">
                Nostr Wallet Connect
              </span>
            </li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Alby Hub</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              4. Click{" "}
              <Link
                to="/apps/new?app=lume"
                className="font-medium text-foreground underline"
              >
                Connect to Lume
              </Link>
            </li>
            <li>5. Set app's wallet permissions (full access recommended)</li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Lume</h3>
          <ul className="list-inside text-muted-foreground">
            <li>6. Paste the connection secret from Alby Hub</li>
          </ul>
        </div>
      </>
    ),
  },
  {
    id: "nakapay",
    title: "NakaPay",
    description: "Non-custodial Lightning payments for businesses via NWC",
    webLink: "https://www.nakapay.app",
    logo: nakapay,
    extendedDescription:
      "Accept Bitcoin Lightning payments directly to your Hub wallet. NakaPay generates invoices that customers pay directly to you - your funds never go through NakaPay (fully non-custodial)",
    guide: (
      <>
        <div>
          <p className="text-sm text-muted-foreground mb-4">
            NakaPay is a{" "}
            <span className="font-medium text-foreground">
              fully non-custodial
            </span>{" "}
            payment service. Customer payments go directly to your Alby Hub
            wallet - NakaPay never holds your funds.
          </p>
        </div>
        <div>
          <h3 className="font-medium">Connect Your Alby Hub to NakaPay</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              1. Click{" "}
              <Link
                to="/apps/new?app=nakapay"
                className="font-medium text-foreground underline"
              >
                Connect to NakaPay
              </Link>{" "}
              to create a new app connection
            </li>
            <li>
              2. Choose "Custom" permissions and select: "Read your balance",
              "Create invoices", and "Send payments". This allows NakaPay to
              check your wallet balance, create invoices for customers, and
              process fee payments - all while your funds remain in your Hub
              wallet
            </li>
            <li>
              3. Copy the NWC connection string that starts with{" "}
              <code className="bg-muted px-1 py-0.5 rounded text-xs">
                nostr+walletconnect://
              </code>
            </li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">Set Up Your NakaPay Account</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              4. Visit{" "}
              <ExternalLink
                to="https://www.nakapay.app"
                className="font-medium text-foreground underline"
              >
                NakaPay
              </ExternalLink>{" "}
              and log in with any Lightning wallet (LNURL-auth)
            </li>
            <li>
              5. Your business account will be automatically created on first
              login
            </li>
            <li>
              6. Go to{" "}
              <span className="font-medium text-foreground">Dashboard</span> →{" "}
              <span className="font-medium text-foreground">Settings</span> →{" "}
              <span className="font-medium text-foreground">Wallet</span>
            </li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">Complete Non-Custodial Setup</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              7. Paste your Alby Hub NWC connection string in the wallet
              connection field
            </li>
            <li>8. Test the connection to verify everything works</li>
            <li>
              9. Create API keys in{" "}
              <span className="font-medium text-foreground">Dashboard</span> →{" "}
              <span className="font-medium text-foreground">
                API Management
              </span>{" "}
              for your applications
            </li>
            <li>
              10. Start accepting payments directly to your Alby Hub wallet!
            </li>
          </ul>
        </div>
      </>
    ),
  },
  // {
  //   id: "kiwi",
  //   title: "Kiwi",
  //   description: "Nostr communities",
  //   webLink: "https://nostr.kiwi/",
  //   logo: kiwi,
  //   extendedDescription: "Tip nostr posts and profiles and pay invoices seamlessly",
  //   guide: (
  //     <>
  //       <div>
  //         <h3 className="font-medium">In Kiwi</h3>
  //         <ul className="list-inside text-muted-foreground">
  //           <li>
  //             1. Open{" "}
  //             <ExternalLink
  //               to="https://nostr.kiwi"
  //               className="font-medium text-foreground underline"
  //             >
  //               nostr.kiwi
  //             </ExternalLink>{" "}
  //             in your browser
  //           </li>
  //           <li>
  //             2. Click on <span className="font-medium text-foreground">⋮</span>{" "}
  //             → <span className="font-medium text-foreground">Settings</span>
  //           </li>
  //           <li>
  //             3. Choose{" "}
  //             <span className="font-medium text-foreground">
  //               Custom Nostr Wallet Connect
  //             </span>
  //           </li>
  //         </ul>
  //       </div>
  //       <div>
  //         <h3 className="font-medium">In Alby Hub</h3>
  //         <ul className="list-inside text-muted-foreground">
  //           <li>
  //             4. Click{" "}
  //             <Link
  //               to="/apps/new?app=kiwi"
  //               className="font-medium text-foreground underline"
  //             >
  //               Connect to Kiwi
  //             </Link>
  //           </li>
  //           <li>5. Set app's wallet permissions (full access recommended)</li>
  //         </ul>
  //       </div>
  //       <div>
  //         <h3 className="font-medium">In Kiwi</h3>
  //         <ul className="list-inside text-muted-foreground">
  //           <li>6. Paste the connection secret from Alby Hub</li>
  //         </ul>
  //       </div>
  //     </>
  //   ),
  // },
  {
    id: "zappy-bird",
    title: "Zappy Bird",
    description: "Lose sats quickly",
    webLink: "https://rolznz.github.io/zappy-bird/",
    logo: zappybird,
    extendedDescription:
      "Makes a payment from your Hub each time your bird flaps its wings",
    guide: (
      <>
        <div>
          <h3 className="font-medium">In Zappy Bird</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              1. Open{" "}
              <ExternalLink
                to="https://rolznz.github.io/zappy-bird"
                className="font-medium text-foreground underline"
              >
                Zappy Bird
              </ExternalLink>{" "}
              in your browser
            </li>
            <li>
              2. Click on{" "}
              <span className="font-medium text-foreground">
                Connect Wallet
              </span>{" "}
              in the top right corner
            </li>
            <li>
              3. Choose{" "}
              <span className="font-medium text-foreground">
                Nostr Wallet Connect
              </span>{" "}
            </li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Alby Hub</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              4. Click{" "}
              <Link
                to="/apps/new?app=zappy-bird"
                className="font-medium text-foreground underline"
              >
                Connect to Zappy Bird
              </Link>
            </li>
            <li>5. Set app's wallet permissions (full access recommended)</li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Zappy Bird</h3>
          <ul className="list-inside text-muted-foreground">
            <li>6. Paste the connection secret from Alby Hub</li>
          </ul>
        </div>
      </>
    ),
  },
  {
    id: "nostur",
    title: "Nostur",
    description: "Social media",
    webLink: "https://nostur.com/",
    appleLink: "https://apps.apple.com/us/app/nostur-nostr-client/id1672780508",
    logo: nostur,
    extendedDescription:
      "Tip nostr posts and profiles and receive zaps seamlessly with your Hub",
    guide: (
      <>
        <div>
          <h3 className="font-medium">In Nostur</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              1. Download and open{" "}
              <span className="font-medium text-foreground">Nostur</span> on
              your iOS device
            </li>
            <li>
              2. Click on your profile image →{" "}
              <span className="font-medium text-foreground">Settings</span> →
              scroll to{" "}
              <span className="font-medium text-foreground">Zapping</span>
            </li>
            <li>
              3. Choose{" "}
              <span className="font-medium text-foreground">
                Custom Nostr Wallet Connect
              </span>
            </li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Alby Hub</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              4. Click{" "}
              <Link
                to="/apps/new?app=nostur"
                className="font-medium text-foreground underline"
              >
                Connect to Nostur
              </Link>
            </li>
            <li>5. Set app's wallet permissions (full access recommended)</li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Nostur</h3>
          <ul className="list-inside text-muted-foreground">
            <li>6. Paste the connection secret from Alby Hub</li>
          </ul>
        </div>
      </>
    ),
  },
  {
    id: "wherostr",
    title: "Wherostr",
    description: "Map of notes",
    webLink: "https://wherostr.social/",
    logo: wherostr,
    extendedDescription:
      "Tip nostr posts and profiles and receive zaps seamlessly with your Hub",
    guide: (
      <>
        <div>
          <h3 className="font-medium">In Wherostr</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              1. Open{" "}
              <ExternalLink
                to="https://wherostr.social"
                className="font-medium text-foreground underline"
              >
                Wherostr
              </ExternalLink>{" "}
              in your browser and log in
            </li>
            <li>
              2. Click on <span className="font-medium text-foreground">≡</span>{" "}
              → <span className="font-medium text-foreground">Settings</span> →{" "}
              <span className="font-medium text-foreground">Wallet</span>
            </li>
            <li>
              3. Choose{" "}
              <span className="font-medium text-foreground">
                Nostr Wallet Connect
              </span>
            </li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Alby Hub</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              4. Click{" "}
              <Link
                to="/apps/new?app=wherostr"
                className="font-medium text-foreground underline"
              >
                Connect to Wherostr
              </Link>
            </li>
            <li>5. Set app's wallet permissions (full access recommended)</li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Wherostr</h3>
          <ul className="list-inside text-muted-foreground">
            <li>6. Scan or paste the connection secret from Alby Hub</li>
          </ul>
        </div>
      </>
    ),
  },
  {
    id: "stackernews",
    title: "stacker news",
    description: "Like Hacker News but with Bitcoin",
    webLink: "https://stacker.news/",
    logo: stackernews,
    extendedDescription:
      "Upvote posts with sats and receive sats for your own posts directly in your Hub",
    guide: (
      <>
        <div>
          <h3 className="font-medium">In stacker news</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              1. Open{" "}
              <ExternalLink
                to="https://stacker.news"
                className="font-medium text-foreground underline"
              >
                stacker news
              </ExternalLink>{" "}
              in your browser
            </li>
            <li>
              2. Click on your username →{" "}
              <span className="font-medium text-foreground">Wallet</span> →{" "}
              <span className="font-medium text-foreground">
                Attach wallets
              </span>{" "}
              → <span className="font-medium text-foreground">Attach NWC</span>
            </li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">
            In Alby Hub: Configure the connection secret for sending
          </h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              3. Click{" "}
              <Link
                to="/apps/new?app=stackernews"
                className="font-medium text-foreground underline"
              >
                Connect to stacker news
              </Link>
            </li>
            <li>4. Set app's wallet permissions (full access recommended)</li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In stacker news</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              5. Paste the connection secret from Alby Hub into{" "}
              <span className="font-medium text-foreground">
                connection for sending
              </span>
            </li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">
            In Alby Hub: Configure the connection secret for receiving
          </h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              6. Click{" "}
              <Link
                to="/apps/new?app=stackernews"
                className="font-medium text-foreground underline"
              >
                Connect to stacker news
              </Link>
            </li>
            <li>
              7. Set app's wallet permissions:{" "}
              <span className="font-medium text-foreground">Custom</span> → only
              check{" "}
              <span className="font-medium text-foreground">
                Create invoices
              </span>{" "}
              → <span className="font-medium text-foreground">Next</span>
            </li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In stacker news</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              8. Paste the connection secret from Alby Hub into{" "}
              <span className="font-medium text-foreground">
                connection for receiving
              </span>
            </li>
          </ul>
        </div>
      </>
    ),
  },
  {
    id: "paper-scissors-hodl",
    title: "Paper Scissors HODL",
    description: "Paper Scissors Rock with bitcoin at stake",
    webLink: "https://paper-scissors-hodl.fly.dev",
    logo: paperScissorsHodl,
    extendedDescription:
      "Uses your Hub to pay to play a round, and receive the reward if you win",
    guide: (
      <>
        <div>
          <h3 className="font-medium">In Paper Scissors HODL</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              1. Open{" "}
              <ExternalLink
                to="https://paper-scissors-hodl.fly.dev/"
                className="font-medium text-foreground underline"
              >
                Paper Scissors HODL
              </ExternalLink>{" "}
              in your browser
            </li>
            <li>2. Start playing until the Bitcoin Connect screen pops up </li>
            <li>
              3. Choose{" "}
              <span className="font-medium text-foreground">
                Nostr Wallet Connect
              </span>
            </li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Alby Hub</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              4. Click{" "}
              <Link
                to="/apps/new?app=paper-scissors-hodl"
                className="font-medium text-foreground underline"
              >
                Connect to Paper Scissors HODL
              </Link>
            </li>
            <li>5. Set app's wallet permissions (full access recommended)</li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Paper Scissors HODL</h3>
          <ul className="list-inside text-muted-foreground">
            <li>6. Paste the connection secret from Alby Hub</li>
          </ul>
        </div>
      </>
    ),
  },
  {
    id: "pullthatupjamie-ai",
    title: "Pull That Up Jamie!",
    description: "Instantly pull up anything with private web search + AI",
    webLink: "https://www.pullthatupjamie.ai/",
    logo: pullthatupjamie,
    extendedDescription: "Pay from your Hub to do private AI-powered searches",
    guide: (
      <>
        <div>
          <h3 className="font-medium">In Pull That Up Jamie!</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              1. Open{" "}
              <ExternalLink
                to="https://www.pullthatupjamie.ai/"
                className="font-medium text-foreground underline"
              >
                pullthatupjamie.ai
              </ExternalLink>{" "}
              in your browser
            </li>
            <li>
              2. Click on the account dropdown and select "Connect Wallet"
            </li>
            <li>
              3. Choose{" "}
              <span className="font-medium text-foreground">
                Nostr Wallet Connect
              </span>
            </li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Alby Hub</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              4. Click{" "}
              <Link
                to="/apps/new?app=pullthatupjamie-ai"
                className="font-medium text-foreground underline"
              >
                Connect to Pull That Up Jamie!
              </Link>
            </li>
            <li>5. Set app's wallet permissions (full access recommended)</li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Pull That Up Jamie!</h3>
          <ul className="list-inside text-muted-foreground">
            <li>6. Paste the connection secret from Alby Hub</li>
          </ul>
        </div>
      </>
    ),
  },
  {
    id: "zapstore",
    title: "Zapstore",
    description: "Discover great apps through your social connections",
    webLink: "https://zapstore.dev/",
    logo: zapstore,
    extendedDescription:
      "Pay to zap apps and support their creators seamlessly with your Hub",
    guide: (
      <>
        <div>
          <h3 className="font-medium">In Zapstore</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              1. Install{" "}
              <ExternalLink
                to="https://www.zapstore.dev/"
                className="font-medium text-foreground underline"
              >
                Zapstore
              </ExternalLink>{" "}
              on your Android smartphone
            </li>
            <li>2. Open the settings</li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Alby Hub</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              3. Click{" "}
              <Link
                to="/apps/new?app=zapstore"
                className="font-medium text-foreground underline"
              >
                Connect to Zapstore
              </Link>
            </li>
            <li>4. Set app's wallet permissions (full access recommended)</li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Zapstore</h3>
          <ul className="list-inside text-muted-foreground">
            <li>5. Paste the connection secret from Alby Hub</li>
          </ul>
        </div>
      </>
    ),
  },
  {
    id: "zeus",
    title: "ZEUS",
    description: "A self-custodial Bitcoin wallet that puts you in control.",
    webLink: "https://zeusln.com/",
    playLink: "https://play.google.com/store/apps/details?id=app.zeusln.zeus",
    appleLink: "https://apps.apple.com/us/app/zeus-wallet/id1456038895",
    zapStoreLink: "https://zapstore.dev/download/",
    logo: zeus,
    extendedDescription:
      "Send and receive payments, to and from your Hub, on the go",
    guide: (
      <>
        <div>
          <h3 className="font-medium">In ZEUS</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              1. Install{" "}
              <ExternalLink
                to="https://zeusln.com/download/"
                className="font-medium text-foreground underline"
              >
                ZEUS
              </ExternalLink>{" "}
              on your Android or iOS smartphone
            </li>
            <li>2. Open the settings</li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Alby Hub</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              3. Click{" "}
              <Link
                to="/apps/new?app=zeus"
                className="font-medium text-foreground underline"
              >
                Connect to ZEUS
              </Link>
            </li>
            <li>4. Set app's wallet permissions (full access recommended)</li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In ZEUS</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              5. Scan the connection QR using the QR scanner in the bottom right
              corner of the bottom of the app, OR manually paste in the
              connection string under{" "}
              <span className="font-medium text-foreground">Menu</span> {">"}{" "}
              <span className="font-medium text-foreground">
                Connect a Wallet
              </span>{" "}
              {">"} <span className="font-medium text-foreground">+</span> after
              selecting{" "}
              <span className="font-medium text-foreground">
                Nostr Wallet Connect
              </span>{" "}
              as the{" "}
              <span className="font-medium text-foreground">
                Wallet interface
              </span>
              .
            </li>
          </ul>
        </div>
      </>
    ),
  },
  {
    id: "bitrefill",
    title: "Bitrefill",
    description: "Live on bitcoin",
    internal: true,
    webLink: "https://bitrefill.com",
    logo: bitrefill,
  },
].sort((a, b) => (a.title.toUpperCase() > b.title.toUpperCase() ? 1 : -1));
