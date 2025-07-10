import { ChevronLeftIcon } from "lucide-react";
import { useEffect, useState } from "react";
import { Outlet, useLocation, useNavigate } from "react-router-dom";
import { AlbyHubLogo } from "src/components/icons/AlbyHubLogo";
import { Button } from "src/components/ui/button.tsx";
import { useInfo } from "src/hooks/useInfo";

import AntonopoulosSVG from "public/images/quotes/antonopoulos.svg";
import BackSVG from "public/images/quotes/back.svg";
import FinneySVG from "public/images/quotes/finney.svg";
import HayekSVG from "public/images/quotes/hayek.svg";
import NakamotoSVG from "public/images/quotes/nakamoto.svg";
import ObamaSVG from "public/images/quotes/obama.svg";
import RolandSVG from "public/images/quotes/roland.svg";
import WilsonSVG from "public/images/quotes/wilson.svg";

const quotes = [
  {
    content: `This isn't about nation-states anymore. This isn't about who adopts
        bitcoin first or who adopts cryptocurrencies first, because the
        internet is adopting cryptocurrencies, and the internet is the world's
        largest economy. It is the first transnational economy, and it needs a
        transnational currency.`,
    author: "Andreas M. Antonopoulos",
    imageUrl: AntonopoulosSVG,
  },
  {
    content: `It might make sense just to get some in case it catches on. If enough people think the same way, that becomes a self fulfilling prophecy. Once it gets bootstrapped, there are so many applications if you could effortlessly pay a few cents to a website as easily as dropping coins in a vending machine.`,
    author: "Satoshi Nakamoto",
    imageUrl: NakamotoSVG,
  },
  {
    content: `Since we're all rich with bitcoins, or we will be once they're worth a million dollars like everyone expects, we ought to put some of this unearned wealth to good use.`,
    author: "Hal Finney",
    imageUrl: FinneySVG,
  },
  {
    content: `I don't believe we shall ever have a good money again before we take the thing out of the hands of government, that is, we can't take it violently out of the hands of government, all we can do is by some sly roundabout way introduce something that they can't stop.`,
    author: "Friedrich August von Hayek",
    imageUrl: HayekSVG,
  },
  {
    content: `Bitcoin is what they fear it is.`,
    author: "Cody Wilson",
    imageUrl: WilsonSVG,
  },
  {
    content: `If in fact you can't crack that at all, government can't get in then —everybody's walking around with a Swiss bank account in their pocket.`,
    author: "Barack Obama",
    imageUrl: ObamaSVG,
  },
  {
    content: `Bitcoin is the new wonder of the world, more work and human ingenuity, than went into the great pyramids of Egypt. The biggest computation ever done, a digital monument, a verifiable artefact of digital gold - the foundation of a new digital age.`,
    author: "Adam Back",
    imageUrl: BackSVG,
  },
  {
    content: `We who choose Bitcoin, are pioneers of a new world. A world filled with freedom, hope and peace.`,
    author: "Roland",
    imageUrl: RolandSVG,
  },
];

export default function TwoColumnFullScreenLayout() {
  const { data: info } = useInfo();
  const navigate = useNavigate();
  const { pathname } = useLocation();
  const [quote, setQuote] = useState(
    quotes[Math.floor(Math.random() * quotes.length)]
  );

  // Change quote on route changes
  useEffect(() => {
    setQuote(quotes[Math.floor(Math.random() * quotes.length)]);
  }, [pathname]);

  return (
    <div className="w-full lg:grid lg:h-screen lg:grid-cols-2 items-stretch text-background">
      <div className="hidden lg:flex flex-col bg-foreground justify-end p-10 gap-2 relative">
        <img
          src={quote.imageUrl}
          alt={quote.author}
          className="absolute inset-0 w-full h-full object-cover object-top opacity-30 dark:opacity-20 pointer-events-none"
        />
        <div className="flex-1 w-full h-full flex flex-col">
          <div className="flex flex-row justify-between items-center">
            <AlbyHubLogo className="text-background" />
            {info?.version && (
              <p className="text-sm text-muted-foreground">{info.version}</p>
            )}
          </div>
        </div>
        <div className="flex flex-row gap-5">
          <div className="flex flex-col justify-center">
            <p className="text-muted-foreground text-lg mb-2">
              {quote.content}
            </p>
            <p className="text-background text-sm">{quote.author}</p>
          </div>
        </div>
      </div>
      <div className="flex items-center justify-center py-12 text-foreground relative">
        {pathname.startsWith("/setup") &&
          !pathname.startsWith("/setup/finish") && (
            // show the back button on setup pages, except the setup finish page
            <Button
              type="button"
              variant="outline"
              onClick={() => {
                navigate(-1);
              }}
              className="top-4 left-4 md:top-10 md:left-10 absolute mr-4"
            >
              <ChevronLeftIcon className="size-4 mr-2" />
              Back
            </Button>
          )}
        <Outlet />
      </div>
    </div>
  );
}
