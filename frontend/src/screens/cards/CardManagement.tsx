import {
  AlertTriangleIcon,
  ArrowRightIcon,
  CheckIcon,
  CopyIcon,
  ExternalLinkIcon,
  LinkIcon,
} from "lucide-react";
import React from "react";
import { Link } from "react-router";
import { toast } from "sonner";
import { MastercardLogo } from "src/components/icons/MastercardLogo";
import { VisaLogo } from "src/components/icons/VisaLogo";
import Loading from "src/components/Loading";
import QRCode from "src/components/QRCode";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { Avatar, AvatarFallback, AvatarImage } from "src/components/ui/avatar";
import { Button } from "src/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "src/components/ui/dialog";
import { Field, FieldGroup, FieldLabel } from "src/components/ui/field";
import { Input } from "src/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "src/components/ui/select";
import { cn } from "src/lib/utils";
import type {
  Provider,
  SupportedCurrency,
  UserCard,
} from "src/screens/cards/useUserCards";

export type { Provider, UserCard } from "src/screens/cards/useUserCards";

const TOPUP_BASE_URL = "https://card.albylabs.com";

// Chains supported by bitcoin-card-topup (LendaSwap).
const SUPPORTED_CHAINS: { id: number; name: string }[] = [
  { id: 42161, name: "Arbitrum One" },
  { id: 137, name: "Polygon" },
  { id: 1, name: "Ethereum" },
];

const SUPPORTED_CURRENCIES: SupportedCurrency[] = ["USDC", "USDT"];

// Brand styling per provider (used only on user card tiles)
const providerStyle: Record<string, { bg: string; text: string }> = {
  redotpay: { bg: "bg-rose-600", text: "text-white" },
  "2fiat": { bg: "bg-violet-600", text: "text-white" },
  freedomia: { bg: "bg-orange-500", text: "text-white" },
};

export function YourCardsSection({
  cards,
  providers,
}: {
  cards: UserCard[];
  providers: Provider[];
}) {
  const providerById = React.useMemo(
    () => Object.fromEntries(providers.map((p) => [p.id, p])),
    [providers]
  );

  if (cards.length === 0) {
    return null;
  }

  return (
    <section>
      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
        {cards.map((card) => (
          <UserCardTile
            key={card.appId}
            card={card}
            provider={providerById[card.providerId]}
          />
        ))}
      </div>
    </section>
  );
}

function NetworkLogo({
  network,
  className,
}: {
  network: "Visa" | "Mastercard";
  className?: string;
}) {
  if (network === "Visa") {
    return <VisaLogo className={cn("h-5 w-auto", className)} />;
  }
  return <MastercardLogo className={cn("h-7 w-auto", className)} />;
}

function UserCardTile({
  card,
  provider,
}: {
  card: UserCard;
  provider?: Provider;
}) {
  const style = providerStyle[card.providerId] ?? {
    bg: "bg-primary",
    text: "text-primary-foreground",
  };

  const detailPath = `/apps/${card.appId}`;

  return (
    <Link
      to={detailPath}
      className={cn(
        "relative w-full text-left rounded-xl overflow-hidden p-5 aspect-[1.85/1]",
        "flex flex-col justify-between transition-transform hover:-translate-y-0.5 hover:shadow-lg",
        "focus:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
        style.bg,
        style.text
      )}
    >
      <div className="absolute inset-0 bg-gradient-to-br from-white/10 via-transparent to-black/15 pointer-events-none" />

      <div className="relative flex items-center gap-3 min-w-0">
        <Avatar className="size-12 rounded-xl shrink-0 shadow-md">
          <AvatarImage
            src={provider?.logo}
            alt={provider?.name}
            className="rounded-xl object-contain bg-white p-2"
          />
          <AvatarFallback className="rounded-xl bg-white text-foreground text-sm font-semibold">
            {provider?.initials ?? "??"}
          </AvatarFallback>
        </Avatar>
        <div className="min-w-0">
          <p className="text-base font-bold tracking-tight truncate">
            {provider?.name ?? "Card"}
          </p>
          <p className="text-[10px] uppercase tracking-widest opacity-70 truncate">
            Debit · {provider?.network ?? "Card"}
          </p>
        </div>
      </div>

      <div className="relative flex items-end justify-end gap-3">
        {provider && <NetworkLogo network={provider.network} />}
      </div>
    </Link>
  );
}

export type ConnectCardInput = {
  providerId: string;
  providerName: string;
  destinationAddress: string;
  chainId: number;
  currency: SupportedCurrency;
};

export function ConnectCardDialog({
  open,
  onOpenChange,
  providers,
  onSubmit,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  providers: Provider[];
  onSubmit: (input: ConnectCardInput) => Promise<void>;
}) {
  const [providerId, setProviderId] = React.useState<string>(
    providers[0]?.id ?? ""
  );
  const [destinationAddress, setDestinationAddress] = React.useState("");
  const [chainId, setChainId] = React.useState<number>(SUPPORTED_CHAINS[0].id);
  const [currency, setCurrency] = React.useState<SupportedCurrency>(
    SUPPORTED_CURRENCIES[0]
  );
  const [submitting, setSubmitting] = React.useState(false);

  React.useEffect(() => {
    if (!open) {
      setDestinationAddress("");
      setChainId(SUPPORTED_CHAINS[0].id);
      setCurrency(SUPPORTED_CURRENCIES[0]);
      setProviderId(providers[0]?.id ?? "");
      setSubmitting(false);
    }
  }, [open, providers]);

  const selectedProvider = providers.find((p) => p.id === providerId);
  const providerName = selectedProvider?.name;
  const appStoreId = selectedProvider?.appStoreId;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!providerId || !destinationAddress || submitting || appStoreId) {
      return;
    }
    setSubmitting(true);
    try {
      await onSubmit({
        providerId,
        providerName: providerName ?? "Card",
        destinationAddress: destinationAddress.trim(),
        chainId,
        currency,
      });
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Connect your card</DialogTitle>
          <DialogDescription>
            Save your card's deposit details once. We'll hand you a top-up link
            you can bookmark.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit}>
          <FieldGroup>
            <Field>
              <FieldLabel htmlFor="provider">Provider</FieldLabel>
              <Select value={providerId} onValueChange={setProviderId}>
                <SelectTrigger id="provider">
                  <SelectValue placeholder="Select a provider" />
                </SelectTrigger>
                <SelectContent>
                  {providers.map((p) => (
                    <SelectItem key={p.id} value={p.id}>
                      {p.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </Field>

            {!appStoreId && (
              <>
                <Field>
                  <FieldLabel htmlFor="address">Top-up address</FieldLabel>
                  <Input
                    id="address"
                    placeholder="0x…"
                    className="font-mono"
                    value={destinationAddress}
                    onChange={(e) => setDestinationAddress(e.target.value)}
                    autoCapitalize="off"
                    autoCorrect="off"
                    spellCheck={false}
                    required
                  />
                </Field>

                <div className="grid grid-cols-2 gap-3">
                  <Field>
                    <FieldLabel htmlFor="network">Network</FieldLabel>
                    <Select
                      value={String(chainId)}
                      onValueChange={(v) => setChainId(Number(v))}
                    >
                      <SelectTrigger id="network">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {SUPPORTED_CHAINS.map((c) => (
                          <SelectItem key={c.id} value={String(c.id)}>
                            {c.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </Field>

                  <Field>
                    <FieldLabel htmlFor="currency">Currency</FieldLabel>
                    <Select
                      value={currency}
                      onValueChange={(v) => setCurrency(v as SupportedCurrency)}
                    >
                      <SelectTrigger id="currency">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {SUPPORTED_CURRENCIES.map((c) => (
                          <SelectItem key={c} value={c}>
                            {c}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </Field>
                </div>
              </>
            )}

            {appStoreId && (
              <p className="text-sm text-muted-foreground">
                {providerName} pairs through its own app. Continue to the setup
                guide to get a Nostr Wallet Connect link and finish the
                connection from there.
              </p>
            )}
          </FieldGroup>

          <DialogFooter className="mt-6">
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={submitting}
            >
              Cancel
            </Button>
            {appStoreId ? (
              <Button asChild>
                <Link
                  to={`/apps/new?app=${appStoreId}`}
                  onClick={() => onOpenChange(false)}
                >
                  Open setup guide
                  <ArrowRightIcon className="size-4" />
                </Link>
              </Button>
            ) : (
              <Button type="submit" disabled={submitting}>
                {submitting ? (
                  <Loading className="size-4" />
                ) : (
                  <LinkIcon className="size-4" />
                )}
                {submitting ? "Connecting…" : "Connect card"}
              </Button>
            )}
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

function buildTopupUrl({
  card,
  provider,
  pairingUri,
}: {
  card: UserCard;
  provider?: Provider;
  pairingUri: string;
}) {
  const params = new URLSearchParams();
  params.set("label", provider?.name ?? "Card");
  params.set("address", card.destinationAddress);
  params.set("chainId", String(card.chainId));
  params.set("currency", card.currency);
  params.set("nwc", pairingUri);
  // Use `#?` so bitcoin-connect's parser (which does indexOf("?") on the
  // hash) takes the expected branch.
  return `${TOPUP_BASE_URL}/#?${params.toString()}`;
}

export function CardCreatedDialog({
  card,
  provider,
  pairingUri,
  onOpenChange,
}: {
  card: UserCard | null;
  provider?: Provider;
  pairingUri: string | null;
  onOpenChange: (open: boolean) => void;
}) {
  const [copied, setCopied] = React.useState(false);

  React.useEffect(() => {
    if (card) {
      setCopied(false);
    }
  }, [card]);

  if (!card || !pairingUri) {
    return null;
  }

  const topupUrl = buildTopupUrl({ card, provider, pairingUri });

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(topupUrl);
      setCopied(true);
      toast.success("Top-up link copied");
      setTimeout(() => setCopied(false), 2000);
    } catch {
      toast.error("Could not copy");
    }
  };

  return (
    <Dialog open={!!card} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{provider?.name ?? "Card"} connected</DialogTitle>
          <DialogDescription>
            This is your top-up link. Save it now — you won't see it again.
          </DialogDescription>
        </DialogHeader>

        <Alert variant="warning">
          <AlertTriangleIcon />
          <AlertTitle>Save this link before you close this dialog</AlertTitle>
          <AlertDescription>
            <p>
              <span className="font-medium">Bookmark it</span> or{" "}
              <span className="font-medium">add it to your home screen</span> —
              the secret in this link can't be recovered later. Lose the link
              and you'll need to connect a new card.
            </p>
          </AlertDescription>
        </Alert>

        <div className="flex flex-col items-center gap-3">
          <div className="hidden sm:block rounded-xl border border-border p-3 bg-white">
            <QRCode value={topupUrl} size={200} />
          </div>
          <p className="hidden sm:block text-sm text-muted-foreground text-center max-w-xs">
            Scan with your phone's{" "}
            <span className="font-medium">camera app</span> to open it, then
            save it as an app to your homescreen.
          </p>

          <div className="w-full">
            <FieldLabel className="text-xs text-muted-foreground mb-1.5">
              Top-up link
            </FieldLabel>
            <div className="flex items-center gap-2">
              <Input
                value={topupUrl}
                readOnly
                className="font-mono text-xs"
                onFocus={(e) => e.currentTarget.select()}
              />
              <Button
                size="icon"
                variant="outline"
                onClick={handleCopy}
                aria-label="Copy top-up link"
              >
                {copied ? (
                  <CheckIcon className="size-4" />
                ) : (
                  <CopyIcon className="size-4" />
                )}
              </Button>
            </div>
          </div>
        </div>

        <DialogFooter className="mt-2 sm:justify-between gap-2">
          <Button variant="outline" asChild>
            <a href={topupUrl} target="_blank" rel="noopener noreferrer">
              <ExternalLinkIcon className="size-4" />
              Open top-up page
            </a>
          </Button>
          <Button onClick={() => onOpenChange(false)}>I've saved it</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
