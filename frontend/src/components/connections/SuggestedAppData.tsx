import { ZapIcon } from "lucide-react";
import { Link } from "react-router-dom";
import albyExtension from "src/assets/suggested-apps/alby-extension.png";
import albyGo from "src/assets/suggested-apps/alby-go.png";
import amethyst from "src/assets/suggested-apps/amethyst.png";
import bitrefill from "src/assets/suggested-apps/bitrefill.png";
import bringin from "src/assets/suggested-apps/bringin.png";
import btcpay from "src/assets/suggested-apps/btcpay.png";
import buzzpay from "src/assets/suggested-apps/buzzpay.png";
import clams from "src/assets/suggested-apps/clams.png";
import claude from "src/assets/suggested-apps/claude.png";
import coracle from "src/assets/suggested-apps/coracle.png";
import damus from "src/assets/suggested-apps/damus.png";
import goose from "src/assets/suggested-apps/goose.png";
import hablanews from "src/assets/suggested-apps/habla-news.png";
import iris from "src/assets/suggested-apps/iris.png";
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
import tictactoe from "src/assets/suggested-apps/tictactoe.png";
import wavespace from "src/assets/suggested-apps/wave-space.png";
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
import { App } from "src/types";

export type AppStoreApp = {
  id: string;
  title: string;
  description: string;
  extendedDescription: string;

  logo?: string;
  categories: (keyof typeof appStoreCategories)[];

  // General links
  webLink?: string;

  // App store links
  playLink?: string;
  appleLink?: string;
  zapStoreLink?: string;

  // Extension store links
  chromeLink?: string;
  firefoxLink?: string;

  installGuide?: React.ReactNode;
  finalizeGuide?: React.ReactNode;
  hideConnectionQr?: boolean;
  internal?: boolean;
  superuser?: boolean;
};

export const appStoreCategories = {
  "wallet-interfaces": {
    title: "Wallet Interfaces",
    priority: 1,
  },
  "social-media": {
    title: "Social Media",
    priority: 2,
  },
  ai: {
    title: "AI",
    priority: 20,
  },
  "merchant-tools": {
    title: "Merchant Tools",
    priority: 10,
  },
  music: {
    title: "Music",
    priority: 20,
  },
  blogging: {
    title: "Blogging",
    priority: 20,
  },
  "payment-tools": {
    title: "Payment Tools",
    priority: 10,
  },
  shopping: {
    title: "Shopping",
    priority: 30,
  },
  "nostr-tools": {
    title: "Nostr Tools",
    priority: 40,
  },
  games: {
    title: "Games",
    priority: 50,
  },
  misc: {
    title: "Misc",
    priority: 100,
  },
} as const;

export const sortedAppStoreCategories = Object.entries(appStoreCategories).sort(
  (a, b) => a[1].priority - b[1].priority
);

export const appStoreApps: AppStoreApp[] = (
  [
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
      extendedDescription:
        "Sends and receives payments seamlessly from your Hub",
      categories: ["wallet-interfaces"],
      superuser: true,
      installGuide: (
        <>
          <p className="text-muted-foreground">
            Download and open{" "}
            <span className="font-medium text-foreground">Alby Go</span> on your
            iOS or Android device
          </p>
        </>
      ),
      finalizeGuide: (
        <>
          <div>
            <h3 className="font-medium">In Alby Go</h3>
            <p className="text-muted-foreground">
              Scan the connection secret from Alby Hub
            </p>
          </div>
        </>
      ),
    },
    {
      id: "buzzpay",
      title: "BuzzPay PoS",
      description: "Receive-only PoS you can safely share with your employees",
      extendedDescription:
        "Receive-only PoS you can safely share with your employees",
      internal: true,
      logo: buzzpay,
      categories: ["merchant-tools"],
      webLink: "https://pos.albylabs.com",
    },
    {
      id: "goose",
      title: "Goose",
      description:
        "Your local AI agent, automating engineering tasks seamlessly",
      internal: true,
      logo: goose,
      categories: ["ai"],
      extendedDescription:
        "Your local AI agent, automating engineering tasks seamlessly",
      webLink: "https://block.github.io/goose",
    },
    {
      id: "claude",
      title: "Claude",
      description: "AI assistant for conversations, analysis, and coding",
      extendedDescription:
        "AI assistant for conversations, analysis, and coding",
      internal: true,
      logo: claude,
      categories: ["ai"],
      webLink: "https://claude.ai/",
    },
    {
      id: "simpleboost",
      title: "SimpleBoost",
      description: "Donation widget for your website",
      extendedDescription: "Donation widget for your website",
      internal: true,
      logo: simpleboost,
      categories: ["merchant-tools"],
      webLink: "https://getalby.github.io/simple-boost/",
    },
    {
      id: "lightning-messageboard",
      title: "Lightning Messageboard",
      description: "Paid messageboard widget for your website",
      extendedDescription: "Paid messageboard widget for your website",
      internal: true,
      logo: lightningMessageboard,
      categories: ["merchant-tools"],
      webLink: "https://github.com/getAlby/lightning-messageboard",
    },
    {
      id: "alby-extension",
      title: "Alby Extension",
      description: "Wallet in your browser",
      webLink: "https://getalby.com/products/browser-extension",
      chromeLink:
        "https://chromewebstore.google.com/detail/iokeahhehimjnekafflcihljlcjccdbe",
      firefoxLink: "https://addons.mozilla.org/en-US/firefox/addon/alby/",
      logo: albyExtension,
      extendedDescription:
        "Connect your Hub to lightning-enabled websites and lets you pay seamlessly on the web",
      hideConnectionQr: true,
      installGuide: (
        <>
          <div>
            <ul className="list-inside list-decimal text-muted-foreground">
              <li>
                Download and open{" "}
                <span className="font-medium text-foreground">
                  Alby Extension
                </span>{" "}
                in your desktop browser or for Firefox mobile
              </li>
              <li>Set your Unlock Passcode</li>
              <li>
                Choose{" "}
                <span className="font-medium text-foreground">
                  Bring your own Wallet {"->"} Alby Hub
                </span>
              </li>
            </ul>
          </div>
        </>
      ),
      finalizeGuide: (
        <>
          <div>
            <ul className="list-inside list-decimal text-muted-foreground">
              <li>Copy the connection secret below</li>
              <li>
                Paste into the extension and then click{" "}
                <span className="font-medium text-foreground">Continue</span>
              </li>
            </ul>
          </div>
        </>
      ),
      categories: ["wallet-interfaces"],
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
      installGuide: (
        <>
          <p className="text-muted-foreground">
            Download and open{" "}
            <span className="font-medium text-foreground">Damus</span> on your
            iOS device
          </p>
        </>
      ),
      finalizeGuide: (
        <>
          <div>
            <h3 className="font-medium">In Damus</h3>
            <ul className="list-inside list-decimal text-muted-foreground">
              <li>
                Go to{" "}
                <span className="font-medium text-foreground">Wallet</span> →{" "}
                <span className="font-medium text-foreground">
                  Attach Wallet
                </span>
              </li>
              <li>Scan or paste the connection secret from Alby Hub</li>
            </ul>
          </div>
        </>
      ),
      categories: ["social-media"],
    },
    {
      id: "amethyst",
      title: "Amethyst",
      description: "Android Nostr client",
      webLink: "https://github.com/vitorpamplona/amethyst",
      playLink:
        "https://play.google.com/store/apps/details?id=com.vitorpamplona.amethyst",
      logo: amethyst,
      extendedDescription:
        "Tip nostr posts and profiles and receive zaps seamlessly with your Hub",
      installGuide: (
        <>
          <p className="text-muted-foreground">
            Download and open{" "}
            <span className="font-medium text-foreground">Amethyst</span> on
            your Android device
          </p>
        </>
      ),
      finalizeGuide: (
        <>
          <div>
            <h3 className="font-medium">In Amethyst</h3>
            <ul className="list-inside list-decimal text-muted-foreground">
              <li>
                Long-press
                <ZapIcon className="inline text-foreground mx-2" />
                below a post in your feed
              </li>
              <li>
                Click on the{" "}
                <span className="font-medium text-foreground">
                  QR code icon
                </span>{" "}
                to activate the QR code scanner
              </li>
              <li>Scan the connection secret from Alby Hub</li>
            </ul>
          </div>
        </>
      ),
      categories: ["social-media"],
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
      installGuide: (
        <>
          <p className="text-muted-foreground">
            Download and open{" "}
            <span className="font-medium text-foreground">Primal</span> on your
            Android or iOS device
          </p>
        </>
      ),
      finalizeGuide: (
        <>
          <div>
            <h3 className="font-medium">In Primal</h3>
            <ul className="list-inside list-decimal text-muted-foreground">
              <li>
                Click on your{" "}
                <span className="font-medium text-foreground">
                  profile image
                </span>{" "}
                in the top left corner →{" "}
                <span className="font-medium text-foreground">Settings</span> →{" "}
                <span className="font-medium text-foreground">Wallet</span> →{" "}
                <span className="font-medium text-foreground">
                  Untoggle Primal wallet
                </span>
              </li>
              <li>
                Click on "Paste NWC String" or "Scan NWC QR Code" to connect
                Alby Hub
              </li>
            </ul>
          </div>
        </>
      ),
      categories: ["social-media"],
    },
    {
      id: "zap-stream",
      title: "Zap Stream",
      description: "Stream and stack sats",
      webLink: "https://zap.stream/",
      logo: zapstream,
      extendedDescription:
        "Tip streamers, zap comments and pay or receive sats for streaming time with your Hub",
      installGuide: (
        <>
          <p className="text-muted-foreground">
            Open{" "}
            <ExternalLink
              to="https://zap.stream"
              className="font-medium text-foreground underline"
            >
              Zap Stream
            </ExternalLink>{" "}
            in your browser and log in
          </p>
        </>
      ),
      finalizeGuide: (
        <>
          <div>
            <h3 className="font-medium">In Zap Stream</h3>
            <ul className="list-inside list-decimal text-muted-foreground">
              <li>
                Click on your{" "}
                <span className="font-medium text-foreground">
                  Profile Image
                </span>{" "}
                → <span className="font-medium text-foreground">Settings</span>{" "}
                → scroll to{" "}
                <span className="font-medium text-foreground">Wallet</span>
              </li>
              <li>
                Paste connection secret from Alby Hub and click on{" "}
                <span className="font-medium text-foreground">Connect</span>
              </li>
            </ul>
          </div>
        </>
      ),
      categories: ["social-media"],
    },
    {
      id: "btcpay",
      title: "BTCPay Server",
      description: "Bitcoin payment processor",
      webLink: "https://btcpayserver.org/",
      logo: btcpay,
      extendedDescription:
        "Receive payments directly to your Hub for products you sell online",
      installGuide: (
        <>
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
        </>
      ),
      finalizeGuide: (
        <>
          <div>
            <h3 className="font-medium">In BTCPay Server</h3>
            <ul className="list-inside list-decimal text-muted-foreground">
              <li>Log in to your BTCPay Server dashboard</li>
              <li>
                Find connection configuration for your Lightning node (
                <span className="font-medium text-foreground">Lightning</span> →
                <span className="font-medium text-foreground">Settings</span> →
                <span className="font-medium text-foreground">
                  Change connection
                </span>
                )
              </li>
              <li>
                Choose{" "}
                <span className="font-medium text-foreground">
                  Use custom node
                </span>
              </li>
              <li>
                Paste the connection secret (nostr+walletconnect://....) in the
                configuration field
              </li>
              <li>
                Click <span className="font-medium text-foreground">Save</span>
              </li>
            </ul>
          </div>
        </>
      ),
      categories: ["merchant-tools"],
    },
    {
      id: "lnbits",
      title: "LNbits",
      description: "Wallet accounts system with extensions",
      webLink: "https://lnbits.com/",
      logo: lnbits,
      extendedDescription:
        "Connect your Alby Hub to LNbits to give extra functionality through plugins such as BOLT cards and lightning vouchers",
      installGuide: (
        <>
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
        </>
      ),
      finalizeGuide: (
        <>
          <div>
            <h3 className="font-medium">In LNbits</h3>
            <ul className="list-inside list-decimal text-muted-foreground">
              <li>Log in to your LNbits admin dashboard</li>
              <li>
                Go to{" "}
                <span className="font-medium text-foreground">Manage</span> →{" "}
                <span className="font-medium text-foreground">Server</span> →{" "}
                <span className="font-medium text-foreground">Funding</span>, to
                configure funding wallet
              </li>
              <li>
                Under{" "}
                <span className="font-medium text-foreground">
                  Active Funding
                </span>{" "}
                choose{" "}
                <span className="font-medium text-foreground">
                  Nostr Wallet Connect
                </span>
              </li>
              <li>
                Paste the connection secret (nostr+walletconnect://....) under{" "}
                <span className="font-medium text-foreground">Pairing URL</span>
              </li>
              <li>
                Click <span className="font-medium text-foreground">Save</span>{" "}
                and{" "}
                <span className="font-medium text-foreground">
                  Restart Server
                </span>
              </li>
            </ul>
          </div>
        </>
      ),
      categories: ["misc", "payment-tools"],
    },
    {
      id: "coracle",
      title: "Coracle.social",
      description: "Desktop Nostr client",
      webLink: "https://coracle.social/",
      logo: coracle,
      extendedDescription:
        "Tip nostr posts and profiles and receive zaps seamlessly with your Hub",
      installGuide: (
        <>
          <p className="text-muted-foreground">
            You can connect your Alby Hub to Coracle to zap Nostr notes directly
            from your node.
          </p>
        </>
      ),
      finalizeGuide: (
        <>
          <div>
            <h3 className="font-medium">In Coracle</h3>
            <ul className="list-inside list-decimal text-muted-foreground">
              <li>
                Log in with your Nostr keys to{" "}
                <ExternalLink
                  to="https://coracle.social/login"
                  className="font-medium text-foreground underline"
                >
                  Coracle
                </ExternalLink>{" "}
                (it is recommended to use the Alby Extension)
              </li>
              <li>
                Click on a zap icon ⚡ and{" "}
                <span className="font-medium text-foreground">Zap!</span> under
                any post, to configure wallet connection and make your first zap
              </li>
              <li>
                Click{" "}
                <span className="font-medium text-foreground">
                  Connect Wallet to Pay
                </span>{" "}
                and choose{" "}
                <span className="font-medium text-foreground">
                  Nostr Wallet Connect
                </span>
              </li>
              <li>
                Paste the connection secret (nostr+walletconnect://....) and
                click{" "}
                <span className="font-medium text-foreground">Connect</span>
              </li>
            </ul>
          </div>
        </>
      ),
      categories: ["social-media"],
    },
    {
      id: "nostter",
      title: "Nostter",
      description: "Minimalistic, desktop Nostr client",
      webLink: "https://nostter.app/",
      logo: nostter,
      extendedDescription:
        "Tip nostr posts and profiles and receive zaps seamlessly with your Hub",
      installGuide: (
        <>
          <p className="text-muted-foreground">
            You can connect your Alby Hub to Nostter to zap Nostr notes directly
            from your node.
          </p>
        </>
      ),
      finalizeGuide: (
        <>
          <div>
            <h3 className="font-medium">In Nostter</h3>
            <ul className="list-inside list-decimal text-muted-foreground">
              <li>
                Log in with your Nostr keys to{" "}
                <ExternalLink
                  to="https://nostter.app/"
                  className="font-medium text-foreground underline"
                >
                  Nostter
                </ExternalLink>{" "}
                (it is recommended to use the Alby Extension)
              </li>
              <li>
                Go to{" "}
                <span className="font-medium text-foreground">Preferences</span>
              </li>
              <li>
                Paste the connection secret (nostr+walletconnect://....) under{" "}
                <span className="font-medium text-foreground">
                  Nostr Wallet Connect
                </span>
              </li>
              <li>Click elsewhere for the connection to be tested and saved</li>
              <li>
                Go to <span className="font-medium text-foreground">Home</span>{" "}
                and click the zap icon (⚡) under any post to add a comment and
                send zap directly from your node
              </li>
            </ul>
          </div>
        </>
      ),
      categories: ["social-media"],
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
      installGuide: (
        <>
          <p className="text-muted-foreground">
            Download and open{" "}
            <span className="font-medium text-foreground">Wavlake</span> on your
            iOS or Android device
          </p>
        </>
      ),
      finalizeGuide: (
        <>
          <div>
            <h3 className="font-medium">In Wavlake</h3>
            <ul className="list-inside list-decimal text-muted-foreground">
              <li>Scan or paste the connection secret from Alby Hub</li>
            </ul>
          </div>
        </>
      ),
      categories: ["music"],
    },
    {
      id: "wavespace",
      title: "wavecard® by wave.space",
      description:
        "Spend Bitcoin from your AlbyHub at 150M+ merchants worldwide",
      webLink:
        "https://app.wave.space/spend/?utm_source=albyhub&affiliate=AlbyHub",
      logo: wavespace,
      extendedDescription:
        "The world's first Bitcoin VISA Debit Card that allows you to spend BTC globally, anywhere VISA is accepted – straight from the safety of your own NWC-enabled wallet. ✨ EXCLUSIVE ALBYHUB SPECIAL🐝 → Get 21% cashback on your wavecard transactions (up to 10,000 sats) using code »AlbyHub«",
      finalizeGuide: (
        <>
          <div>
            <h3 className="font-medium">In wave.space</h3>
            <ul className="list-inside list-decimal text-muted-foreground">
              <li>
                Open{" "}
                <ExternalLink
                  to="https://app.wave.space/spend/?utm_source=albyhub&affiliate=AlbyHub"
                  className="font-medium text-foreground underline"
                >
                  wave.space/spend
                </ExternalLink>{" "}
                in your browser and{" "}
                <span className="font-medium text-foreground">
                  Sign up or log in
                </span>{" "}
                to your account
              </li>
              <li>
                Click on{" "}
                <span className="font-medium text-foreground">
                  Connect Wallet
                </span>
              </li>
              <li>
                Choose{" "}
                <span className="font-medium text-foreground">
                  Nostr Wallet Connect
                </span>
              </li>
              <li>Paste the connection secret from Alby Hub</li>
            </ul>
          </div>
        </>
      ),
      categories: ["payment-tools"],
    },
    {
      id: "snort",
      title: "Snort",
      description: "Web Nostr client",
      webLink: "https://snort.social/",
      logo: snort,
      extendedDescription:
        "Tip nostr posts and profiles and receive zaps seamlessly with your Hub",
      installGuide: (
        <>
          <p className="text-muted-foreground">
            Open{" "}
            <ExternalLink
              to="https://snort.social"
              className="font-medium text-foreground underline"
            >
              Snort
            </ExternalLink>{" "}
            in your browser
          </p>
        </>
      ),
      finalizeGuide: (
        <>
          <div>
            <h3 className="font-medium">In Snort</h3>
            <ul className="list-inside list-decimal text-muted-foreground">
              <li>
                Click on{" "}
                <span className="font-medium text-foreground">Settings</span> →{" "}
                <span className="font-medium text-foreground">Wallet</span>
              </li>
              <li>
                Choose{" "}
                <span className="font-medium text-foreground">
                  Nostr Wallet Connect
                </span>
              </li>
              <li>Paste the connection secret from Alby Hub</li>
            </ul>
          </div>
        </>
      ),
      categories: ["social-media"],
    },
    {
      id: "habla-news",
      title: "Habla News",
      description: "Blogging platform",
      webLink: "https://habla.news/",
      logo: hablanews,
      extendedDescription:
        "Tip nostr posts and profiles and receive zaps seamlessly with your Hub",
      installGuide: (
        <>
          <p className="text-muted-foreground">
            Open{" "}
            <ExternalLink
              to="https://habla.news"
              className="font-medium text-foreground underline"
            >
              habla.news
            </ExternalLink>{" "}
            in your browser
          </p>
        </>
      ),
      finalizeGuide: (
        <>
          <div>
            <h3 className="font-medium">In Habla News</h3>
            <ul className="list-inside list-decimal text-muted-foreground">
              <li>
                Go to{" "}
                <span className="font-medium text-foreground">Settings</span> →{" "}
                <span className="font-medium text-foreground">Wallet</span> →{" "}
                Click{" "}
                <span className="font-medium text-foreground">
                  Connect Wallet
                </span>
              </li>
              <li>
                Choose{" "}
                <span className="font-medium text-foreground">
                  NWC Generic
                </span>{" "}
              </li>
              <li>Paste the connection secret from Alby Hub</li>
            </ul>
          </div>
        </>
      ),
      categories: ["blogging"],
    },
    {
      id: "iris",
      title: "Iris",
      description: "The nostr client for better social networks",
      webLink: "https://iris.to/",
      logo: iris,
      extendedDescription:
        "Tip nostr posts and profiles and receive zaps seamlessly with your Hub",
      installGuide: (
        <>
          <p className="text-muted-foreground">
            Open{" "}
            <ExternalLink
              to="https://iris.to"
              className="font-medium text-foreground underline"
            >
              iris.to
            </ExternalLink>{" "}
            in your browser
          </p>
        </>
      ),
      finalizeGuide: (
        <>
          <div>
            <h3 className="font-medium">In Iris</h3>
            <ul className="list-inside list-decimal text-muted-foreground">
              <li>
                Go to{" "}
                <span className="font-medium text-foreground">Settings</span> →{" "}
                <span className="font-medium text-foreground">Wallet</span> →{" "}
                Click{" "}
                <span className="font-medium text-foreground">
                  + Add NWC Wallet
                </span>
              </li>
              <li>Paste the connection secret from Alby Hub</li>
            </ul>
          </div>
        </>
      ),
      categories: ["social-media"],
    },
    {
      id: "clams",
      title: "Clams",
      description: "Multi wallet accounting tool",
      webLink: "https://clams.tech/",
      logo: clams,
      extendedDescription:
        "Get insights into your transaction history and accounting tools by connecting your Hub",
      installGuide: (
        <>
          <p className="text-muted-foreground">
            Download and open{" "}
            <ExternalLink
              to="https://clams.tech/"
              className="font-medium text-foreground underline"
            >
              Clams
            </ExternalLink>{" "}
            on your device
          </p>
        </>
      ),
      finalizeGuide: (
        <>
          <div>
            <h3 className="font-medium">In Clams</h3>
            <ul className="list-inside list-decimal text-muted-foreground">
              <li>Add a connection: "+ Add Connection" → NWC</li>
              <li>Add label & paste connection secret</li>
              <li>Click Connect and Save</li>
            </ul>
          </div>
        </>
      ),
      categories: ["payment-tools"],
    },
    {
      id: "nostrcheck-server",
      title: "Nostrcheck Server",
      description: "Sovereign Nostr services",
      webLink: "https://github.com/quentintaranpino/nostrcheck-server",
      logo: nostrcheckserver,
      extendedDescription:
        "Enable payments to your Hub from users who register or upload and download files",
      installGuide: (
        <>
          <p className="text-muted-foreground">
            Click{" "}
            <Link
              to="/apps/new?app=nostrcheck-server"
              className="font-medium text-foreground underline"
            >
              Connect to Nostrcheck Server
            </Link>
          </p>
        </>
      ),
      finalizeGuide: (
        <>
          <div>
            <h3 className="font-medium">In Nostrcheck server</h3>
            <ul className="list-inside list-decimal text-muted-foreground">
              <li>
                Go to{" "}
                <span className="font-medium text-foreground">Settings</span>{" "}
                and choose{" "}
                <span className="font-medium text-foreground">Payments</span>{" "}
                tab
              </li>
              <li>
                Scroll to Nostr wallet connect settings and paste the{" "}
                <span className="font-medium text-foreground">
                  connection secret
                </span>{" "}
                from Alby Hub
              </li>
              <li>
                Press the{" "}
                <span className="font-medium text-foreground">Save</span> button
              </li>
            </ul>
          </div>
        </>
      ),
      categories: ["nostr-tools"],
    },
    {
      id: "nostrudel",
      title: "noStrudel",
      description: "Web Nostr client",
      webLink: "https://nostrudel.ninja/",
      logo: nostrudel,
      extendedDescription:
        "Tip nostr posts and profiles and receive zaps seamlessly with your Hub",
      installGuide: (
        <>
          <p className="text-muted-foreground">
            Open{" "}
            <ExternalLink
              to="https://nostrudel.ninja"
              className="font-medium text-foreground underline"
            >
              noStrudel
            </ExternalLink>{" "}
            in your browser
          </p>
        </>
      ),
      finalizeGuide: (
        <>
          <div>
            <h3 className="font-medium">In Nostrudel</h3>
            <ul className="list-inside list-decimal text-muted-foreground">
              <li>
                Click on{" "}
                <span className="font-medium text-foreground">Settings</span> →{" "}
                <span className="font-medium text-foreground">Lightning</span> →{" "}
                <span className="font-medium text-foreground">
                  Connect Wallet
                </span>
              </li>
              <li>
                Choose{" "}
                <span className="font-medium text-foreground">
                  Custom Nostr Wallet Connect
                </span>
              </li>
              <li>Paste the connection secret from Alby Hub</li>
            </ul>
          </div>
        </>
      ),
      categories: ["social-media"],
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
      installGuide: (
        <>
          <p className="text-muted-foreground">
            Open{" "}
            <ExternalLink
              to="https://yakihonne.com/wallet"
              className="font-medium text-foreground underline"
            >
              YakiHonne
            </ExternalLink>{" "}
            in your browser
          </p>
        </>
      ),
      finalizeGuide: (
        <>
          <div>
            <h3 className="font-medium">In YakiHonne</h3>
            <ul className="list-inside list-decimal text-muted-foreground">
              <li>
                Click on{" "}
                <span className="font-medium text-foreground">Add wallet</span>
              </li>
              <li>
                Choose{" "}
                <span className="font-medium text-foreground">
                  Nostr Wallet Connect
                </span>
              </li>
              <li>Paste the connection secret from Alby Hub</li>
            </ul>
          </div>
        </>
      ),
      categories: ["social-media", "blogging"],
    },
    {
      id: "zapplanner",
      title: "ZapPlanner",
      description: "Schedule payments",
      extendedDescription: "Schedule payments to a lightning address",
      webLink: "https://zapplanner.albylabs.com/",
      logo: zapplanner,
      internal: true,
      categories: ["payment-tools"],
    },
    {
      id: "zapplepay",
      title: "Zapple Pay",
      description: "Zap from any client",
      webLink: "https://www.zapplepay.com/",
      logo: zapplepay,
      extendedDescription:
        "ZapplePay will make payments from your Hub to zap posts when you react to them",
      installGuide: (
        <>
          <p className="text-muted-foreground">
            Open{" "}
            <ExternalLink
              to="https://www.zapplepay.com"
              className="font-medium text-foreground"
            >
              Zapple Pay
            </ExternalLink>{" "}
            in your browser
          </p>
        </>
      ),
      finalizeGuide: (
        <>
          <div>
            <h3 className="font-medium">In Zapple Pay</h3>
            <ul className="list-inside list-decimal text-muted-foreground">
              <li>
                Add your details (e.g. you npub, etc.), then choose{" "}
                <span className="font-medium text-foreground">Wallet</span> →{" "}
                <span className="font-medium text-foreground">
                  Manual Connect
                </span>
              </li>
              <li>Paste the connection secret from Alby Hub</li>
            </ul>
          </div>
        </>
      ),
      categories: ["nostr-tools"],
    },
    {
      id: "lume",
      title: "Lume",
      description: "macOS Nostr client",
      webLink: "https://lume.nu/",
      logo: lume,
      extendedDescription:
        "Tip nostr posts and profiles and receive zaps seamlessly with your Hub",
      installGuide: (
        <>
          <p className="text-muted-foreground">
            Download Lume from{" "}
            <ExternalLink
              to="https://github.com/lumehq/lume/releases"
              className="font-medium text-foreground underline"
            >
              GitHub
            </ExternalLink>{" "}
            and install it on your computer
          </p>
        </>
      ),
      finalizeGuide: (
        <>
          <div>
            <h3 className="font-medium">In Lume</h3>
            <ul className="list-inside list-decimal text-muted-foreground">
              <li>
                Click on your profile image →{" "}
                <span className="font-medium text-foreground">Settings</span> →{" "}
                <span className="font-medium text-foreground">Wallet</span> →{" "}
                <span className="font-medium text-foreground">
                  Connect Wallet
                </span>
              </li>
              <li>
                Choose{" "}
                <span className="font-medium text-foreground">
                  Nostr Wallet Connect
                </span>
              </li>
              <li>Paste the connection secret from Alby Hub</li>
            </ul>
          </div>
        </>
      ),
      categories: ["social-media"],
    },
    {
      id: "nakapay",
      title: "NakaPay",
      description: "Non-custodial Lightning payments for businesses via NWC",
      webLink: "https://www.nakapay.app",
      logo: nakapay,
      extendedDescription:
        "Accept Bitcoin Lightning payments directly to your Hub wallet. NakaPay generates invoices that customers pay directly to you - your funds never go through NakaPay (fully non-custodial)",
      installGuide: (
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
            <ul className="list-inside list-decimal text-muted-foreground">
              <li>
                Visit{" "}
                <ExternalLink
                  to="https://www.nakapay.app"
                  className="font-medium text-foreground underline"
                >
                  NakaPay
                </ExternalLink>{" "}
                and log in with any Lightning wallet (LNURL-auth)
              </li>
              <li>
                Your business account will be automatically created on first
                login
              </li>
            </ul>
          </div>
        </>
      ),
      finalizeGuide: (
        <>
          <div>
            <h3 className="font-medium">In NakaPay</h3>
            <ul className="list-inside list-decimal text-muted-foreground">
              <li>
                Go to{" "}
                <span className="font-medium text-foreground">Dashboard</span> →{" "}
                <span className="font-medium text-foreground">Settings</span> →{" "}
                <span className="font-medium text-foreground">Wallet</span>
              </li>
              <li>
                Paste your Alby Hub NWC connection string in the wallet
                connection field
              </li>
              <li>Test the connection to verify everything works</li>
              <li>
                Create API keys in{" "}
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
      categories: ["merchant-tools"],
    },
    {
      id: "zappy-bird",
      title: "Zappy Bird",
      description: "Lose sats quickly",
      webLink: "https://rolznz.github.io/zappy-bird/",
      logo: zappybird,
      extendedDescription:
        "Makes a payment from your Hub each time your bird flaps its wings",
      installGuide: (
        <>
          <p className="text-muted-foreground">
            Open{" "}
            <ExternalLink
              to="https://rolznz.github.io/zappy-bird"
              className="font-medium text-foreground underline"
            >
              Zappy Bird
            </ExternalLink>{" "}
            in your browser
          </p>
        </>
      ),
      finalizeGuide: (
        <>
          <div>
            <h3 className="font-medium">In Zappy Bird</h3>
            <ul className="list-inside list-decimal text-muted-foreground">
              <li>
                Click on{" "}
                <span className="font-medium text-foreground">
                  Connect Wallet
                </span>{" "}
                in the top right corner
              </li>
              <li>
                Choose{" "}
                <span className="font-medium text-foreground">
                  Nostr Wallet Connect
                </span>{" "}
              </li>
              <li>Paste the connection secret from Alby Hub</li>
            </ul>
          </div>
        </>
      ),
      categories: ["games"],
    },
    {
      id: "tictactoe",
      title: "Tic Tac Toe",
      description:
        "Earn satoshis while playing multiplayer tic-tac-toe. Lightning network fast.",
      extendedDescription:
        "Earn satoshis while playing multiplayer tic-tac-toe. Lightning network fast.",
      webLink: "https://lntictactoe.com/",
      internal: true,
      logo: tictactoe,
      categories: ["games"],
    },
    {
      id: "nostur",
      title: "Nostur",
      description: "Social media",
      webLink: "https://nostur.com/",
      appleLink:
        "https://apps.apple.com/us/app/nostur-nostr-client/id1672780508",
      logo: nostur,
      extendedDescription:
        "Tip nostr posts and profiles and receive zaps seamlessly with your Hub",

      installGuide: (
        <>
          <p className="text-muted-foreground">
            Download and open{" "}
            <span className="font-medium text-foreground">Nostur</span> on your
            iOS device
          </p>
        </>
      ),
      finalizeGuide: (
        <>
          <div>
            <h3 className="font-medium">In Nostur</h3>
            <ul className="list-inside list-decimal text-muted-foreground">
              <li>
                Click on your profile image →{" "}
                <span className="font-medium text-foreground">Settings</span> →
                scroll to{" "}
                <span className="font-medium text-foreground">Zapping</span>
              </li>
              <li>
                Choose{" "}
                <span className="font-medium text-foreground">
                  Custom Nostr Wallet Connect
                </span>
              </li>
              <li>Paste the connection secret from Alby Hub</li>
            </ul>
          </div>
        </>
      ),
      categories: ["social-media"],
    },
    {
      id: "wherostr",
      title: "Wherostr",
      description: "Map of notes",
      webLink: "https://wherostr.social/",
      playLink:
        "https://play.google.com/store/apps/details?id=th.co.mapboss.wherostr.social.wherostr_social&hl=en&pli=1",
      appleLink: "https://apps.apple.com/us/app/wherostr/id6503808206",
      logo: wherostr,
      extendedDescription:
        "Tip nostr posts and profiles and receive zaps seamlessly with your Hub",
      installGuide: (
        <>
          <p className="text-muted-foreground">
            Open{" "}
            <ExternalLink
              to="https://wherostr.social"
              className="font-medium text-foreground underline"
            >
              Wherostr
            </ExternalLink>{" "}
            in your browser and log in
          </p>
        </>
      ),
      finalizeGuide: (
        <>
          <div>
            <h3 className="font-medium">In Wherostr</h3>
            <ul className="list-inside list-decimal text-muted-foreground">
              <li>
                Click on <span className="font-medium text-foreground">≡</span>{" "}
                → <span className="font-medium text-foreground">Settings</span>{" "}
                → <span className="font-medium text-foreground">Wallet</span>
              </li>
              <li>
                Choose{" "}
                <span className="font-medium text-foreground">
                  Nostr Wallet Connect
                </span>
              </li>
              <li>Scan or paste the connection secret from Alby Hub</li>
            </ul>
          </div>
        </>
      ),
      categories: ["social-media"],
    },
    {
      id: "stackernews",
      title: "stacker news",
      description: "Like Hacker News but with Bitcoin",
      webLink: "https://stacker.news/",
      logo: stackernews,
      extendedDescription:
        "Upvote posts with sats and receive sats for your own posts directly in your Hub",
      installGuide: (
        <>
          <p className="text-muted-foreground">
            Open{" "}
            <ExternalLink
              to="https://stacker.news"
              className="font-medium text-foreground underline"
            >
              stacker news
            </ExternalLink>{" "}
            in your browser
          </p>
        </>
      ),
      finalizeGuide: (
        <>
          <div>
            <h3 className="font-medium">In stacker news</h3>
            <ul className="list-inside list-decimal text-muted-foreground">
              <li>
                Click on your username →{" "}
                <span className="font-medium text-foreground">Wallet</span> →{" "}
                <span className="font-medium text-foreground">
                  Attach wallets
                </span>{" "}
                →{" "}
                <span className="font-medium text-foreground">Attach NWC</span>
              </li>
              <li>
                Paste the connection secret from Alby Hub into{" "}
                <span className="font-medium text-foreground">
                  connection for sending
                </span>
              </li>
              <li className="text-muted-foreground">
                You can make another stacker news connection with receive-only
                permissions, for receiving zaps.
              </li>
            </ul>
          </div>
        </>
      ),
      categories: ["social-media"],
    },
    {
      id: "paper-scissors-hodl",
      title: "Paper Scissors HODL",
      description: "Paper Scissors Rock with bitcoin at stake",
      webLink: "https://paper-scissors-hodl.fly.dev",
      logo: paperScissorsHodl,
      extendedDescription:
        "Uses your Hub to pay to play a round, and receive the reward if you win",
      installGuide: (
        <>
          <p className="text-muted-foreground">
            Open{" "}
            <ExternalLink
              to="https://paper-scissors-hodl.fly.dev/"
              className="font-medium text-foreground underline"
            >
              Paper Scissors HODL
            </ExternalLink>{" "}
            in your browser
          </p>
        </>
      ),
      finalizeGuide: (
        <>
          <div>
            <h3 className="font-medium">In Paper Scissors HODL</h3>
            <ul className="list-inside list-decimal text-muted-foreground">
              <li>Start playing until the Bitcoin Connect screen pops up </li>
              <li>
                Choose{" "}
                <span className="font-medium text-foreground">
                  Nostr Wallet Connect
                </span>
              </li>
              <li>Paste the connection secret from Alby Hub</li>
            </ul>
          </div>
        </>
      ),
      categories: ["games"],
    },
    {
      id: "pullthatupjamie-ai",
      title: "Pull That Up Jamie!",
      description: "Instantly pull up anything with private web search + AI",
      webLink: "https://www.pullthatupjamie.ai/",
      logo: pullthatupjamie,
      extendedDescription:
        "Pay from your Hub to do private AI-powered searches",
      installGuide: (
        <>
          <p className="text-muted-foreground">
            Open{" "}
            <ExternalLink
              to="https://www.pullthatupjamie.ai/"
              className="font-medium text-foreground underline"
            >
              pullthatupjamie.ai
            </ExternalLink>{" "}
            in your browser
          </p>
        </>
      ),
      finalizeGuide: (
        <>
          <div>
            <h3 className="font-medium">In Pull That Up Jamie!</h3>
            <ul className="list-inside list-decimal text-muted-foreground">
              <li>Click on the account dropdown and select "Connect Wallet"</li>
              <li>
                Choose{" "}
                <span className="font-medium text-foreground">
                  Nostr Wallet Connect
                </span>
              </li>
              <li>Paste the connection secret from Alby Hub</li>
            </ul>
          </div>
        </>
      ),
      categories: ["misc"],
    },
    {
      id: "zapstore",
      title: "Zapstore",
      description: "Discover great apps through your social connections",
      webLink: "https://zapstore.dev/",
      logo: zapstore,
      extendedDescription:
        "Pay to zap apps and support their creators seamlessly with your Hub",
      installGuide: (
        <>
          <p className="text-muted-foreground">
            Install{" "}
            <ExternalLink
              to="https://www.zapstore.dev/"
              className="font-medium text-foreground underline"
            >
              Zapstore
            </ExternalLink>{" "}
            on your Android smartphone
          </p>
        </>
      ),
      finalizeGuide: (
        <>
          <div>
            <h3 className="font-medium">In Zapstore</h3>
            <ul className="list-inside list-decimal text-muted-foreground">
              <li>Open the settings</li>
              <li>Paste the connection secret from Alby Hub</li>
            </ul>
          </div>
        </>
      ),
      categories: ["misc"],
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
      installGuide: (
        <>
          <p className="text-muted-foreground">
            Install{" "}
            <ExternalLink
              to="https://zeusln.com/download/"
              className="font-medium text-foreground underline"
            >
              ZEUS
            </ExternalLink>{" "}
            on your Android or iOS smartphone
          </p>
        </>
      ),
      finalizeGuide: (
        <>
          <div>
            <h3 className="font-medium">In ZEUS</h3>
            <ul className="list-inside list-decimal text-muted-foreground">
              <li>Open the settings</li>
              <li>
                Scan the connection QR using the QR scanner in the bottom right
                corner of the bottom of the app, OR manually paste in the
                connection string under{" "}
                <span className="font-medium text-foreground">Menu</span> {">"}{" "}
                <span className="font-medium text-foreground">
                  Connect a Wallet
                </span>{" "}
                {">"} <span className="font-medium text-foreground">+</span>{" "}
                after selecting{" "}
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
      categories: ["wallet-interfaces"],
    },
    {
      id: "bitrefill",
      title: "Bitrefill",
      description: "Live on bitcoin",
      extendedDescription: "Buy gift cards and e-sims with no KYC",
      internal: true,
      webLink: "https://bitrefill.com",
      logo: bitrefill,
      categories: ["shopping"],
    },
    {
      id: "bringin",
      title: "Bringin",
      description:
        "Spend Bitcoin from your Alby Hub anywhere VISA is accepted (150M+ merchants).",
      webLink: "https://bringin.xyz",
      playLink:
        "https://play.google.com/store/apps/details?id=xyz.bringin.client",
      appleLink: "https://testflight.apple.com/join/HVh6eZsF",
      zapStoreLink: "https://zapstore.dev/download/",
      logo: bringin,
      extendedDescription:
        "Bringin turns your Alby Hub into a global spending account. Connect once and instantly swap sats into Euro with a linked IBAN and debit card. Spend Bitcoin at 150M+ VISA merchants worldwide, or pay directly with Euro IBAN transfers—all while keeping your funds in self-custody via Alby Hub. Lightning in, real-world payments out.",
      categories: ["payment-tools"],
      installGuide: (
        <>
          <p className="text-muted-foreground">
            Install Bringin on your Android or iOS smartphone
          </p>
        </>
      ),
      finalizeGuide: (
        <>
          <div>
            <h3 className="font-medium">Connect Alby Hub to Bringin</h3>
            <ul className="list-inside list-decimal text-muted-foreground">
              <li>
                Log in to Bringin, open the{" "}
                <span className="font-medium text-foreground">Bitcoin</span>{" "}
                tab, and tap{" "}
                <span className="font-medium text-foreground">Get started</span>
                .
              </li>
              <li>
                Select{" "}
                <span className="font-medium text-foreground">Alby Hub</span>{" "}
                and paste your{" "}
                <span className="font-medium text-foreground">NWC secret</span>.
              </li>
              <li>
                That's it—your Alby Hub is linked. Use Bringin to swap sats to
                Euro and spend at 150M+ VISA merchants worldwide.
              </li>
            </ul>
          </div>
        </>
      ),
    },
  ] satisfies AppStoreApp[]
).sort((a, b) => (a.title.toUpperCase() > b.title.toUpperCase() ? 1 : -1));

export const getAppStoreApp = (app: App) => {
  return appStoreApps.find(
    (suggestedApp) =>
      suggestedApp.id === (app.metadata?.app_store_app_id ?? "") ||
      app.name.includes(suggestedApp.title)
  );
};
