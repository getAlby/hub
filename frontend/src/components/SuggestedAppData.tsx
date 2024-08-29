import alby from "src/assets/suggested-apps/alby.png";
import amethyst from "src/assets/suggested-apps/amethyst.png";
import damus from "src/assets/suggested-apps/damus.png";
import hablanews from "src/assets/suggested-apps/habla-news.png";
import kiwi from "src/assets/suggested-apps/kiwi.png";
import lume from "src/assets/suggested-apps/lume.png";
import nostrudel from "src/assets/suggested-apps/nostrudel.png";
import nostur from "src/assets/suggested-apps/nostur.png";
import paperScissorsHodl from "src/assets/suggested-apps/paper-scissors-hodl.png";
import primal from "src/assets/suggested-apps/primal.png";
import snort from "src/assets/suggested-apps/snort.png";
import stackernews from "src/assets/suggested-apps/stackernews.png";
import uncleJim from "src/assets/suggested-apps/uncle-jim.png";
import wavelake from "src/assets/suggested-apps/wavelake.png";
import wherostr from "src/assets/suggested-apps/wherostr.png";
import yakihonne from "src/assets/suggested-apps/yakihonne.png";
import zapstream from "src/assets/suggested-apps/zap-stream.png";
import zapplanner from "src/assets/suggested-apps/zapplanner.png";
import zapplepay from "src/assets/suggested-apps/zapple-pay.png";
import zappybird from "src/assets/suggested-apps/zappy-bird.png";

export type SuggestedApp = {
  id: string;
  webLink?: string;
  internal?: boolean;
  playLink?: string;
  appleLink?: string;
  title: string;
  description: string;
  logo?: string;
};

export const suggestedApps: SuggestedApp[] = [
  {
    id: "uncle-jim",
    title: "Friends & Family",
    description: "Subaccounts for your friends and family powered by your Hub",
    internal: true,
    logo: uncleJim,
  },
  {
    id: "alby-extension",
    title: "Alby Extension",
    description: "Wallet in your browser",
    webLink: "https://getalby.com/",
    logo: alby,
  },
  {
    id: "damus",
    title: "Damus",
    description: "iOS Nostr client",
    webLink: "https://damus.io/?utm_source=getalby",
    appleLink: "https://apps.apple.com/ca/app/damus/id1628663131",
    logo: damus,
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
    logo: wavelake,
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
    description: "Blogging platform",
    webLink: "https://yakihonne.com/",
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
].sort((a, b) => (a.title.toUpperCase() > b.title.toUpperCase() ? 1 : -1));
