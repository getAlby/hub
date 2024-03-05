import { Link } from "react-router-dom";

import alby from "src/assets/suggested/alby.png";
import damus from "src/assets/suggested/damus.png";
import amethyst from "src/assets/suggested/amethyst.png";
import primal from "src/assets/suggested/primal.png";
import zapstream from "src/assets/suggested/zap-stream.png";
import wavelake from "src/assets/suggested/wavelake.png";
import snort from "src/assets/suggested/snort.png";
import hablanews from "src/assets/suggested/habla-news.png";
import nostrudel from "src/assets/suggested/nostrudel.png";
import yakihonne from "src/assets/suggested/yakihonne.png";
import zapplanner from "src/assets/suggested/zapplanner.png";
import zapplepay from "src/assets/suggested/zapple-pay.png";
import lume from "src/assets/suggested/lume.png";
import bc from "src/assets/suggested/bitcoin-connect.png";
import kiwi from "src/assets/suggested/kiwi.png";
import zappybird from "src/assets/suggested/zappy-bird.png";
import nostur from "src/assets/suggested/nostur.png";
import wherostr from "src/assets/suggested/wherostr.png";

const suggestedApps = [
  {
    title: "Alby Extension",
    description: "Wallet in your browser",
    link: "https://getalby.com/",
    logo: alby,
  },
  {
    title: "Damus",
    description: "iOS Nostr client",
    link: "https://damus.io/?utm_source=getalby",
    logo: damus,
  },
  {
    title: "Amethyst",
    description: "Android Nostr client",
    link: "https://play.google.com/store/apps/details?id=com.vitorpamplona.amethyst&hl=de&gl=US",
    logo: amethyst,
  },
  {
    title: "Primal",
    description: "Cross-platform social",
    link: "https://primal.net/",
    logo: primal,
  },
  {
    title: "Zap Stream",
    description: "Stream and stack sats",
    link: "https://zap.stream/",
    logo: zapstream,
  },
  {
    title: "Wavlake",
    description: "Creators platform",
    link: "https://www.wavlake.com/",
    logo: wavelake,
  },
  {
    title: "Snort",
    description: "Web Nostr client",
    link: "https://snort.social/",
    logo: snort,
  },
  {
    title: "Habla News",
    description: "Blogging platform",
    link: "https://habla.news/",
    logo: hablanews,
  },
  {
    title: "noStrudel",
    description: "Web Nostr client",
    link: "https://nostrudel.ninja/",
    logo: nostrudel,
  },
  {
    title: "YakiHonne",
    description: "Blogging platform",
    link: "https://yakihonne.com/",
    logo: yakihonne,
  },
  {
    title: "ZapPlanner",
    description: "Schedule payments",
    link: "https://zapplanner.albylabs.com/",
    logo: zapplanner,
  },
  {
    title: "Zapple Pay",
    description: "Zap from any client",
    link: "https://www.zapplepay.com/",
    logo: zapplepay,
  },
  {
    title: "Lume",
    description: "macOS Nostr client",
    link: "https://lume.nu/",
    logo: lume,
  },
  {
    title: "Bitcoin Connect",
    description: "Connect to apps",
    link: "https://bitcoin-connect.com/",
    logo: bc,
  },
  {
    title: "Kiwi",
    description: "Nostr communities",
    link: "https://nostr.kiwi/",
    logo: kiwi,
  },
  {
    title: "Zappy Bird",
    description: "Loose sats quickly",
    link: "https://rolznz.github.io/zappy-bird/",
    logo: zappybird,
  },
  {
    title: "Nostur",
    description: "Social media",
    link: "https://nostur.com/",
    logo: nostur,
  },
  {
    title: "Wherostr",
    description: "Map of notes",
    link: "https://wherostr.social/",
    logo: wherostr,
  },
];

type Props = {
  to: string;
  title: string;
  description: string;
  logo?: string;
};

function SuggestedApp({ to, title, description, logo }: Props) {
  return (
    <Link
      to={to}
      target="_blank"
      className="h-24 p-6 border border-gray-200 rounded-2xl bg-white hover:bg-gray-50 flex items-center"
    >
      <img src={logo} alt="logo" className="inline rounded-lg w-10 h-10" />
      <div className="ml-4">
        <h2 className="dark:text-white font-semibold mb-1">{title}</h2>
        <p className="text-gray-600 text-sm">{description}</p>
      </div>
    </Link>
  );
}

export default function SuggestedApps() {
  return (
    <>
      <h2 className="text-center font-medium px-8 sm:px-0">
        Use NWC to connect your wallet to any of apps below
      </h2>

      <div className="my-8 grid sm:grid-cols-2 md:grid-cols-3 gap-4">
        {suggestedApps.map((app, index) => (
          <SuggestedApp
            key={index}
            to={app.link}
            title={app.title}
            description={app.description}
            logo={app.logo}
          />
        ))}
      </div>
    </>
  );
}
