import {
  ArrowUpRightIcon,
  CheckIcon,
  ClockIcon,
  CreditCardIcon,
  FingerprintIcon,
  InfoIcon,
  LinkIcon,
  PlusIcon,
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
import { AlbyIcon } from "src/components/icons/Alby";
import { AppleIcon } from "src/components/icons/Apple";
import { GooglePayIcon } from "src/components/icons/GooglePay";
import { VisaLogo } from "src/components/icons/VisaLogo";
import { Avatar, AvatarFallback, AvatarImage } from "src/components/ui/avatar";
import { Badge } from "src/components/ui/badge";
import { Button } from "src/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "src/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "src/components/ui/select";
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
import { localStorageKeys } from "src/constants";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useInfo } from "src/hooks/useInfo";
import { cn } from "src/lib/utils";
import { sendEvent } from "src/utils/sendEvent";

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
  cardType: "Physical" | "Virtual" | "Both";
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
    regions: ["Global"],
    applePay: true,
    googlePay: true,
    selfCustody: false,
    lightningNative: false,
    kyc: "Light",
    timeToGet: "<10 minutes",
    cardCost: "$10",
    fees: "~2.2% + FX",
    appStoreId: "bitcoin-card-topup",
  },
  {
    id: "2fiat",
    name: "2fiat",
    url: "https://2fiat.com/getalby",
    logo: twoFiatLogo,
    initials: "2F",
    network: "Mastercard",
    cardType: "Virtual",
    regions: ["Global"],
    applePay: true,
    googlePay: true,
    selfCustody: false,
    lightningNative: true,
    kyc: "None",
    timeToGet: "Instant",
    cardCost: "$50",
    fees: "~6.8% top-up + $0.50",
    appStoreId: "2fiat",
  },
  {
    id: "freedomia",
    name: "Freedomia",
    url: "https://www.freedomia.io/a/getalby",
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
    appStoreId: "bitcoin-card-topup",
  },
  {
    id: "bringin",
    name: "Bringin",
    url: "https://bringin.app",
    logo: bringinLogo,
    initials: "BR",
    network: "Visa",
    cardType: "Both",
    regions: ["EU"],
    // Direct Apple Pay isn't live — usable today via Curve only, so don't
    // claim it. Google/Samsung/Fitbit/Garmin Pay are direct.
    applePay: false,
    googlePay: true,
    selfCustody: false,
    lightningNative: true,
    kyc: "Full",
    timeToGet: "Minutes",
    // Their cards page advertises a €3.49/mo subscription that bundles
    // both cards. Worth re-confirming during checkout — sources disagree.
    cardCost: "€3.49 / mo",
    fees: "1% + 0.5%",
    appStoreId: "bringin",
  },
  {
    id: "wavespace",
    name: "wavecard by wave.space",
    url: "https://app.wave.space/spend/?utm_source=albyhub&affiliate=AlbyHub",
    logo: wavespaceLogo,
    initials: "WS",
    network: "Visa",
    cardType: "Both",
    // Card is accepted worldwide, but issuance requires EEA residency.
    regions: ["EU"],
    applePay: false,
    googlePay: true,
    selfCustody: false,
    lightningNative: true,
    kyc: "Light",
    timeToGet: "Minutes",
    cardCost: "€2.99 / €29.99",
    fees: "1% + 0.5%",
    appStoreId: "wavespace",
  },
];

const howItWorksSteps = [
  {
    icon: CreditCardIcon,
    title: "Get a card",
    description:
      "Pick a debit card that fits your region. Physical or virtual, Apple Pay or Google Pay ready.",
  },
  {
    icon: LinkIcon,
    title: "Get a top-up link",
    description:
      "Connect your card once and save the top-up link to your phone. Open it anytime to fund your card.",
  },
  {
    icon: ZapIcon,
    title: "Top up in seconds",
    description:
      "Pay from your Lightning balance. Your card is funded in seconds, ready to spend.",
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
  const [connectOpen, setConnectOpen] = React.useState(false);
  const [heroDismissed, setHeroDismissed] = React.useState(
    () => localStorage.getItem(localStorageKeys.cardsHeroDismissed) === "true"
  );

  const dismissHero = React.useCallback(() => {
    setHeroDismissed(true);
    localStorage.setItem(localStorageKeys.cardsHeroDismissed, "true");
  }, []);

  const filtered = providers.filter((p) => {
    if (
      region !== "All" &&
      !p.regions.includes(region) &&
      !(region !== "Global" && p.regions.includes("Global"))
    ) {
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

      <ConnectCardDialog
        open={connectOpen}
        onOpenChange={setConnectOpen}
        providers={providers}
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
                  Spend Bitcoin anywhere. Top up your debit card from your
                  Lightning balance in under a minute.
                </p>
              </div>

              {/* Right — card visual */}
              <div className="flex-1 p-6 pt-12 lg:p-8 lg:pt-14 flex items-center justify-center">
                <div className="relative w-full max-w-sm aspect-[1.6/1]">
                  <div className="absolute inset-0 rounded-2xl bg-primary text-primary-foreground shadow-2xl overflow-hidden">
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

      {/* Filter bar */}
      <section className="rounded-xl border border-border bg-card overflow-hidden">
        <div className="px-5 py-4 flex flex-col sm:flex-row sm:flex-wrap sm:items-center gap-x-5 gap-y-3">
          <div className="flex items-center gap-3 shrink-0">
            <p className="text-xs font-medium text-muted-foreground">Region</p>
            {/* Mobile: dropdown to avoid a horizontally-scrolling pill row */}
            <Select
              value={region}
              onValueChange={(v) => setRegion(v as Region | "All")}
            >
              <SelectTrigger className="h-8 flex-1 sm:hidden">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {regionFilters.map((r) => (
                  <SelectItem key={r.value} value={r.value}>
                    {r.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {/* Desktop: connected pills */}
            <ToggleGroup
              type="single"
              value={region}
              onValueChange={(v) => v && setRegion(v as Region | "All")}
              variant="outline"
              size="sm"
              className="hidden sm:flex *:data-[state=on]:bg-primary *:data-[state=on]:text-primary-foreground *:data-[state=on]:border-primary"
            >
              {regionFilters.map((r) => (
                <ToggleGroupItem key={r.value} value={r.value}>
                  {r.label}
                </ToggleGroupItem>
              ))}
            </ToggleGroup>
          </div>

          <div className="flex items-center gap-3">
            <p className="text-xs font-medium text-muted-foreground">Filter</p>
            <ToggleGroup
              type="multiple"
              value={features}
              onValueChange={(v) => setFeatures(v as FeatureFilter[])}
              variant="outline"
              size="sm"
              spacing={1}
              className="flex-wrap flex-1 min-w-0 *:data-[state=on]:bg-primary *:data-[state=on]:text-primary-foreground *:data-[state=on]:border-primary"
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

        {(region !== "All" || features.length > 0) && (
          <div className="bg-muted/30 px-5 flex items-center justify-between gap-3 min-h-12">
            <p className="text-xs text-muted-foreground">
              Showing{" "}
              <span className="font-semibold text-foreground">
                {filtered.length}
              </span>{" "}
              of {providers.length} providers
            </p>
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
          </div>
        )}
      </section>

      {/* Provider table (desktop) / card list (mobile) */}
      <section>
        <div className="hidden md:block rounded-xl border border-border bg-card overflow-x-auto">
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

        {/* Mobile: stacked cards instead of a sideways-scrolling table */}
        <div className="md:hidden flex flex-col gap-3">
          {filtered.length === 0 && (
            <div className="rounded-xl border border-border bg-card text-center text-sm text-muted-foreground py-10">
              No providers match these filters.
            </div>
          )}
          {filtered.map((p) => (
            <ProviderCard key={p.id} provider={p} />
          ))}
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

function openProvider(provider: Provider) {
  sendEvent("debit_card_url_clicked", {
    name: provider.name,
    url: provider.url,
  });
  window.open(provider.url, "_blank", "noopener,noreferrer");
}

function ProviderRow({ provider }: { provider: Provider }) {
  return (
    <TableRow
      className="[&_td]:py-3 cursor-pointer hover:bg-accent/30 transition-colors"
      onClick={() => openProvider(provider)}
    >
      <TableCell>
        <div className="flex items-center gap-3 group">
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
        </div>
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
    </TableRow>
  );
}

function ProviderCard({ provider }: { provider: Provider }) {
  return (
    <div className="rounded-xl border border-border bg-card p-4">
      <div className="flex items-center gap-3">
        <Avatar className="size-10 rounded-lg shrink-0">
          <AvatarImage
            src={provider.logo}
            alt={provider.name}
            className="rounded-lg object-contain bg-secondary p-1"
          />
          <AvatarFallback className="rounded-lg bg-secondary text-secondary-foreground text-xs font-semibold">
            {provider.initials}
          </AvatarFallback>
        </Avatar>
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-1.5">
            <span className="font-medium truncate">{provider.name}</span>
            {provider.selfCustody && (
              <ShieldCheckIcon className="size-3 shrink-0 text-positive-foreground" />
            )}
            {provider.lightningNative && (
              <ZapIcon className="size-3 shrink-0 text-primary" />
            )}
          </div>
          <p className="text-xs text-muted-foreground">
            {provider.network} · {provider.cardType}
          </p>
        </div>
      </div>

      <div className="mt-3 flex flex-wrap items-center gap-1.5">
        {provider.regions.map((r) => (
          <Badge
            key={r}
            variant="secondary"
            className="text-[10px] font-medium px-1.5 py-0"
          >
            {r}
          </Badge>
        ))}
        {(provider.applePay || provider.googlePay) && (
          <span className="flex items-center gap-1.5 text-muted-foreground ml-auto">
            {provider.applePay && <AppleIcon />}
            {provider.googlePay && <GooglePayIcon />}
          </span>
        )}
      </div>

      <div className="mt-3 grid grid-cols-2 gap-x-4 gap-y-2">
        <CardFact label="KYC">
          <KycBadge kyc={provider.kyc} />
        </CardFact>
        <CardFact label="Time to get">
          <span className="inline-flex items-center gap-1.5 text-xs text-muted-foreground">
            <ClockIcon className="size-3" />
            {provider.timeToGet}
          </span>
        </CardFact>
        <CardFact label="Card cost">
          <span className="text-xs text-muted-foreground">
            {provider.cardCost}
          </span>
        </CardFact>
        <CardFact label="Fees">
          <span className="text-xs text-muted-foreground">{provider.fees}</span>
        </CardFact>
      </div>

      <Button
        variant="outline"
        size="sm"
        className="mt-4 w-full"
        onClick={() => openProvider(provider)}
      >
        Visit
        <ArrowUpRightIcon className="size-4" />
      </Button>
    </div>
  );
}

function CardFact({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div className="flex flex-col gap-1">
      <p className="text-[10px] uppercase tracking-wider text-muted-foreground">
        {label}
      </p>
      {children}
    </div>
  );
}

function ConnectCardDialog({
  open,
  onOpenChange,
  providers,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  providers: Provider[];
}) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Pick your card provider</DialogTitle>
          <DialogDescription>
            We'll take you to the setup guide for the one you choose.
          </DialogDescription>
        </DialogHeader>

        <div className="grid grid-cols-1 gap-2">
          {providers.map((p) => {
            if (!p.appStoreId) {
              return null;
            }
            return (
              <Link
                key={p.id}
                to={`/apps/new?app=${p.appStoreId}`}
                onClick={() => {
                  sendEvent("debit_card_connect", { name: p.name });
                  onOpenChange(false);
                }}
                className="flex items-center gap-3 rounded-lg border border-border p-3 hover:bg-accent/40 transition-colors"
              >
                <Avatar className="size-10 rounded-lg shrink-0">
                  <AvatarImage
                    src={p.logo}
                    alt={p.name}
                    className="rounded-lg object-contain bg-secondary p-1"
                  />
                  <AvatarFallback className="rounded-lg bg-secondary text-secondary-foreground text-xs font-semibold">
                    {p.initials}
                  </AvatarFallback>
                </Avatar>
                <div className="min-w-0 flex-1">
                  <p className="font-medium truncate">{p.name}</p>
                  <p className="text-xs text-muted-foreground">
                    {p.network} · {p.cardType}
                  </p>
                </div>
                <ArrowUpRightIcon className="size-4 text-muted-foreground" />
              </Link>
            );
          })}

          <Link
            to="/apps/new?app=bitcoin-card-topup"
            onClick={() => {
              sendEvent("debit_card_connect", { name: "Other" });
              onOpenChange(false);
            }}
            className="flex items-center gap-3 rounded-lg border border-dashed border-border p-3 hover:bg-accent/40 transition-colors"
          >
            <div className="flex items-center justify-center size-10 rounded-lg shrink-0 bg-secondary text-secondary-foreground">
              <CreditCardIcon className="size-5" />
            </div>
            <div className="min-w-0 flex-1">
              <p className="font-medium truncate">Other card</p>
              <p className="text-xs text-muted-foreground">
                Any crypto card not listed
              </p>
            </div>
            <ArrowUpRightIcon className="size-4 text-muted-foreground" />
          </Link>
        </div>
      </DialogContent>
    </Dialog>
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
