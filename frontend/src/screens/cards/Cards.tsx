import {
  ArrowUpRightIcon,
  CheckIcon,
  ClockIcon,
  CreditCardIcon,
  FingerprintIcon,
  InfoIcon,
  LinkIcon,
  ShieldCheckIcon,
  XIcon,
  ZapIcon,
} from "lucide-react";
import React from "react";
import { Link } from "react-router";
import twoFiatLogo from "src/assets/cards/2fiat.png";
import freedomiaLogo from "src/assets/cards/freedomia.png";
import redotpayLogo from "src/assets/cards/redotpay.png";
import bringinLogo from "src/assets/suggested-apps/bringin.png";
import wavespaceLogo from "src/assets/suggested-apps/wave-space.png";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import { AlbyIcon } from "src/components/icons/Alby";
import { AppleIcon } from "src/components/icons/Apple";
import { GooglePayIcon } from "src/components/icons/GooglePay";
import { VisaLogo } from "src/components/icons/VisaLogo";
import { Avatar, AvatarFallback, AvatarImage } from "src/components/ui/avatar";
import { Badge } from "src/components/ui/badge";
import { Button } from "src/components/ui/button";
import { localStorageKeys } from "src/constants";
import {
  CardCreatedDialog,
  ConnectCardDialog,
  YourCardsSection,
} from "src/screens/cards/CardManagement";
import {
  useUserCards,
  type UserCard as UserCardType,
} from "src/screens/cards/useUserCards";
import { PlusIcon } from "lucide-react";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useInfo } from "src/hooks/useInfo";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "src/components/ui/table";
import { ToggleGroup, ToggleGroupItem } from "src/components/ui/toggle-group";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "src/components/ui/tooltip";
import { cn } from "src/lib/utils";

type Region = "Global" | "US" | "EU" | "UK" | "LATAM" | "Africa" | "Asia";

type FeatureFilter =
  | "ApplePay"
  | "GooglePay"
  | "Self-custody"
  | "Lightning-native"
  | "No KYC";

type Provider = {
  id: string;
  name: string;
  url: string;
  logo: string;
  initials: string;
  network: "Visa" | "Mastercard";
  cardType: "Physical" | "Virtual";
  regions: Region[];
  applePay: boolean;
  googlePay: boolean;
  selfCustody: boolean;
  lightningNative: boolean;
  kyc: "Full" | "Light" | "None";
  timeToGet: string;
  cardCost: string;
  fees: string;
  /**
   * If set, the provider connects via its own NWC flow in the in-hub app store
   * (the row's action links to /appstore/<id>) rather than through the
   * stablecoin top-up Connect-card dialog.
   */
  appStoreId?: string;
};

const providers: Provider[] = [
  {
    id: "redotpay",
    name: "RedotPay",
    url: "https://www.redotpay.com",
    logo: redotpayLogo,
    initials: "RP",
    network: "Visa",
    cardType: "Virtual",
    regions: ["Global", "US", "UK", "EU", "LATAM", "Asia"],
    applePay: true,
    googlePay: true,
    selfCustody: false,
    lightningNative: false,
    kyc: "Light",
    timeToGet: "<10 minutes",
    cardCost: "$10",
    fees: "~2.2% + FX",
  },
  {
    id: "2fiat",
    name: "2fiat",
    url: "https://2fiat.com",
    logo: twoFiatLogo,
    initials: "2F",
    network: "Mastercard",
    cardType: "Virtual",
    regions: ["Global"],
    applePay: true,
    googlePay: true,
    selfCustody: false,
    lightningNative: false,
    kyc: "None",
    timeToGet: "Instant",
    cardCost: "$50",
    fees: "~6.8% top-up",
  },
  {
    id: "freedomia",
    name: "Freedomia",
    url: "https://freedomia.io",
    logo: freedomiaLogo,
    initials: "FR",
    network: "Visa",
    cardType: "Virtual",
    regions: ["Global"],
    applePay: false,
    googlePay: true,
    selfCustody: false,
    lightningNative: false,
    kyc: "None",
    timeToGet: "Instant",
    cardCost: "$5–30 / mo",
    fees: "1.3–4.3%",
  },
  {
    id: "bringin",
    name: "Bringin",
    url: "https://bringin.xyz",
    logo: bringinLogo,
    initials: "BR",
    network: "Visa",
    cardType: "Virtual",
    regions: ["EU"],
    applePay: true,
    googlePay: true,
    selfCustody: false,
    lightningNative: true,
    kyc: "Full",
    timeToGet: "Minutes",
    cardCost: "Free",
    fees: "~1%",
    appStoreId: "bringin",
  },
  {
    id: "wavespace",
    name: "wavecard by wave.space",
    url: "https://wave.space",
    logo: wavespaceLogo,
    initials: "WS",
    network: "Visa",
    cardType: "Virtual",
    regions: ["Global"],
    applePay: true,
    googlePay: true,
    selfCustody: false,
    lightningNative: true,
    kyc: "Full",
    timeToGet: "Minutes",
    cardCost: "Free",
    fees: "~1%",
    appStoreId: "wavespace",
  },
];

const howItWorksSteps = [
  {
    icon: CreditCardIcon,
    title: "Get a card",
    description:
      "Pick a debit card that fits your region — physical or virtual, Apple Pay / Google Pay ready.",
  },
  {
    icon: LinkIcon,
    title: "Connect it",
    description:
      "Save your card's deposit details to your hub once. We hand you a bookmarkable top-up link.",
  },
  {
    icon: ZapIcon,
    title: "Top up in seconds",
    description:
      "Pay from your Lightning balance — your card is funded in under a minute, ready to spend on the go.",
  },
];

const regionFilters: { label: string; value: Region | "All" }[] = [
  { label: "All", value: "All" },
  { label: "Global", value: "Global" },
  { label: "US", value: "US" },
  { label: "EU", value: "EU" },
  { label: "UK", value: "UK" },
  { label: "LATAM", value: "LATAM" },
  { label: "Africa", value: "Africa" },
  { label: "Asia", value: "Asia" },
];

export function Cards() {
  const { data: albyMe } = useAlbyMe();
  const { data: info } = useInfo();
  const cardholderName = (
    albyMe?.name ||
    info?.nodeAlias ||
    "Alby Hub User"
  ).toUpperCase();

  const [region, setRegion] = React.useState<Region | "All">("All");
  const [features, setFeatures] = React.useState<FeatureFilter[]>([]);
  const [heroDismissed, setHeroDismissed] = React.useState(
    () => localStorage.getItem(localStorageKeys.cardsHeroDismissed) === "true"
  );

  const dismissHero = React.useCallback(() => {
    setHeroDismissed(true);
    localStorage.setItem(localStorageKeys.cardsHeroDismissed, "true");
  }, []);

  // Card management state (lifted so Connect card lives in the header)
  const { cards, addCard } = useUserCards();
  const [connectOpen, setConnectOpen] = React.useState(false);
  // Only set immediately after creating a card — that's the one moment we
  // hold the NWC pairing URI needed to mint a usable top-up URL. Cleared on close.
  const [justCreated, setJustCreated] = React.useState<{
    card: UserCardType;
    pairingUri: string;
  } | null>(null);

  const justCreatedProvider = justCreated
    ? providers.find((p) => p.id === justCreated.card.providerId)
    : undefined;

  const filtered = providers.filter((p) => {
    if (region !== "All" && !p.regions.includes(region)) {
      return false;
    }
    for (const f of features) {
      if (f === "ApplePay" && !p.applePay) {
        return false;
      }
      if (f === "GooglePay" && !p.googlePay) {
        return false;
      }
      if (f === "Self-custody" && !p.selfCustody) {
        return false;
      }
      if (f === "Lightning-native" && !p.lightningNative) {
        return false;
      }
      if (f === "No KYC" && p.kyc !== "None") {
        return false;
      }
    }
    return true;
  });

  return (
    <>
      <AppHeader
        title="Cards"
        pageTitle="Cards"
        contentRight={
          <Button size="sm" onClick={() => setConnectOpen(true)}>
            <PlusIcon className="size-4" />
            Connect card
          </Button>
        }
      />

      {/* Hero — AI-style with 3-step strip */}
      {!heroDismissed && (
        <div className="overflow-hidden">
          <div className="bg-card text-card-foreground rounded-xl overflow-hidden relative border border-border">
            <button
              onClick={dismissHero}
              className="absolute top-4 right-4 z-10 text-muted-foreground hover:text-foreground transition-colors"
              aria-label="Dismiss"
            >
              <XIcon className="w-5 h-5" />
            </button>
            <div className="flex flex-col lg:flex-row min-h-[360px]">
              {/* Left */}
              <div className="flex-1 p-8 lg:p-12 flex flex-col justify-center">
                <p className="text-muted-foreground text-sm font-medium tracking-widest uppercase mb-4">
                  Cards + Bitcoin
                </p>
                <h2 className="text-4xl lg:text-5xl font-bold tracking-tight leading-[1.1] mb-4">
                  Got your card ready?
                  <br />
                  <span className="text-primary">Connect it.</span>
                </h2>
                <p className="text-muted-foreground text-lg max-w-md">
                  Spend Bitcoin anywhere — top up your debit card in less than a
                  minute, with one click.
                </p>
              </div>

              {/* Right — card visual */}
              <div className="flex-1 p-6 pt-12 lg:p-8 lg:pt-14 flex items-center justify-center">
                <div className="relative w-full max-w-sm aspect-[1.6/1]">
                  <div className="absolute inset-0 rounded-2xl bg-primary text-primary-foreground shadow-2xl overflow-hidden">
                    <div className="absolute inset-0 bg-gradient-to-br from-white/10 via-transparent to-black/10 pointer-events-none" />
                    <div className="relative p-6 h-full flex flex-col justify-between">
                      <div className="flex items-start justify-between">
                        <span className="text-xs font-bold tracking-wider">
                          ALBY HUB
                        </span>
                        <AlbyIcon className="size-7" />
                      </div>

                      <div className="font-mono text-lg leading-none flex items-baseline gap-3 whitespace-nowrap">
                        <span>••••</span>
                        <span>••••</span>
                        <span>••••</span>
                        <span>4421</span>
                      </div>

                      <div className="flex items-end justify-between gap-3">
                        <div className="min-w-0">
                          <p className="text-[9px] tracking-widest uppercase opacity-60 leading-none mb-1">
                            Cardholder
                          </p>
                          <p className="text-xs font-semibold tracking-wider truncate">
                            {cardholderName}
                          </p>
                        </div>
                        <VisaLogo className="h-5 w-auto" />
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            {/* Three steps */}
            <div className="border-t border-border grid grid-cols-1 sm:grid-cols-3 divide-y sm:divide-y-0 sm:divide-x divide-border">
              {howItWorksSteps.map((item) => {
                const Icon = item.icon;
                return (
                  <div key={item.title} className="p-6 lg:p-8">
                    <Icon className="w-5 h-5 text-primary mb-2" />
                    <h3 className="font-semibold text-lg mb-1">{item.title}</h3>
                    <p className="text-muted-foreground text-sm">
                      {item.description}
                    </p>
                  </div>
                );
              })}
            </div>
          </div>
        </div>
      )}

      {/* Your card connections */}
      {cards.length > 0 && (
        <div className="pt-2">
          <h2 className="text-lg font-semibold">Your card connections</h2>
          <p className="text-xs text-muted-foreground">
            Tap a card to view or revoke its connection.
          </p>
        </div>
      )}
      <YourCardsSection cards={cards} providers={providers} />

      <ConnectCardDialog
        open={connectOpen}
        onOpenChange={setConnectOpen}
        providers={providers}
        onSubmit={async (input) => {
          const result = await addCard(input);
          setConnectOpen(false);
          setJustCreated(result);
        }}
      />

      <CardCreatedDialog
        card={justCreated?.card ?? null}
        pairingUri={justCreated?.pairingUri ?? null}
        provider={justCreatedProvider}
        onOpenChange={(open) => {
          if (!open) {
            setJustCreated(null);
          }
        }}
      />

      {/* Section heading for directory */}
      <div className="pt-2">
        <h2 className="text-lg font-semibold">Get a card</h2>
        <p className="text-xs text-muted-foreground">
          Pick a provider that works in your region, then connect it above.
        </p>
      </div>

      {/* Filter bar */}
      <section className="rounded-xl border border-border bg-card overflow-hidden">
        <div className="px-5 py-4 flex flex-wrap items-center gap-x-5 gap-y-3 overflow-x-auto">
          <div className="flex items-center gap-3 shrink-0">
            <p className="text-xs font-medium text-muted-foreground">Region</p>
            <ToggleGroup
              type="single"
              value={region}
              onValueChange={(v) => v && setRegion(v as Region | "All")}
              variant="outline"
              size="sm"
              className="*:data-[state=on]:bg-primary *:data-[state=on]:text-primary-foreground *:data-[state=on]:border-primary"
            >
              {regionFilters.map((r) => (
                <ToggleGroupItem key={r.value} value={r.value}>
                  {r.label}
                </ToggleGroupItem>
              ))}
            </ToggleGroup>
          </div>

          <div className="flex items-center gap-3 shrink-0">
            <p className="text-xs font-medium text-muted-foreground">Filter</p>
            <ToggleGroup
              type="multiple"
              value={features}
              onValueChange={(v) => setFeatures(v as FeatureFilter[])}
              variant="outline"
              size="sm"
              spacing={1}
              className="*:data-[state=on]:bg-primary *:data-[state=on]:text-primary-foreground *:data-[state=on]:border-primary"
            >
              {providers.some((p) => p.applePay) && (
                <ToggleGroupItem value="ApplePay" aria-label="Apple Pay">
                  <AppleIcon />
                  Apple Pay
                </ToggleGroupItem>
              )}
              {providers.some((p) => p.googlePay) && (
                <ToggleGroupItem value="GooglePay" aria-label="Google Pay">
                  <GooglePayIcon />
                  Google Pay
                </ToggleGroupItem>
              )}
              {providers.some((p) => p.selfCustody) && (
                <ToggleGroupItem value="Self-custody" aria-label="Self-custody">
                  <ShieldCheckIcon />
                  Self-custody
                </ToggleGroupItem>
              )}
              {providers.some((p) => p.lightningNative) && (
                <ToggleGroupItem
                  value="Lightning-native"
                  aria-label="Lightning-native"
                >
                  <ZapIcon />
                  Lightning
                </ToggleGroupItem>
              )}
              {providers.some((p) => p.kyc === "None") && (
                <ToggleGroupItem value="No KYC" aria-label="No KYC">
                  <FingerprintIcon />
                  No KYC
                </ToggleGroupItem>
              )}
            </ToggleGroup>
          </div>
        </div>

        <div className="bg-muted/30 px-5 flex items-center justify-between gap-3 min-h-12">
          <p className="text-xs text-muted-foreground">
            Showing{" "}
            <span className="font-semibold text-foreground">
              {filtered.length}
            </span>{" "}
            of {providers.length} providers
          </p>
          {(region !== "All" || features.length > 0) && (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => {
                setRegion("All");
                setFeatures([]);
              }}
              className="h-7 text-xs"
            >
              <XIcon className="size-3" />
              Clear filters
            </Button>
          )}
        </div>
      </section>

      {/* Provider table */}
      <section>
        <div className="rounded-xl border border-border bg-card overflow-x-auto">
          <Table className="min-w-[720px]">
            <TableHeader>
              <TableRow className="bg-muted/40 hover:bg-muted/40">
                <TableHead className="w-[260px]">Provider</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Regions</TableHead>
                <TableHead className="text-center">Mobile</TableHead>
                <TableHead>KYC</TableHead>
                <TableHead>Time to get</TableHead>
                <TableHead>Card cost</TableHead>
                <TableHead>Fees</TableHead>
                <TableHead className="w-10" />
              </TableRow>
            </TableHeader>
            <TableBody>
              {filtered.length === 0 && (
                <TableRow>
                  <TableCell
                    colSpan={9}
                    className="text-center text-muted-foreground py-10"
                  >
                    No providers match these filters.
                  </TableCell>
                </TableRow>
              )}
              {filtered.map((p) => (
                <ProviderRow key={p.id} provider={p} />
              ))}
            </TableBody>
          </Table>
        </div>

        <p className="text-xs text-muted-foreground mt-3">
          Alby Hub does not issue or operate these cards. Availability, fees,
          and KYC are set by each provider — values here are approximate and may
          change.
        </p>
      </section>
    </>
  );
}

function ProviderRow({ provider }: { provider: Provider }) {
  return (
    <TableRow className="[&_td]:py-3">
      <TableCell>
        <ExternalLink
          to={provider.url}
          className="flex items-center gap-3 group"
        >
          <Avatar className="size-9 rounded-lg">
            <AvatarImage
              src={provider.logo}
              alt={provider.name}
              className="rounded-lg object-contain bg-secondary p-1"
            />
            <AvatarFallback className="rounded-lg bg-secondary text-secondary-foreground text-xs font-semibold">
              {provider.initials}
            </AvatarFallback>
          </Avatar>
          <div className="min-w-0">
            <div className="flex items-center gap-1.5">
              <span className="font-medium group-hover:underline truncate">
                {provider.name}
              </span>
              {provider.selfCustody && (
                <Tooltip>
                  <TooltipTrigger asChild>
                    <span>
                      <ShieldCheckIcon className="size-3 text-positive-foreground" />
                    </span>
                  </TooltipTrigger>
                  <TooltipContent>Self-custodial</TooltipContent>
                </Tooltip>
              )}
              {provider.lightningNative && (
                <Tooltip>
                  <TooltipTrigger asChild>
                    <span>
                      <ZapIcon className="size-3 text-primary" />
                    </span>
                  </TooltipTrigger>
                  <TooltipContent>Lightning-native</TooltipContent>
                </Tooltip>
              )}
            </div>
            <p className="text-xs text-muted-foreground">{provider.network}</p>
          </div>
        </ExternalLink>
      </TableCell>
      <TableCell>
        <span className="flex items-center text-xs text-muted-foreground">
          {provider.cardType}
        </span>
      </TableCell>
      <TableCell>
        <div className="flex flex-wrap items-center gap-1">
          {provider.regions.map((r) => (
            <Badge
              key={r}
              variant="secondary"
              className="text-[10px] font-medium px-1.5 py-0"
            >
              {r}
            </Badge>
          ))}
        </div>
      </TableCell>
      <TableCell className="text-center">
        <div className="flex items-center justify-center gap-2 text-muted-foreground">
          <Tooltip>
            <TooltipTrigger asChild>
              <span className={cn(provider.applePay ? "" : "opacity-20")}>
                <AppleIcon />
              </span>
            </TooltipTrigger>
            <TooltipContent>
              {provider.applePay ? "Apple Pay supported" : "No Apple Pay"}
            </TooltipContent>
          </Tooltip>
          <Tooltip>
            <TooltipTrigger asChild>
              <span className={cn(provider.googlePay ? "" : "opacity-20")}>
                <GooglePayIcon />
              </span>
            </TooltipTrigger>
            <TooltipContent>
              {provider.googlePay ? "Google Pay supported" : "No Google Pay"}
            </TooltipContent>
          </Tooltip>
        </div>
      </TableCell>
      <TableCell>
        <div className="flex items-center">
          <KycBadge kyc={provider.kyc} />
        </div>
      </TableCell>
      <TableCell>
        <span className="inline-flex items-center gap-1.5 text-xs text-muted-foreground whitespace-nowrap">
          <ClockIcon className="size-3" />
          {provider.timeToGet}
        </span>
      </TableCell>
      <TableCell>
        <div className="flex items-center text-xs text-muted-foreground">
          {provider.cardCost}
        </div>
      </TableCell>
      <TableCell>
        <div className="flex items-center text-xs text-muted-foreground">
          {provider.fees}
        </div>
      </TableCell>
      <TableCell>
        <div className="flex items-center">
          {provider.appStoreId ? (
            <Tooltip>
              <TooltipTrigger asChild>
                <Link
                  to={`/apps/new?app=${provider.appStoreId}`}
                  className="text-muted-foreground hover:text-foreground transition-colors inline-flex"
                >
                  <ArrowUpRightIcon className="size-4" />
                </Link>
              </TooltipTrigger>
              <TooltipContent>Open setup guide</TooltipContent>
            </Tooltip>
          ) : (
            <ExternalLink
              to={provider.url}
              className="text-muted-foreground hover:text-foreground transition-colors inline-flex"
            >
              <ArrowUpRightIcon className="size-4" />
            </ExternalLink>
          )}
        </div>
      </TableCell>
    </TableRow>
  );
}

function KycBadge({ kyc }: { kyc: Provider["kyc"] }) {
  if (kyc === "None") {
    return (
      <span className="inline-flex items-center gap-1 text-xs text-positive-foreground">
        <CheckIcon className="size-3" />
        None
        <Tooltip>
          <TooltipTrigger asChild>
            <span className="text-muted-foreground hover:text-foreground cursor-help">
              <InfoIcon className="size-3" />
            </span>
          </TooltipTrigger>
          <TooltipContent>
            No-KYC cards typically operate via a single merchant-of-record
            account. Privacy-friendly, but operationally fragile — the program
            can be paused or shut down without notice.
          </TooltipContent>
        </Tooltip>
      </span>
    );
  }
  if (kyc === "Light") {
    return (
      <span className="inline-flex items-center gap-1 text-xs text-muted-foreground">
        Light
        <Tooltip>
          <TooltipTrigger asChild>
            <span className="text-muted-foreground hover:text-foreground cursor-help">
              <InfoIcon className="size-3" />
            </span>
          </TooltipTrigger>
          <TooltipContent>
            ID verification only — no proof of address, employer details, or
            source-of-funds questions.
          </TooltipContent>
        </Tooltip>
      </span>
    );
  }
  return (
    <span className="inline-flex items-center gap-1 text-xs text-muted-foreground">
      {kyc}
    </span>
  );
}
