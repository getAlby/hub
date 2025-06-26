import {
  CircleAlertIcon,
  CircleCheckIcon,
  CircleXIcon,
  CopyIcon,
} from "lucide-react";
import Lottie from "react-lottie";
import { useParams } from "react-router-dom";
import animationDataDark from "src/assets/lotties/loading-dark.json";
import animationDataLight from "src/assets/lotties/loading-light.json";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import Loading from "src/components/Loading";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { useTheme } from "src/components/ui/theme-provider";
import { useToast } from "src/components/ui/use-toast";
import { useInfo } from "src/hooks/useInfo";
import { useSwap } from "src/hooks/useSwaps";
import { copyToClipboard } from "src/lib/clipboard";

export default function SwapOutStatus() {
  const { toast } = useToast();
  const { data: info } = useInfo();
  const { isDarkMode } = useTheme();
  const { swapId } = useParams() as { swapId: string };
  const { data: swap } = useSwap(swapId, true);

  if (!swap) {
    return <Loading />;
  }

  const copyTxId = () => {
    copyToClipboard(swap.claimTxId as string, toast);
  };

  const defaultOptions = {
    loop: true,
    autoplay: true,
    animationData: isDarkMode ? animationDataDark : animationDataLight,
    rendererSettings: {
      preserveAspectRatio: "xMidYMid slice",
    },
  };

  const swapStatus = swap.state;
  const statusText = {
    SUCCESS: "Swap Successful",
    FAILED: "Swap Failed",
    REFUNDED: "Swap Refunded",
    PENDING: swap.lockupTxId
      ? "Waiting for lockup confirmation"
      : "Waiting for deposit from boltz",
  };

  return (
    <div className="grid gap-5">
      <AppHeader title="Swap Out" />
      <div className="w-full max-w-lg">
        <Card className="w-full md:max-w-xs">
          <CardHeader>
            <CardTitle className="flex justify-center">
              {swapStatus === "PENDING" && <Loading className="w-4 h-4 mr-2" />}
              {statusText[swapStatus]}
            </CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col items-center gap-4">
            {swapStatus === "SUCCESS" && (
              <>
                <CircleCheckIcon className="w-60 h-60" />
                <div className="flex flex-col gap-2 items-center">
                  <p className="text-xl font-bold slashed-zero text-center">
                    {new Intl.NumberFormat().format(
                      swap.receivedAmount as number
                    )}{" "}
                    sats
                  </p>
                  <FormattedFiatAmount amount={swap.receivedAmount as number} />
                </div>
                <div className="flex justify-center gap-4 flex-wrap">
                  <Button onClick={copyTxId} variant="outline">
                    <CopyIcon className="w-4 h-4 mr-2" />
                    Copy TxId
                  </Button>
                </div>
              </>
            )}
            {swapStatus === "FAILED" && (
              <>
                <CircleXIcon className="w-60 h-60" />
                <div className="flex flex-col gap-2 items-center">
                  <p className="text-xl font-bold slashed-zero text-center">
                    ~{new Intl.NumberFormat().format(swap.sendAmount)} sats
                  </p>
                  <div className="flex items-center">
                    <span className="text-sm text-muted-foreground">~</span>
                    <FormattedFiatAmount amount={swap.sendAmount} />
                  </div>
                </div>
              </>
            )}
            {swapStatus === "PENDING" && (
              <>
                <Lottie options={defaultOptions} />
                <div className="flex flex-col gap-2 items-center">
                  <p className="text-xl font-bold slashed-zero text-center">
                    ~{new Intl.NumberFormat().format(swap.sendAmount)} sats
                  </p>
                  <div className="flex items-center">
                    <span className="text-sm text-muted-foreground">~</span>
                    <FormattedFiatAmount amount={swap.sendAmount} />
                  </div>
                </div>
              </>
            )}
            <div className="flex flex-col justify-start gap-2 w-full mt-2">
              {(swapStatus === "SUCCESS" || swapStatus === "PENDING") && (
                <>
                  {swap.lockupTxId ? (
                    <>
                      {swap.claimTxId ? (
                        <>
                          <div className="flex items-center text-muted-foreground text-sm">
                            <CircleCheckIcon className="w-5 h-5 mr-2 text-green-600 dark:text-emerald-500" />
                            <div className="flex items-center gap-2">
                              <p>Claim tx broadcasted.</p>
                              <ExternalLink
                                to={`${info?.mempoolUrl}/tx/${swap.claimTxId}`}
                                className="flex items-center underline text-foreground"
                              >
                                View
                              </ExternalLink>
                            </div>
                          </div>
                          <Divider color="border-green-600 dark:border-emerald-500" />
                          <div className="flex items-center text-muted-foreground text-sm">
                            <CircleCheckIcon className="w-5 h-5 mr-2 text-green-600 dark:text-emerald-500" />
                            <div className="flex items-center gap-2">
                              <p>Lockup confirmed.</p>
                              <ExternalLink
                                to={`${info?.mempoolUrl}/tx/${swap.lockupTxId}`}
                                className="flex items-center underline text-foreground"
                              >
                                View
                              </ExternalLink>
                            </div>
                          </div>
                        </>
                      ) : (
                        <>
                          <div className="flex items-center text-muted-foreground text-sm">
                            <Loading className="w-5 h-5 mr-2" />
                            Confirming lockup deposit...
                          </div>
                          <Divider color="border-green-600 dark:border-emerald-500" />
                          <div className="flex items-center text-muted-foreground text-sm">
                            <CircleCheckIcon className="w-5 h-5 mr-2 text-green-600 dark:text-emerald-500" />
                            <div className="flex items-center gap-2">
                              <p>Lockup found in mempool.</p>
                              <ExternalLink
                                to={`${info?.mempoolUrl}/tx/${swap.lockupTxId}`}
                                className="flex items-center underline text-foreground"
                              >
                                View
                              </ExternalLink>
                            </div>
                          </div>
                        </>
                      )}{" "}
                      <Divider color="border-green-600 dark:border-emerald-500" />
                      <div className="flex items-center text-muted-foreground text-sm">
                        <CircleCheckIcon className="w-5 h-5 mr-2 text-green-600 dark:text-emerald-500" />
                        Swap hold invoice paid
                      </div>
                    </>
                  ) : (
                    <div className="flex items-center text-muted-foreground text-sm">
                      <Loading className="w-5 h-5 mr-2" />
                      Paying swap hold invoice...
                    </div>
                  )}
                  <Divider color="border-green-600 dark:border-emerald-500" />
                </>
              )}
              {swapStatus === "FAILED" && (
                <>
                  <div className="flex items-center text-muted-foreground text-sm">
                    <CircleCheckIcon className="w-5 h-5 mr-2 text-green-600 dark:text-emerald-500" />
                    Swap hold invoice cancelled
                  </div>
                  <Divider color="border-green-600 dark:border-emerald-500" />
                  {swap.lockupTxId ? (
                    <>
                      <div className="flex items-center text-muted-foreground text-sm">
                        <CircleAlertIcon className="w-5 h-5 mr-2 text-red-500" />
                        Failed to claim swap in time
                      </div>
                      <Divider color="border-green-600 dark:border-emerald-500" />
                    </>
                  ) : (
                    <>
                      <div className="flex items-center text-muted-foreground text-sm">
                        <CircleAlertIcon className="w-5 h-5 mr-2 text-red-500" />
                        Failed to pay swap hold invoice
                      </div>
                      <Divider color="border-red-500" />
                    </>
                  )}
                </>
              )}
              <div className="flex items-center text-muted-foreground text-sm">
                <CircleCheckIcon className="w-5 h-5 mr-2 text-green-600 dark:text-emerald-500" />
                Swap initiated
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

const Divider = ({ color }: { color: string }) => (
  <div className={`ml-[9px] py-1 border-l ${color}`}></div>
);
