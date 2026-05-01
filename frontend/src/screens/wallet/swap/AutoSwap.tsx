import {
  ArrowDownUpIcon,
  ClipboardPasteIcon,
  ClockIcon,
  CopyIcon,
  MoveRightIcon,
  XCircleIcon,
} from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import Loading from "src/components/Loading";
import PasswordInput from "src/components/password/PasswordInput";
import ResponsiveLinkButton from "src/components/ResponsiveLinkButton";
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "src/components/ui/alert-dialog";
import { Button } from "src/components/ui/button";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { RadioGroup, RadioGroupItem } from "src/components/ui/radio-group";
import { useAutoSwapsConfig, useSwapInfo } from "src/hooks/useSwaps";
import { copyToClipboard } from "src/lib/clipboard";
import { AutoSwapConfig, AutoSwapRequest } from "src/types";
import { request } from "src/utils/request";

export default function AutoSwap() {
  const { data: swapConfig } = useAutoSwapsConfig();

  if (!swapConfig) {
    return <Loading />;
  }

  return (
    <div className="grid gap-5">
      <AppHeader
        pageTitle="Auto Swap Out"
        title="Auto Swap Out"
        contentRight={
          <ResponsiveLinkButton
            to="/wallet/swap"
            variant="outline"
            icon={ArrowDownUpIcon}
            text="Swap"
          />
        }
      />
      <div className="w-full lg:max-w-lg min-w-0">
        {swapConfig.enabled ? (
          <ActiveSwapOutConfig swapConfig={swapConfig} />
        ) : (
          <AutoSwapOutForm />
        )}
      </div>
    </div>
  );
}

function AutoSwapOutForm() {
  const { mutate } = useAutoSwapsConfig();
  const { data: swapInfo } = useSwapInfo("out");

  const [isInternalSwap, setInternalSwap] = useState(true);
  const [balanceThresholdSat, setBalanceThresholdSat] = useState("");
  const [swapAmountSat, setSwapAmountSat] = useState("");
  const [destination, setDestination] = useState("");
  const [externalType, setExternalType] = useState<"address" | "xpub">(
    "address"
  );
  const [unlockPassword, setUnlockPassword] = useState("");
  const [showUnlockPasswordDialog, setShowUnlockPasswordDialog] =
    useState(false);
  const [loading, setLoading] = useState(false);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (Number(swapAmountSat) > Number(balanceThresholdSat)) {
      toast.info(
        "Balance threshold must be greater than or equal to swap amount"
      );
      return;
    }

    if (externalType === "xpub" && !isInternalSwap) {
      setShowUnlockPasswordDialog(true);
      return;
    }

    await submitAutoSwap();
  };

  const onConfirmUnlockPassword = async (e: React.FormEvent) => {
    e.preventDefault();
    await submitAutoSwap(unlockPassword);
  };

  const submitAutoSwap = async (password?: string) => {
    try {
      setLoading(true);
      const payload: AutoSwapRequest = {
        swapAmountSat: parseInt(swapAmountSat),
        balanceThresholdSat: parseInt(balanceThresholdSat),
        destination,
        destinationType: !isInternalSwap ? externalType : undefined,
        unlockPassword: password,
      };
      await request("/api/autoswap", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
      });
      setUnlockPassword("");
      setShowUnlockPasswordDialog(false);
      toast("Auto swap enabled successfully");
      await mutate();
    } catch (error) {
      toast("Failed to save auto swap settings", {
        description: (error as Error).message,
      });
    } finally {
      setLoading(false);
    }
  };

  const paste = async () => {
    const text = await navigator.clipboard.readText();
    setDestination(text.trim());
  };

  if (!swapInfo) {
    return <Loading />;
  }

  return (
    <>
      <form onSubmit={onSubmit} className="flex flex-col gap-6">
        <div>
          <h2 className="font-medium text-foreground flex items-center gap-1">
            Lightning <MoveRightIcon /> On-chain
          </h2>
          <p className="mt-1 text-muted-foreground">
            Setup automatic swap of lightning funds into your on-chain balance
            every time a set threshold is reached.
          </p>
          <p className="mt-2 text-muted-foreground flex gap-2 items-center text-sm">
            <ClockIcon className="w-4 h-4" />
            Swaps will be made once per hour
          </p>
        </div>

        <div className="grid gap-1.5">
          <Label>Spending balance threshold</Label>
          <Input
            type="number"
            placeholder="Amount in satoshis"
            value={balanceThresholdSat}
            min={swapAmountSat}
            onChange={(e) => setBalanceThresholdSat(e.target.value)}
            required
          />
          <p className="text-xs text-muted-foreground">
            Swap out as soon as this amount is reached
          </p>
        </div>

        <div className="grid gap-1.5">
          <Label>Swap amount</Label>
          <Input
            type="number"
            placeholder="Amount in satoshis"
            value={swapAmountSat}
            min={swapInfo.minAmountSat}
            max={swapInfo.maxAmountSat}
            onChange={(e) => setSwapAmountSat(e.target.value)}
            required
          />
          <p className="text-xs text-muted-foreground">
            Minimum{" "}
            <FormattedBitcoinAmount amountMsat={swapInfo.minAmountSat * 1000} />
          </p>
        </div>
        <div className="flex flex-col gap-4">
          <Label>Swap to</Label>
          <RadioGroup
            value={isInternalSwap ? "internal" : "external"}
            onValueChange={() => {
              setDestination("");
              setInternalSwap(!isInternalSwap);
            }}
            className="flex gap-4 flex-row"
          >
            <div className="flex items-start space-x-2 mb-2">
              <RadioGroupItem
                value="internal"
                id="internal"
                className="shrink-0"
              />
              <Label htmlFor="internal" className="cursor-pointer">
                On-chain balance
              </Label>
            </div>
            <div className="flex items-start space-x-2">
              <RadioGroupItem
                value="external"
                id="external"
                className="shrink-0"
              />
              <Label htmlFor="external" className="cursor-pointer">
                External on-chain wallet
              </Label>
            </div>
          </RadioGroup>
        </div>
        {!isInternalSwap && (
          <div className="grid gap-4">
            <div className="flex flex-col gap-3">
              <Label>Destination Type</Label>
              <RadioGroup
                value={externalType}
                onValueChange={(value) => {
                  setExternalType(value as "address" | "xpub");
                  setDestination("");
                }}
                className="flex gap-4 flex-row"
              >
                <div className="flex items-start space-x-2">
                  <RadioGroupItem
                    value="address"
                    id="address"
                    className="shrink-0"
                  />
                  <div className="grid gap-1.5">
                    <Label htmlFor="address" className="cursor-pointer">
                      Single Address
                    </Label>
                    <p className="text-xs text-muted-foreground">
                      Send to the same address each time
                    </p>
                  </div>
                </div>
                <div className="flex items-start space-x-2">
                  <RadioGroupItem value="xpub" id="xpub" className="shrink-0" />
                  <div className="grid gap-1.5">
                    <Label htmlFor="xpub" className="cursor-pointer">
                      XPUB
                    </Label>
                    <p className="text-xs text-muted-foreground">
                      Generate new addresses from extended public key
                    </p>
                  </div>
                </div>
              </RadioGroup>
            </div>
            <div className="grid gap-1.5">
              <Label>
                {externalType === "address"
                  ? "Receiving on-chain address"
                  : "Extended Public Key (XPUB)"}
              </Label>
              <div className="flex gap-2">
                <Input
                  placeholder={
                    externalType === "address" ? "bc1..." : "xpub..."
                  }
                  value={destination}
                  onChange={(e) => setDestination(e.target.value)}
                  required
                />
                <Button
                  type="button"
                  variant="outline"
                  className="px-2"
                  onClick={paste}
                >
                  <ClipboardPasteIcon className="w-4 h-4" />
                </Button>
              </div>
              <p className="text-xs text-muted-foreground">
                {externalType === "address"
                  ? "Enter a Bitcoin address to receive swapped funds"
                  : "Enter an XPUB to automatically generate new addresses for each swap. You will be asked to enter your unlock password to encrypt it safely"}
              </p>
            </div>
          </div>
        )}

        <div className="flex items-center justify-between border-t pt-4">
          <Label>Fee</Label>
          <p className="text-muted-foreground text-sm">
            {swapInfo.albyServiceFee + swapInfo.boltzServiceFee}% + on-chain
            fees
          </p>
        </div>
        <div className="grid gap-1">
          <LoadingButton className="w-full" loading={loading}>
            Begin Auto Swap
          </LoadingButton>
          <p className="text-xs text-muted-foreground text-right">
            powered by{" "}
            <ExternalLink
              to="https://boltz.exchange"
              className="font-medium text-foreground"
            >
              boltz.exchange
            </ExternalLink>
          </p>
        </div>
      </form>
      <AlertDialog
        open={showUnlockPasswordDialog}
        onOpenChange={(open) => {
          setShowUnlockPasswordDialog(open);
          if (!open) {
            setUnlockPassword("");
          }
        }}
      >
        <AlertDialogContent>
          <form onSubmit={onConfirmUnlockPassword}>
            <AlertDialogHeader>
              <AlertDialogTitle>Confirm Auto Swap Setup</AlertDialogTitle>
              <AlertDialogDescription>
                <div className="flex flex-col gap-4">
                  <p>
                    Please enter your unlock password to encrypt and securely
                    store the XPUB.
                  </p>
                  <div className="grid gap-1.5">
                    <Label htmlFor="unlockPassword">Unlock Password</Label>
                    <PasswordInput
                      id="unlockPassword"
                      onChange={setUnlockPassword}
                      autoFocus
                      value={unlockPassword}
                    />
                  </div>
                </div>
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter className="mt-3">
              <AlertDialogCancel>Cancel</AlertDialogCancel>
              <Button type="submit" disabled={!unlockPassword || loading}>
                Confirm
              </Button>
            </AlertDialogFooter>
          </form>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}

function ActiveSwapOutConfig({ swapConfig }: { swapConfig: AutoSwapConfig }) {
  const { mutate } = useAutoSwapsConfig();
  const { data: swapInfo } = useSwapInfo("out");

  const [loading, setLoading] = useState(false);

  const onDeactivate = async () => {
    try {
      setLoading(true);
      await request(`/api/autoswap`, {
        method: "DELETE",
        headers: {
          "Content-Type": "application/json",
        },
      });
      toast("Deactivated auto swap successfully");
      await mutate();
    } catch (error) {
      toast.error("Deactivating auto swaps failed", {
        description: (error as Error).message,
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <h2 className="font-medium text-foreground flex items-center gap-1">
        Active Lightning <MoveRightIcon /> On-chain Swap
      </h2>
      <p className="mt-1 text-muted-foreground">
        Alby Hub will try to perform a swap every time the balance reaches the
        threshold.
      </p>
      <p className="mt-2 text-muted-foreground flex gap-2 items-center text-sm">
        <ClockIcon className="w-4 h-4" />
        Swaps will be made once per hour
      </p>

      <div className="my-6 space-y-4 text-sm">
        <div className="flex justify-between items-center gap-2">
          <span className="font-medium">Type</span>
          <span className="truncate text-muted-foreground text-right">
            Lightning to On-chain
          </span>
        </div>
        <div className="flex justify-between items-center gap-2">
          <div className="font-medium">Destination</div>
          <div className="flex min-w-0 items-center justify-end gap-2 text-muted-foreground">
            <div className="truncate text-right">
              {swapConfig.destination || "On-chain Balance"}
            </div>
            {swapConfig.destination && (
              <CopyIcon
                className="cursor-pointer size-4 shrink-0"
                onClick={() => copyToClipboard(swapConfig.destination)}
              />
            )}
          </div>
        </div>
        <div className="flex justify-between items-center gap-2">
          <span className="font-medium truncate">
            Spending Balance Threshold
          </span>
          <span className="shrink-0 text-muted-foreground text-right">
            <FormattedBitcoinAmount
              amountMsat={swapConfig.balanceThresholdSat * 1000}
            />
          </span>
        </div>
        <div className="flex justify-between items-center gap-2">
          <span className="font-medium truncate">Swap amount</span>
          <span className="shrink-0 text-muted-foreground text-right">
            <FormattedBitcoinAmount
              amountMsat={swapConfig.swapAmountSat * 1000}
            />
          </span>
        </div>
        <div className="flex justify-between items-center gap-2">
          <span className="font-medium">Fee</span>
          {swapInfo ? (
            <span className="truncate text-muted-foreground text-right">
              {swapInfo.albyServiceFee + swapInfo.boltzServiceFee}% + on-chain
              fees
            </span>
          ) : (
            <Loading className="w-4 h-4" />
          )}
        </div>
      </div>
      <Button onClick={onDeactivate} disabled={loading} variant="outline">
        <XCircleIcon />
        Deactivate Auto Swap
      </Button>
    </>
  );
}
