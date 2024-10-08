import { Link } from "react-router-dom";
import albyGo from "src/assets/suggested-apps/alby-go.png";
import alby from "src/assets/suggested-apps/alby.png";
import amethyst from "src/assets/suggested-apps/amethyst.png";
import buzzpay from "src/assets/suggested-apps/buzzpay.png";
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
    id: "alby-extension",
    title: "Alby Extension",
    description: "Wallet in your browser",
    webLink: "https://getalby.com/",
    chromeLink:
      "https://chromewebstore.google.com/detail/iokeahhehimjnekafflcihljlcjccdbe",
    firefoxLink: "https://addons.mozilla.org/en-US/firefox/addon/alby/",
    logo: alby,
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
                className="font-semibold text-foreground underline"
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
  },
  {
    id: "primal",
    title: "Primal",
    description: "Cross-platform social",
    webLink: "https://primal.net/",
    playLink:
      "https://play.google.com/store/apps/details?id=net.primal.android",
    appleLink: "https://apps.apple.com/us/app/primal/id1673134518",
    logo: primal,
  },
  {
    id: "zap-stream",
    title: "Zap Stream",
    description: "Stream and stack sats",
    webLink: "https://zap.stream/",
    logo: zapstream,
  },
  {
    id: "wavlake",
    title: "Wavlake",
    description: "Creators platform",
    webLink: "https://www.wavlake.com/",
    logo: wavlake,
  },
  {
    id: "snort",
    title: "Snort",
    description: "Web Nostr client",
    webLink: "https://snort.social/",
    logo: snort,
  },
  {
    id: "habla-news",
    title: "Habla News",
    description: "Blogging platform",
    webLink: "https://habla.news/",
    logo: hablanews,
  },
  {
    id: "nostrudel",
    title: "noStrudel",
    description: "Web Nostr client",
    webLink: "https://nostrudel.ninja/",
    logo: nostrudel,
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
  },
  {
    id: "zapplanner",
    title: "ZapPlanner",
    description: "Schedule payments",
    webLink: "https://zapplanner.albylabs.com/",
    logo: zapplanner,
  },
  {
    id: "zapplepay",
    title: "Zapple Pay",
    description: "Zap from any client",
    webLink: "https://www.zapplepay.com/",
    logo: zapplepay,
  },
  {
    id: "lume",
    title: "Lume",
    description: "macOS Nostr client",
    webLink: "https://lume.nu/",
    logo: lume,
  },
  {
    id: "kiwi",
    title: "Kiwi",
    description: "Nostr communities",
    webLink: "https://nostr.kiwi/",
    logo: kiwi,
  },
  {
    id: "zappy-bird",
    title: "Zappy Bird",
    description: "Lose sats quickly",
    webLink: "https://rolznz.github.io/zappy-bird/",
    logo: zappybird,
  },
  {
    id: "nostur",
    title: "Nostur",
    description: "Social media",
    webLink: "https://nostur.com/",
    appleLink: "https://apps.apple.com/us/app/nostur-nostr-client/id1672780508",
    logo: nostur,
  },
  {
    id: "wherostr",
    title: "Wherostr",
    description: "Map of notes",
    webLink: "https://wherostr.social/",
    logo: wherostr,
  },
  {
    id: "stackernews",
    title: "stacker news",
    description: "Like Hacker News but with Bitcoin",
    webLink: "https://stacker.news/",
    logo: stackernews,
  },
  {
    id: "paper-scissors-hodl",
    title: "Paper Scissors HODL",
    description: "Paper Scissors Rock with bitcoin at stake",
    webLink: "https://paper-scissors-hodl.fly.dev",
    logo: paperScissorsHodl,
  },
  {
    id: "alby-go",
    title: "Alby Go",
    description: "A simple mobile wallet that works great with Alby Hub",
    webLink: "https://albygo.com",
    playLink:
      "https://play.google.com/store/apps/details?id=com.getalby.mobile",
    appleLink: "https://apps.apple.com/us/app/alby-go/id6471335774",
    zapStoreLink: "https://zap.store",
    logo: albyGo,
  },
].sort((a, b) => (a.title.toUpperCase() > b.title.toUpperCase() ? 1 : -1));
