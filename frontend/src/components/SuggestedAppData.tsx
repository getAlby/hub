import alby from "src/assets/suggested-apps/alby.png";
import amethyst from "src/assets/suggested-apps/amethyst.png";
import bc from "src/assets/suggested-apps/bitcoin-connect.png";
import damus from "src/assets/suggested-apps/damus.png";
import hablanews from "src/assets/suggested-apps/habla-news.png";
import kiwi from "src/assets/suggested-apps/kiwi.png";
import lume from "src/assets/suggested-apps/lume.png";
import nostrudel from "src/assets/suggested-apps/nostrudel.png";
import nostur from "src/assets/suggested-apps/nostur.png";
import primal from "src/assets/suggested-apps/primal.png";
import snort from "src/assets/suggested-apps/snort.png";
import wavelake from "src/assets/suggested-apps/wavelake.png";
import wherostr from "src/assets/suggested-apps/wherostr.png";
import yakihonne from "src/assets/suggested-apps/yakihonne.png";
import zapstream from "src/assets/suggested-apps/zap-stream.png";
import zapplanner from "src/assets/suggested-apps/zapplanner.png";
import zapplepay from "src/assets/suggested-apps/zapple-pay.png";
import zappybird from "src/assets/suggested-apps/zappy-bird.png";

export type SuggestedApp = {
  id: string;
  to: string;
  title: string;
  description: string;
  logo?: string;
};

export const suggestedApps: SuggestedApp[] = [
  {
    id: "alby-extension",
    title: "Alby Extension",
    description: "Wallet in your browser",
    to: "https://getalby.com/",
    logo: alby,
  },
  {
    id: "damus",
    title: "Damus",
    description: "iOS Nostr client",
    to: "https://damus.io/?utm_source=getalby",
    logo: damus,
  },
  {
    id: "amethyst",
    title: "Amethyst",
    description: "Android Nostr client",
    to: "https://play.google.com/store/apps/details?id=com.vitorpamplona.amethyst&hl=de&gl=US",
    logo: amethyst,
  },
  {
    id: "primal",
    title: "Primal",
    description: "Cross-platform social",
    to: "https://primal.net/",
    logo: primal,
  },
  {
    id: "zap-stream",
    title: "Zap Stream",
    description: "Stream and stack sats",
    to: "https://zap.stream/",
    logo: zapstream,
  },
  {
    id: "wavlake",
    title: "Wavlake",
    description: "Creators platform",
    to: "https://www.wavlake.com/",
    logo: wavelake,
  },
  {
    id: "snort",
    title: "Snort",
    description: "Web Nostr client",
    to: "https://snort.social/",
    logo: snort,
  },
  {
    id: "habla-news",
    title: "Habla News",
    description: "Blogging platform",
    to: "https://habla.news/",
    logo: hablanews,
  },
  {
    id: "nostrudel",
    title: "noStrudel",
    description: "Web Nostr client",
    to: "https://nostrudel.ninja/",
    logo: nostrudel,
  },
  {
    id: "yakihonne",
    title: "YakiHonne",
    description: "Blogging platform",
    to: "https://yakihonne.com/",
    logo: yakihonne,
  },
  {
    id: "zapplanner",
    title: "ZapPlanner",
    description: "Schedule payments",
    to: "https://zapplanner.albylabs.com/",
    logo: zapplanner,
  },
  {
    id: "zapplepay",
    title: "Zapple Pay",
    description: "Zap from any client",
    to: "https://www.zapplepay.com/",
    logo: zapplepay,
  },
  {
    id: "lume",
    title: "Lume",
    description: "macOS Nostr client",
    to: "https://lume.nu/",
    logo: lume,
  },
  {
    id: "bitcoin-connect",
    title: "Bitcoin Connect",
    description: "Connect to apps",
    to: "https://bitcoin-connect.com/",
    logo: bc,
  },
  {
    id: "kiwi",
    title: "Kiwi",
    description: "Nostr communities",
    to: "https://nostr.kiwi/",
    logo: kiwi,
  },
  {
    id: "zappy-bird",
    title: "Zappy Bird",
    description: "Lose sats quickly",
    to: "https://rolznz.github.io/zappy-bird/",
    logo: zappybird,
  },
  {
    id: "nostur",
    title: "Nostur",
    description: "Social media",
    to: "https://nostur.com/",
    logo: nostur,
  },
  {
    id: "wherostr",
    title: "Wherostr",
    description: "Map of notes",
    to: "https://wherostr.social/",
    logo: wherostr,
  },
].sort((a, b) => (a.title.toUpperCase() > b.title.toUpperCase() ? 1 : -1));
