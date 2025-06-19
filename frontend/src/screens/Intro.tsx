import { EmblaCarouselType } from "embla-carousel";
import {
  ArrowRightIcon,
  CloudLightningIcon,
  LucideIcon,
  ShieldCheckIcon,
  WalletIcon,
} from "lucide-react";
import React, { ReactElement } from "react";
import { useNavigate } from "react-router-dom";
import Cloud from "src/assets/images/cloud.png";
import Cloud2 from "src/assets/images/cloud2.png";
import { Button } from "src/components/ui/button";
import {
  Carousel,
  CarouselApi,
  CarouselContent,
  CarouselDots,
  CarouselItem,
} from "src/components/ui/carousel";
import { useTheme } from "src/components/ui/theme-provider";
import { useInfo } from "src/hooks/useInfo";
import { cn } from "src/lib/utils";

export function Intro() {
  const { data: info } = useInfo();
  const navigate = useNavigate();
  const [api, setApi] = React.useState<CarouselApi>();
  const [progress, setProgress] = React.useState<number>(0);
  const { setDarkMode } = useTheme();
  const windowWidth = useWindowWidth();

  React.useEffect(() => {
    // Force dark mode on intro screen
    setDarkMode("dark");
    return () => {
      // Revert to default after exiting intro
      setDarkMode("system");
    };
  }, [setDarkMode]);

  React.useEffect(() => {
    if (!info?.setupCompleted) {
      return;
    }
    navigate("/");
  }, [info, navigate]);

  React.useEffect(() => {
    api?.on("scroll", (x) => {
      setProgress(x.scrollProgress());
    });
  }, [api]);

  const cloudDesktopSizePositionModifier = Math.floor(
    Math.max(1920 - windowWidth, 0) * 0.1
  );

  return (
    <Carousel className={cn("w-full bg-background")} setApi={setApi}>
      <div
        className="w-full h-full absolute top-0 left-0 bg-no-repeat"
        style={{
          backgroundImage: `url(${Cloud})`,
          backgroundPositionX: `calc(${-Math.max(progress, 0) * 120}px - ${windowWidth * 0.06}px - ${cloudDesktopSizePositionModifier * 4}px)`,
        }}
      />
      <div
        className="w-full h-full absolute top-0 left-0 bg-no-repeat"
        style={{
          backgroundImage: `url(${Cloud2})`,
          backgroundPositionX: `calc(${-Math.max(progress, 0) * 120}px + ${windowWidth * 0.5}px + ${Math.floor(cloudDesktopSizePositionModifier * 0.1)}px)`,
          backgroundPositionY: "100%",
        }}
      />
      <CarouselContent className="select-none bg-transparent">
        <CarouselItem>
          <div className="flex flex-col justify-center items-center h-screen p-5">
            <div className="flex flex-col gap-4 text-center max-w-lg">
              <div className="text-4xl font-extrabold text-foreground">
                Welcome to Alby Hub
              </div>
              <div className="text-2xl text-muted-foreground font-semibold">
                A powerful, all-in-one bitcoin lightning wallet with the
                superpower of connecting to applications.
              </div>
              <div className="mt-20">
                <Button onClick={() => api?.scrollNext()} size="lg">
                  Next
                </Button>
              </div>
            </div>
          </div>
        </CarouselItem>
        <CarouselItem>
          <Slide
            api={api}
            icon={CloudLightningIcon}
            title="Anywhere & Anytime"
            description="Your wallet is always online and ready to use on any device."
          />
        </CarouselItem>
        <CarouselItem>
          <Slide
            api={api}
            icon={ShieldCheckIcon}
            title="Your Keys Are Safe"
            description="Your wallet is encrypted by a password of your choice. No one can access your funds but you."
          />
        </CarouselItem>
        <CarouselItem>
          <Slide
            api={api}
            icon={WalletIcon}
            title="Take Your Wallet With You"
            description="Connect your wallet to dozens of apps and participate in the bitcoin digital economy."
          />
        </CarouselItem>
      </CarouselContent>
      <div className="absolute bottom-5 left-1/2 -translate-x-1/2">
        <CarouselDots />
      </div>
    </Carousel>
  );
}

function Slide({
  api,
  title,
  description,
  icon: Icon,
}: {
  api: EmblaCarouselType | undefined;
  title: string;
  description: string;
  icon: LucideIcon;
  button?: ReactElement;
}) {
  const navigate = useNavigate();

  const slideNext = function () {
    if (api?.canScrollNext()) {
      api.scrollNext();
    } else {
      navigate("/welcome");
    }
  };

  return (
    <div className="flex flex-col justify-center items-center h-screen gap-8 p-5">
      <Icon className="w-16 h-16 text-primary-background" />
      <div className="flex flex-col gap-4 text-center items-center max-w-lg">
        <div className="text-3xl font-semibold text-primary-background">
          {title}
        </div>
        <div className="text-lg text-muted-foreground font-semibold">
          {description}
        </div>
      </div>
      <Button size="icon" onClick={slideNext} className="">
        <ArrowRightIcon className="w-4 h-4" />
      </Button>
    </div>
  );
}

function useWindowWidth() {
  const [width, setWidth] = React.useState(window.innerWidth);
  React.useLayoutEffect(() => {
    function updateSize() {
      setWidth(window.innerWidth);
    }
    window.addEventListener("resize", updateSize);
    updateSize();
    return () => window.removeEventListener("resize", updateSize);
  }, []);
  return width;
}
