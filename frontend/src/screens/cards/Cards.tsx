import {
  ArrowUpRightIcon,
  CheckIcon,
  FingerprintIcon,
  GlobeIcon,
  ShieldCheckIcon,
  SmartphoneIcon,
  SparklesIcon,
  XIcon,
  ZapIcon,
} from "lucide-react";
import React from "react";
import twoFiatLogo from "src/assets/cards/2fiat.png";
import bitnobLogo from "src/assets/cards/bitnob.png";
import bitrefillLogo from "src/assets/cards/bitrefill.png";
import cryptoComLogo from "src/assets/cards/cryptocom.png";
import foldLogo from "src/assets/cards/fold.png";
import freedomiaLogo from "src/assets/cards/freedomia.png";
import gnosisPayLogo from "src/assets/cards/gnosispay.png";
import metamaskLogo from "src/assets/cards/metamask.png";
import moonLogo from "src/assets/cards/moon.png";
import redotpayLogo from "src/assets/cards/redotpay.png";
import wirexLogo from "src/assets/cards/wirex.png";
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
  | "No KYC"
  | "Experimental";

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
  topUpVia: string;
  fees: string;
  experimental?: boolean;
};

const providers: Provider[] = [
  {
    id: "redotpay",
    name: "RedotPay",
    url: "https://url.hub.so/redotpay",
    logo: redotpayLogo,
    initials: "RP",
    network: "Visa",
    cardType: "Physical",
    regions: ["Global", "EU", "LATAM", "Asia"],
    applePay: true,
    googlePay: true,
    selfCustody: false,
    lightningNative: false,
    kyc: "Light",
    topUpVia: "USDT (TRC20)",
    fees: "~1.5% + FX",
  },
  {
    id: "metamask",
    name: "MetaMask Card",
    url: "https://card.metamask.io",
    logo: metamaskLogo,
    initials: "MM",
    network: "Mastercard",
    cardType: "Physical",
    regions: ["Global", "EU", "LATAM"],
    applePay: true,
    googlePay: true,
    selfCustody: true,
    lightningNative: false,
    kyc: "Light",
    topUpVia: "USDC on Linea",
    fees: "~1% + FX",
  },
  {
    id: "gnosispay",
    name: "Gnosis Pay",
    url: "https://gnosispay.com",
    logo: gnosisPayLogo,
    initials: "GP",
    network: "Visa",
    cardType: "Physical",
    regions: ["EU", "UK"],
    applePay: true,
    googlePay: true,
    selfCustody: true,
    lightningNative: false,
    kyc: "Full",
    topUpVia: "EURe / GBPe on Gnosis Chain",
    fees: "0% top-up · FX 2%",
  },
  {
    id: "bitnob",
    name: "Bitnob",
    url: "https://bitnob.com",
    logo: bitnobLogo,
    initials: "BN",
    network: "Mastercard",
    cardType: "Virtual",
    regions: ["Africa", "LATAM"],
    applePay: false,
    googlePay: false,
    selfCustody: false,
    lightningNative: true,
    kyc: "Light",
    topUpVia: "Direct Lightning",
    fees: "Free top-up · FX 1%",
  },
  {
    id: "bitrefill",
    name: "Bitrefill Virtual Visa",
    url: "https://bitrefill.com",
    logo: bitrefillLogo,
    initials: "BR",
    network: "Visa",
    cardType: "Virtual",
    regions: ["Global"],
    applePay: false,
    googlePay: false,
    selfCustody: false,
    lightningNative: true,
    kyc: "None",
    topUpVia: "Direct Lightning",
    fees: "~3% spread",
  },
  {
    id: "moon",
    name: "Moon",
    url: "https://paywithmoon.com",
    logo: moonLogo,
    initials: "MN",
    network: "Visa",
    cardType: "Virtual",
    regions: ["Global"],
    applePay: false,
    googlePay: false,
    selfCustody: false,
    lightningNative: true,
    kyc: "None",
    topUpVia: "Direct Lightning",
    fees: "~3–5% spread",
  },
  {
    id: "cryptocom",
    name: "Crypto.com Visa",
    url: "https://crypto.com/cards",
    logo: cryptoComLogo,
    initials: "CC",
    network: "Visa",
    cardType: "Physical",
    regions: ["Global", "US", "EU"],
    applePay: true,
    googlePay: true,
    selfCustody: false,
    lightningNative: false,
    kyc: "Full",
    topUpVia: "BTC / USDC deposit",
    fees: "0% top-up · FX 0.5%",
  },
  {
    id: "fold",
    name: "Fold Card",
    url: "https://foldapp.com",
    logo: foldLogo,
    initials: "FD",
    network: "Visa",
    cardType: "Physical",
    regions: ["US"],
    applePay: true,
    googlePay: true,
    selfCustody: false,
    lightningNative: false,
    kyc: "Full",
    topUpVia: "On-chain BTC",
    fees: "Free + swap fee",
  },
  {
    id: "wirex",
    name: "Wirex",
    url: "https://wirexapp.com",
    logo: wirexLogo,
    initials: "WX",
    network: "Visa",
    cardType: "Physical",
    regions: ["EU", "UK", "Asia"],
    applePay: true,
    googlePay: true,
    selfCustody: false,
    lightningNative: false,
    kyc: "Full",
    topUpVia: "BTC deposit",
    fees: "~1% + FX",
  },
  {
    id: "2fiat",
    name: "2fiat",
    url: "https://2fiat.com",
    logo: twoFiatLogo,
    initials: "2F",
    network: "Visa",
    cardType: "Virtual",
    regions: ["Global"],
    applePay: false,
    googlePay: false,
    selfCustody: false,
    lightningNative: false,
    kyc: "Light",
    topUpVia: "Crypto deposit",
    fees: "~3% top-up",
    experimental: true,
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
    googlePay: false,
    selfCustody: false,
    lightningNative: false,
    kyc: "None",
    topUpVia: "Crypto deposit",
    fees: "~3% top-up",
    experimental: true,
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

  const showExperimental = features.includes("Experimental");

  const filtered = providers.filter((p) => {
    if (p.experimental && !showExperimental) {
      return false;
    }
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

  const experimentalCount = providers.filter((p) => p.experimental).length;

  return (
    <>
      <AppHeader title="Cards" pageTitle="Cards" />

      {/* Compact hero */}
      {!heroDismissed && (
        <section className="relative overflow-hidden rounded-xl border border-border bg-card">
          <div className="absolute inset-0 bg-gradient-to-br from-primary/15 via-primary/5 to-transparent pointer-events-none" />
          <button
            onClick={dismissHero}
            className="absolute top-4 right-4 z-10 text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
            aria-label="Dismiss"
          >
            <XIcon className="size-5" />
          </button>
          <div className="relative grid grid-cols-1 md:grid-cols-[1fr_auto] gap-6 items-center p-6 md:p-8">
            <div className="max-w-xl">
              <h2 className="text-2xl md:text-3xl font-bold tracking-tight leading-tight">
                Pay with Bitcoin{" "}
                <span className="text-primary">at any checkout</span>
              </h2>
              <p className="text-muted-foreground mt-2 text-sm md:text-base">
                Connect a debit card to your hub once, then top it up from
                Lightning in one tap. No address copy-paste, no wrong-network
                mistakes, no surprises at the till.
              </p>
              <div className="flex items-center gap-4 mt-4 text-xs text-muted-foreground whitespace-nowrap">
                <span className="inline-flex items-center gap-1.5">
                  <GlobeIcon className="size-3.5 text-primary" />
                  {providers.length} cards worldwide
                </span>
                <span className="inline-flex items-center gap-1.5">
                  <SmartphoneIcon className="size-3.5 text-primary" />
                  Apple Pay & Google Pay
                </span>
                <span className="inline-flex items-center gap-1.5">
                  <ShieldCheckIcon className="size-3.5 text-primary" />
                  Self-custody
                </span>
              </div>
            </div>

            <div className="hidden md:block md:pr-2 lg:pr-4">
              <div className="relative w-72 aspect-[1.6/1]">
                <div className="absolute inset-0 rounded-2xl bg-primary text-primary-foreground shadow-xl overflow-hidden">
                  {/* subtle highlight */}
                  <div className="absolute inset-0 bg-gradient-to-br from-white/10 via-transparent to-black/10 pointer-events-none" />
                  <div className="relative p-5 h-full flex flex-col justify-between">
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
        </section>
      )}

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
              <ToggleGroupItem value="ApplePay" aria-label="Apple Pay">
                <AppleIcon />
                Apple Pay
              </ToggleGroupItem>
              <ToggleGroupItem value="GooglePay" aria-label="Google Pay">
                <GooglePayIcon />
                Google Pay
              </ToggleGroupItem>
              <ToggleGroupItem value="Self-custody" aria-label="Self-custody">
                <ShieldCheckIcon />
                Self-custody
              </ToggleGroupItem>
              <ToggleGroupItem
                value="Lightning-native"
                aria-label="Lightning-native"
              >
                <ZapIcon />
                Lightning
              </ToggleGroupItem>
              <ToggleGroupItem value="No KYC" aria-label="No KYC">
                <FingerprintIcon />
                No KYC
              </ToggleGroupItem>
              <ToggleGroupItem value="Experimental" aria-label="Experimental">
                <SparklesIcon />
                Experimental ({experimentalCount})
              </ToggleGroupItem>
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
        <div className="rounded-xl border border-border bg-card overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow className="bg-muted/40 hover:bg-muted/40">
                <TableHead className="w-[260px]">Provider</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Regions</TableHead>
                <TableHead className="text-center">Mobile pay</TableHead>
                <TableHead>KYC</TableHead>
                <TableHead>Top up via</TableHead>
                <TableHead>Fees</TableHead>
                <TableHead className="w-10" />
              </TableRow>
            </TableHeader>
            <TableBody>
              {filtered.length === 0 && (
                <TableRow>
                  <TableCell
                    colSpan={8}
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
              {provider.experimental && (
                <Badge variant="outline" className="text-[10px] px-1 py-0">
                  Experimental
                </Badge>
              )}
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
        <div className="flex items-center text-xs text-muted-foreground">
          {provider.topUpVia}
        </div>
      </TableCell>
      <TableCell>
        <div className="flex items-center text-xs text-muted-foreground">
          {provider.fees}
        </div>
      </TableCell>
      <TableCell>
        <div className="flex items-center">
          <ExternalLink
            to={provider.url}
            className="text-muted-foreground hover:text-foreground transition-colors inline-flex"
          >
            <ArrowUpRightIcon className="size-4" />
          </ExternalLink>
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
      </span>
    );
  }
  return (
    <span className="inline-flex items-center gap-1 text-xs text-muted-foreground">
      {kyc}
    </span>
  );
}
