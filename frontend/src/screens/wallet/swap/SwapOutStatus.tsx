import {
  CircleAlertIcon,
  CircleCheckIcon,
  CircleXIcon,
  CopyIcon,
} from "lucide-react";
import { useParams } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import Loading from "src/components/Loading";
import LottieLoading from "src/components/LottieLoading";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { useInfo } from "src/hooks/useInfo";
import { useSwap } from "src/hooks/useSwaps";
import { copyToClipboard } from "src/lib/clipboard";
import { SwapOut } from "src/types";

export default function SwapOutStatus() {
  const { data: info } = useInfo();
  const { swapId } = useParams() as { swapId: string };
  const { data: swap } = useSwap<SwapOut>(swapId, true);

  if (!swap) {
    return <Loading />;
  }

  const copyTxId = () => {
    copyToClipboard(swap.claimTxId as string);
  };

  const swapStatus = swap.state;
  const statusText = {
    SUCCESS: "Swap Successful",
    FAILED: "Swap Failed",
    PENDING: swap.lockupTxId
      ? "Waiting for 1 confirmation"
      : "Making lightning payment",
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
            <CardDescription className="flex justify-center text-muted-foreground text-sm">
              {swap.autoSwap && (
                <span>Auto swap{swap.usedXpub && <> to xpub</>} • </span>
              )}
              Swap ID: {swap.id}
            </CardDescription>
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
                <div className="flex justify-center gap-4 flex-wrap">
                  <Button onClick={copyTxId} variant="outline">
                    <CopyIcon />
                    Copy Transaction ID
                  </Button>
                </div>
              </>
            ) : (
              <>
                {swapStatus === "PENDING" ? (
                  <LottieLoading />
                ) : (
                  <CircleXIcon className="w-60 h-60" />
                )}
                <div className="flex flex-col gap-2 items-center">
                  <p className="text-xl font-bold slashed-zero text-center">
                    {new Intl.NumberFormat().format(swap.sendAmount)} sats
                  </p>
                  <div className="flex items-center">
                    <span className="text-sm text-muted-foreground">~</span>
                    <FormattedFiatAmount amount={swap.sendAmount} />
                  </div>
                </div>
              </>
            )}
            <div className="flex flex-col justify-start gap-2 w-full mt-2">
              {swapStatus !== "FAILED" ? (
                swap.lockupTxId ? (
                  <>
                    {swapStatus === "SUCCESS" && (
                      <div className="flex items-center text-muted-foreground text-sm">
                        <CircleCheckIcon className="w-5 h-5 mr-2 text-green-600 dark:text-emerald-500" />
                        <div className="flex items-center gap-2">
                          <p>Confirmed onchain</p>
                          <ExternalLink
                            to={`${info?.mempoolUrl}/tx/${swap.claimTxId}`}
                            className="flex items-center underline text-foreground"
                          >
                            View
                          </ExternalLink>
                        </div>
                      </div>
                    )}
                    {swapStatus === "PENDING" && (
                      <div className="flex items-center text-muted-foreground text-sm">
                        <Loading className="w-5 h-5 mr-2" />
                        <div className="flex items-center gap-2">
                          <p>
                            Waiting for{" "}
                            {swap.claimTxId
                              ? "onchain confirmation"
                              : "2 onchain confirmations"}
                            ...
                          </p>
                          <ExternalLink
                            to={`${info?.mempoolUrl}/tx/${swap.claimTxId || swap.lockupTxId}`}
                            className="flex items-center underline text-foreground"
                          >
                            View
                          </ExternalLink>
                        </div>
                      </div>
                    )}
                    <Divider color="border-green-600 dark:border-emerald-500" />
                    <div className="flex items-center text-muted-foreground text-sm">
                      <CircleCheckIcon className="w-5 h-5 mr-2 text-green-600 dark:text-emerald-500" />
                      Lightning invoice paid
                    </div>
                    <Divider color="border-green-600 dark:border-emerald-500" />
                  </>
                ) : (
                  <>
                    <div className="flex items-center text-muted-foreground text-sm">
                      <Loading className="w-5 h-5 mr-2" />
                      Paying lightning invoice...
                    </div>
                    <Divider color="border-green-600 dark:border-emerald-500" />
                  </>
                )
              ) : (
                <>
                  <div className="flex items-center text-muted-foreground text-sm">
                    <CircleAlertIcon className="w-5 h-5 mr-2 text-red-500" />
                    Swap failed
                  </div>
                  <Divider color="border-red-500" />
                  {swap.lockupTxId ? (
                    <>
                      <div className="flex items-center text-muted-foreground text-sm">
                        <CircleAlertIcon className="w-5 h-5 mr-2 text-red-500" />
                        Failed to claim swap in time
                      </div>
                      <Divider color="border-red-500" />
                    </>
                  ) : (
                    <>
                      <div className="flex items-center text-muted-foreground text-sm">
                        <CircleAlertIcon className="w-5 h-5 mr-2 text-red-500" />
                        Failed to pay swap invoice
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
