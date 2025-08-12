import {
  CircleAlertIcon,
  CircleCheckIcon,
  CircleHelpIcon,
  CircleXIcon,
  CopyIcon,
  ZapIcon,
} from "lucide-react";
import { useEffect, useState } from "react";
import Lottie from "react-lottie";
import { useParams } from "react-router-dom";
import animationDataDark from "src/assets/lotties/loading-dark.json";
import animationDataLight from "src/assets/lotties/loading-light.json";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import Loading from "src/components/Loading";
import QRCode from "src/components/QRCode";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { useTheme } from "src/components/ui/theme-provider";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "src/components/ui/tooltip";
import { useToast } from "src/components/ui/use-toast";
import { useBalances } from "src/hooks/useBalances";
import { useInfo } from "src/hooks/useInfo";
import { useMempoolApi } from "src/hooks/useMempoolApi";
import { useSwap } from "src/hooks/useSwaps";
import { useSyncWallet } from "src/hooks/useSyncWallet";
import { copyToClipboard } from "src/lib/clipboard";
import { RedeemOnchainFundsResponse, SwapIn } from "src/types";
import { request } from "src/utils/request";

export default function SwapInStatus() {
  const { toast } = useToast();
  const { isDarkMode } = useTheme();
  const { data: info } = useInfo();
  const { data: balances } = useBalances();
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

  if (!swap) {
    return <Loading />;
  }

  const copyPaymentHash = () => {
    copyToClipboard(swap.paymentHash, toast);
  };

  const copyAddress = () => {
    copyToClipboard(swap.lockupAddress, toast);
  };

  const defaultOptions = {
    loop: true,
    autoplay: true,
    animationData: isDarkMode ? animationDataDark : animationDataLight,
    rendererSettings: {
      preserveAspectRatio: "xMidYMid slice",
    },
  };

  async function payWithAlbyHub() {
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
      toast({
        variant: "destructive",
        title: "Failed to redeem onchain funds",
        description: "" + error,
      });
      setPaying(false);
    }
  }

  const swapStatus = swap.state;
  const statusText = {
    SUCCESS: "Swap Successful",
    FAILED: "Swap Failed",
    REFUNDED: "Swap Refunded",
    PENDING: swap.lockupTxId
      ? "Waiting for confirmation"
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
          </CardHeader>
          <CardContent className="flex flex-col items-center gap-4">
            {swapStatus === "SUCCESS" ? (
              <>
                <CircleCheckIcon className="w-60 h-60" />
                <div className="flex flex-col gap-2 items-center">
                  <p className="text-xl font-bold slashed-zero text-center">
                    {new Intl.NumberFormat().format(
                      swap.receiveAmount as number
                    )}{" "}
                    sats
                  </p>
                  <FormattedFiatAmount amount={swap.receiveAmount as number} />
                </div>
                <Button onClick={copyPaymentHash} variant="outline">
                  <CopyIcon className="w-4 h-4 mr-2" />
                  Copy Payment Hash
                </Button>
              </>
            ) : (
              <>
                {(swapStatus === "REFUNDED" || swapStatus === "FAILED") && (
                  <CircleXIcon className="w-60 h-60" />
                )}
                {swapStatus === "PENDING" &&
                  (swap.lockupTxId ? (
                    <Lottie options={defaultOptions} />
                  ) : (
                    <QRCode value={swap.lockupAddress} />
                  ))}
                <div className="flex flex-col gap-2 items-center">
                  <p className="text-xl font-bold slashed-zero text-center">
                    {new Intl.NumberFormat().format(swap.sendAmount)} sats
                  </p>
                  <FormattedFiatAmount amount={swap.sendAmount} />
                </div>
                {!swap.lockupTxId && (
                  <div className="flex justify-center gap-4 flex-wrap">
                    {swap.state !== "FAILED" && (
                      <Button onClick={copyAddress} variant="outline">
                        <CopyIcon className="w-4 h-4 mr-2" />
                        Copy Address
                      </Button>
                    )}
                    {swap.state === "PENDING" &&
                      balances &&
                      balances.onchain.spendable - 25000 /* anchor reserve */ >
                        swap.sendAmount && (
                        <LoadingButton
                          loading={isPaying}
                          onClick={payWithAlbyHub}
                          disabled={!recommendedFees}
                        >
                          <ZapIcon className="w-4 h-4 mr-2" />
                          Use Hub On-Chain Funds
                        </LoadingButton>
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
                        <p>Waiting for confirmation...</p>
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
                          <TooltipContent className="w-[300px]">
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
                      <TooltipContent className="w-[300px]">
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
