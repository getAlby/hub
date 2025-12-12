import {
  CircleAlertIcon,
  CircleCheckIcon,
  CircleHelpIcon,
  CircleXIcon,
  CopyIcon,
  ExternalLinkIcon,
} from "lucide-react";
import React, { useEffect, useState } from "react";
import { useParams, useSearchParams } from "react-router-dom";
import { toast } from "sonner";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import Loading from "src/components/Loading";
import LottieLoading from "src/components/LottieLoading";
import QRCode from "src/components/QRCode";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { ExternalLinkButton } from "src/components/ui/custom/external-link-button";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "src/components/ui/tooltip";
import { useInfo } from "src/hooks/useInfo";
import { useMempoolApi } from "src/hooks/useMempoolApi";
import { useSwap } from "src/hooks/useSwaps";
import { useSyncWallet } from "src/hooks/useSyncWallet";
import { copyToClipboard } from "src/lib/clipboard";
import { RedeemOnchainFundsResponse, SwapIn } from "src/types";
import { request } from "src/utils/request";

export default function SwapInStatus() {
  const { data: info } = useInfo();
  const { data: recommendedFees } = useMempoolApi<{
    fastestFee: number;
    halfHourFee: number;
    economyFee: number;
    minimumFee: number;
  }>("/v1/fees/recommended");
  useSyncWallet(); // ensure funds show up on node page after swap completes
  const { swapId } = useParams() as { swapId: string };
  const { data: swap } = useSwap<SwapIn>(swapId, true);

  const [feeRate, setFeeRate] = useState("");
  const [isPaying, setPaying] = useState(false);
  const [searchParams] = useSearchParams();

  const isInternalSwap = searchParams.has("internal", "true");
  const [, setPaidWithAlbyHub] = React.useState(false);

  useEffect(() => {
    if (recommendedFees?.fastestFee) {
      setFeeRate(recommendedFees.fastestFee.toString());
    }
  }, [recommendedFees]);

  useEffect(() => {
    if (isPaying && swap?.lockupTxId) {
      setPaying(false);
    }
  }, [isPaying, swap?.lockupTxId]);

  const payWithAlbyHub = React.useCallback(() => {
    (async () => {
      setPaying(true);
      try {
        if (!swap) {
          throw new Error("swap not loaded");
        }
        if (!feeRate) {
          throw new Error("No fee rate set");
        }
        const response = await request<RedeemOnchainFundsResponse>(
          "/api/wallet/redeem-onchain-funds",
          {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
            },
            body: JSON.stringify({
              toAddress: swap.lockupAddress,
              amount: swap.sendAmount,
              feeRate: +feeRate,
            }),
          }
        );
        if (!response?.txId) {
          throw new Error("No address in response");
        }
        console.info("Redeemed onchain funds", response);
      } catch (error) {
        console.error(error);
        toast.error("Failed to redeem onchain funds", {
          description: "" + error,
        });
        setPaying(false);
      }
    })();
  }, [feeRate, swap]);

  React.useEffect(() => {
    if (isInternalSwap && feeRate && swap) {
      setPaidWithAlbyHub((current) => {
        if (current) {
          return current;
        }
        setTimeout(() => {
          payWithAlbyHub();
        }, 1);
        return true;
      });
    }
  }, [feeRate, isInternalSwap, payWithAlbyHub, swap]);

  if (!swap) {
    return <Loading />;
  }

  const copyPaymentHash = () => {
    copyToClipboard(swap.paymentHash);
  };

  const copyAddress = () => {
    copyToClipboard(swap.lockupAddress);
  };

  const copyAmount = () => {
    copyToClipboard(swap.sendAmount.toString());
  };

  const swapStatus = swap.state;
  const statusText = {
    SUCCESS: "Swap Successful",
    FAILED: "Swap Failed",
    REFUNDED: "Swap Refunded",
    PENDING: swap.lockupTxId
      ? "Waiting for confirmation"
      : isInternalSwap
        ? "Depositing on-chain funds"
        : "Waiting for deposit",
  };

  return (
    <div className="grid gap-5">
      <AppHeader title="Swap In" />
      <div className="w-full max-w-lg">
        <Card className="w-full md:max-w-xs">
          <CardHeader>
            <CardTitle className="flex justify-center">
              {swapStatus === "PENDING" && <Loading className="w-4 h-4 mr-2" />}
              {statusText[swapStatus]}
            </CardTitle>
            <CardDescription className="flex items-center justify-center gap-2 text-muted-foreground text-sm">
              Swap ID: {swap.id}{" "}
              <CopyIcon
                className="cursor-pointer text-muted-foreground size-4"
                onClick={() => {
                  copyToClipboard(swap.id);
                }}
              />
            </CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col items-center gap-4">
            {swapStatus === "SUCCESS" ? (
              <>
                <CircleCheckIcon className="w-60 h-60" />
                <div className="flex flex-col gap-2 items-center">
                  <p className="text-xl font-bold slashed-zero text-center">
                    <FormattedBitcoinAmount
                      amount={(swap.receiveAmount as number) * 1000}
                    />
                  </p>
                  <FormattedFiatAmount amount={swap.receiveAmount as number} />
                </div>
                <Button onClick={copyPaymentHash} variant="outline">
                  <CopyIcon />
                  Copy Payment Hash
                </Button>
              </>
            ) : (
              <>
                {(swapStatus === "REFUNDED" || swapStatus === "FAILED") && (
                  <CircleXIcon className="w-60 h-60" />
                )}
                {swapStatus === "PENDING" &&
                  (swap.lockupTxId || isInternalSwap ? (
                    <LottieLoading />
                  ) : (
                    <QRCode
                      value={`bitcoin:${swap.lockupAddress}?amount=${swap.sendAmount / 100_000_000}`}
                    />
                  ))}
                <div className="flex flex-col gap-2 items-center">
                  <div className="flex items-center gap-2">
                    <p className="text-xl font-bold slashed-zero text-center">
                      <FormattedBitcoinAmount amount={swap.sendAmount * 1000} />
                    </p>
                    {!swap.lockupTxId && !isInternalSwap && (
                      <CopyIcon
                        className="cursor-pointer text-muted-foreground size-4 shrink-0"
                        onClick={copyAmount}
                      />
                    )}
                  </div>
                  <FormattedFiatAmount amount={swap.sendAmount} />
                </div>
                {!swap.lockupTxId && !isInternalSwap && (
                  <div className="flex justify-center gap-4 flex-wrap">
                    {swap.state !== "FAILED" && (
                      <Button onClick={copyAddress} variant="outline">
                        <CopyIcon />
                        Copy Address
                      </Button>
                    )}
                    {swap.state === "PENDING" && (
                      <ExternalLinkButton
                        to={`bitcoin:${swap.lockupAddress}?amount=${swap.sendAmount / 100_000_000}`}
                        variant="secondary"
                      >
                        Open in External Wallet
                        <ExternalLinkIcon />
                      </ExternalLinkButton>
                    )}
                  </div>
                )}
              </>
            )}
            {/* We only show status screen once bitcoin is locked up */}
            {swap.lockupTxId ? (
              <div className="flex flex-col justify-start gap-2 w-full mt-2">
                {swapStatus === "SUCCESS" && (
                  <>
                    <div className="flex items-center text-muted-foreground text-sm">
                      <CircleCheckIcon className="w-5 h-5 mr-2 text-green-600 dark:text-emerald-500" />
                      Funds received via lightning
                    </div>
                    <Divider color="border-green-600 dark:border-emerald-500" />
                    <div className="flex items-center text-muted-foreground text-sm">
                      <CircleCheckIcon className="w-5 h-5 mr-2 text-green-600 dark:text-emerald-500" />
                      <div className="flex items-center gap-2">
                        <p>Onchain deposit confirmed</p>
                        <ExternalLink
                          to={`${info?.mempoolUrl}/tx/${swap.lockupTxId}`}
                          className="flex items-center underline text-foreground"
                        >
                          View
                        </ExternalLink>
                      </div>
                    </div>
                    <Divider color="border-green-600 dark:border-emerald-500" />
                  </>
                )}
                {swapStatus === "PENDING" && (
                  <>
                    <div className="flex items-center text-muted-foreground text-sm">
                      <Loading className="w-5 h-5 mr-2" />
                      <div className="flex items-center gap-2">
                        <p>Waiting for 1 on-chain confirmation...</p>
                        <ExternalLink
                          to={`${info?.mempoolUrl}/tx/${swap.lockupTxId}`}
                          className="flex items-center underline text-foreground"
                        >
                          View
                        </ExternalLink>
                      </div>
                    </div>
                    <Divider color="border-green-600 dark:border-emerald-500" />
                  </>
                )}
                {swapStatus === "REFUNDED" && (
                  <>
                    <div className="flex items-center text-muted-foreground text-sm">
                      <CircleCheckIcon className="w-5 h-5 mr-2 text-green-600 dark:text-emerald-500" />
                      <div className="flex items-center gap-2">
                        <p>Refund initiated</p>
                        <ExternalLink
                          to={`${info?.mempoolUrl}/tx/${swap.claimTxId}`}
                          className="flex items-center underline text-foreground"
                        >
                          View
                        </ExternalLink>
                      </div>
                    </div>
                    <Divider color="border-green-600 dark:border-emerald-500" />
                  </>
                )}
                {(swapStatus === "FAILED" || swapStatus === "REFUNDED") && (
                  <>
                    <div className="flex items-center text-muted-foreground text-sm">
                      <CircleAlertIcon className="w-5 h-5 mr-2 text-red-500" />
                      <TooltipProvider>
                        <Tooltip>
                          <TooltipTrigger>
                            <div className="flex items-center gap-2">
                              <p>Onchain deposit failed</p>
                              <ExternalLink
                                to={`${info?.mempoolUrl}/tx/${swap.lockupTxId}`}
                                className="flex items-center underline text-foreground"
                              >
                                View
                              </ExternalLink>
                              <CircleHelpIcon className="h-4 w-4 text-muted-foreground" />
                            </div>
                          </TooltipTrigger>
                          <TooltipContent>
                            Deposit usually fails when there is an amount
                            mismatch or if Boltz failed to send the lightning
                            payment to your node.
                            {swapStatus !== "REFUNDED" &&
                              " You can use the Swap Refund button in Settings -> Debug Tools to claim the locked up bitcoin."}
                          </TooltipContent>
                        </Tooltip>
                      </TooltipProvider>
                    </div>
                    <Divider color="border-red-500" />
                  </>
                )}
                <div className="flex items-center text-muted-foreground text-sm">
                  <CircleCheckIcon className="w-5 h-5 mr-2 text-green-600 dark:text-emerald-500" />
                  Swap initiated
                </div>
              </div>
            ) : swapStatus === "FAILED" ? (
              <>
                <div className="flex items-center text-muted-foreground text-sm">
                  <CircleAlertIcon className="w-5 h-5 mr-2 text-red-500" />
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger>
                        <div className="flex items-center gap-2">
                          <p>Onchain deposit failed</p>
                          <ExternalLink
                            to={`${info?.mempoolUrl}/address/${swap.lockupAddress}`}
                            className="flex items-center underline text-foreground"
                          >
                            View
                          </ExternalLink>
                          <CircleHelpIcon className="h-4 w-4 text-muted-foreground" />
                        </div>
                      </TooltipTrigger>
                      <TooltipContent>
                        Deposit usually fails when there is an amount mismatch
                        or if Boltz failed to send the lightning payment to your
                        node. You can use the Swap Refund button in Settings{" "}
                        {"->"} Debug Tools to claim the locked up bitcoin.
                      </TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                </div>
              </>
            ) : null}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

const Divider = ({ color }: { color: string }) => (
  <div className={`ml-[9px] py-1 border-l ${color}`}></div>
);
