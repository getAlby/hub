import { ZapIcon } from "lucide-react";
import { Link } from "react-router-dom";
import albyGo from "src/assets/suggested-apps/alby-go.png";
import alby from "src/assets/suggested-apps/alby.png";
import amethyst from "src/assets/suggested-apps/amethyst.png";
import buzzpay from "src/assets/suggested-apps/buzzpay.png";
import clams from "src/assets/suggested-apps/clams.png";
import damus from "src/assets/suggested-apps/damus.png";
import hablanews from "src/assets/suggested-apps/habla-news.png";
import kiwi from "src/assets/suggested-apps/kiwi.png";
import lume from "src/assets/suggested-apps/lume.png";
import nostrudel from "src/assets/suggested-apps/nostrudel.png";
import nostur from "src/assets/suggested-apps/nostur.png";
import paperScissorsHodl from "src/assets/suggested-apps/paper-scissors-hodl.png";
import primal from "src/assets/suggested-apps/primal.png";
import snort from "src/assets/suggested-apps/snort.png";
import stackernews from "src/assets/suggested-apps/stacker-news.png";
import uncleJim from "src/assets/suggested-apps/uncle-jim.png";
import wavlake from "src/assets/suggested-apps/wavlake.png";
import wherostr from "src/assets/suggested-apps/wherostr.png";
import yakihonne from "src/assets/suggested-apps/yakihonne.png";
import zapstream from "src/assets/suggested-apps/zap-stream.png";
import zapplanner from "src/assets/suggested-apps/zapplanner.png";
import zapplepay from "src/assets/suggested-apps/zapple-pay.png";
import zappybird from "src/assets/suggested-apps/zappy-bird.png";
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

  guide?: React.ReactNode;
  internal?: boolean;
};

export const suggestedApps: SuggestedApp[] = [
  {
    id: "uncle-jim",
    title: "Friends & Family",
    description: "Subaccounts powered by your Hub",
    internal: true,
    logo: uncleJim,
  },
  {
    id: "buzzpay",
    title: "BuzzPay PoS",
    description: "Receive-only PoS you can safely share with your employees",
    internal: true,
    logo: buzzpay,
  },
  {
    id: "simpleboost",
    title: "SimpleBoost",
    description: "Donation widget for your website",
    internal: true,
    logo: buzzpay,
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
    playLink:
      "https://play.google.com/store/apps/details?id=net.primal.android",
    // NWC is not supported on iOS
    // appleLink: "https://apps.apple.com/us/app/primal/id1673134518",
    logo: primal,
    guide: (
      <>
        <div>
          <h3 className="font-medium">In Primal</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              1. Download and open{" "}
              <span className="font-medium text-foreground">Primal</span> on
              your Android device
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
            <li>
              3. Choose{" "}
              <span className="font-medium text-foreground">
                Connect Other Wallet
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
            <li>4. Set app's wallet permissions (full access recommended)</li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Primal</h3>
          <ul className="list-inside text-muted-foreground">
            <li>5. Scan or paste the connection secret from Alby Hub</li>
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
    id: "wavlake",
    title: "Wavlake",
    description: "Creators platform",
    webLink: "https://www.wavlake.com/",
    playLink:
      "https://play.google.com/store/apps/details?id=com.wavlake.mobile",
    appleLink: "https://testflight.apple.com/join/eWnqECG4",
    logo: wavlake,
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
    id: "nostrudel",
    title: "noStrudel",
    description: "Web Nostr client",
    webLink: "https://nostrudel.ninja/",
    logo: nostrudel,
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
    id: "kiwi",
    title: "Kiwi",
    description: "Nostr communities",
    webLink: "https://nostr.kiwi/",
    logo: kiwi,
    guide: (
      <>
        <div>
          <h3 className="font-medium">In Kiwi</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              1. Open{" "}
              <ExternalLink
                to="https://nostr.kiwi"
                className="font-medium text-foreground underline"
              >
                nostr.kiwi
              </ExternalLink>{" "}
              in your browser
            </li>
            <li>
              2. Click on <span className="font-medium text-foreground">⋮</span>{" "}
              → <span className="font-medium text-foreground">Settings</span>
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
                to="/apps/new?app=kiwi"
                className="font-medium text-foreground underline"
              >
                Connect to Kiwi
              </Link>
            </li>
            <li>5. Set app's wallet permissions (full access recommended)</li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Kiwi</h3>
          <ul className="list-inside text-muted-foreground">
            <li>6. Paste the connection secret from Alby Hub</li>
          </ul>
        </div>
      </>
    ),
  },
  {
    id: "zappy-bird",
    title: "Zappy Bird",
    description: "Lose sats quickly",
    webLink: "https://rolznz.github.io/zappy-bird/",
    logo: zappybird,
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
    id: "alby-go",
    title: "Alby Go",
    description: "A simple mobile wallet that works great with Alby Hub",
    webLink: "https://albygo.com",
    playLink:
      "https://play.google.com/store/apps/details?id=com.getalby.mobile",
    appleLink: "https://apps.apple.com/us/app/alby-go/id6471335774",
    zapStoreLink: "https://zapstore.dev",
    logo: albyGo,
    guide: (
      <>
        <div>
          <h3 className="font-medium">In Alby Go</h3>
          <ul className="list-inside text-muted-foreground">
            <li>
              1. Download and open{" "}
              <span className="font-medium text-foreground">Alby Go</span> on
              your Android or iOS device
            </li>
            <li>
              2. Click on{" "}
              <span className="font-medium text-foreground">
                Connect Wallet
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
                to="/apps/new?app=alby-go"
                className="font-medium text-foreground underline"
              >
                Connect to Alby Go
              </Link>
            </li>
            <li>5. Set app's wallet permissions (full access recommended)</li>
          </ul>
        </div>
        <div>
          <h3 className="font-medium">In Alby Go</h3>
          <ul className="list-inside text-muted-foreground">
            <li>6. Scan or paste the connection secret from Alby Hub</li>
          </ul>
        </div>
      </>
    ),
  },
].sort((a, b) => (a.title.toUpperCase() > b.title.toUpperCase() ? 1 : -1));
