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
import ExternalLink from "src/components/ExternalLink";
import {
  Card,
  CardContent,
  CardDescription,
  CardTitle,
} from "src/components/ui/card";

export type Props = {
  to: string;
  title: string;
  description: string;
  logo?: string;
};

const suggestedApps: Props[] = [
  {
    title: "Alby Extension",
    description: "Wallet in your browser",
    to: "https://getalby.com/",
    logo: alby,
  },
  {
    title: "Damus",
    description: "iOS Nostr client",
    to: "https://damus.io/?utm_source=getalby",
    logo: damus,
  },
  {
    title: "Amethyst",
    description: "Android Nostr client",
    to: "https://play.google.com/store/apps/details?id=com.vitorpamplona.amethyst&hl=de&gl=US",
    logo: amethyst,
  },
  {
    title: "Primal",
    description: "Cross-platform social",
    to: "https://primal.net/",
    logo: primal,
  },
  {
    title: "Zap Stream",
    description: "Stream and stack sats",
    to: "https://zap.stream/",
    logo: zapstream,
  },
  {
    title: "Wavlake",
    description: "Creators platform",
    to: "https://www.wavlake.com/",
    logo: wavelake,
  },
  {
    title: "Snort",
    description: "Web Nostr client",
    to: "https://snort.social/",
    logo: snort,
  },
  {
    title: "Habla News",
    description: "Blogging platform",
    to: "https://habla.news/",
    logo: hablanews,
  },
  {
    title: "noStrudel",
    description: "Web Nostr client",
    to: "https://nostrudel.ninja/",
    logo: nostrudel,
  },
  {
    title: "YakiHonne",
    description: "Blogging platform",
    to: "https://yakihonne.com/",
    logo: yakihonne,
  },
  {
    title: "ZapPlanner",
    description: "Schedule payments",
    to: "https://zapplanner.albylabs.com/",
    logo: zapplanner,
  },
  {
    title: "Zapple Pay",
    description: "Zap from any client",
    to: "https://www.zapplepay.com/",
    logo: zapplepay,
  },
  {
    title: "Lume",
    description: "macOS Nostr client",
    to: "https://lume.nu/",
    logo: lume,
  },
  {
    title: "Bitcoin Connect",
    description: "Connect to apps",
    to: "https://bitcoin-connect.com/",
    logo: bc,
  },
  {
    title: "Kiwi",
    description: "Nostr communities",
    to: "https://nostr.kiwi/",
    logo: kiwi,
  },
  {
    title: "Zappy Bird",
    description: "Lose sats quickly",
    to: "https://rolznz.github.io/zappy-bird/",
    logo: zappybird,
  },
  {
    title: "Nostur",
    description: "Social media",
    to: "https://nostur.com/",
    logo: nostur,
  },
  {
    title: "Wherostr",
    description: "Map of notes",
    to: "https://wherostr.social/",
    logo: wherostr,
  },
];

function SuggestedAppCard({ to, title, description, logo }: Props) {
  return (
    <ExternalLink to={to}>
      <Card>
        <CardContent className="pt-6">
          <div className="flex gap-3 items-center">
            <img
              src={logo}
              alt="logo"
              className="inline rounded-lg w-10 h-10"
            />
            <div className="flex-grow">
              <CardTitle>{title}</CardTitle>
              <CardDescription>{description}</CardDescription>
            </div>
          </div>
        </CardContent>
      </Card>
    </ExternalLink>
  );
}

export default function SuggestedApps() {
  return (
    <>
      <div className="grid sm:grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
        {suggestedApps.map((app) => (
          <SuggestedAppCard
            key={app.title}
            to={app.to}
            title={app.title}
            description={app.description}
            logo={app.logo}
          />
        ))}
      </div>
    </>
  );
}
