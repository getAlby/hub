import { CircleCheckIcon, CopyIcon, ExternalLinkIcon } from "lucide-react";
import { Link, useLocation } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import { Badge } from "src/components/ui/badge";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { useToast } from "src/components/ui/use-toast";
import { useSyncWallet } from "src/hooks/useSyncWallet";
import { useTransaction } from "src/hooks/useTransaction";
import { copyToClipboard } from "src/lib/clipboard";
import { SwapOutResponse } from "src/types";

export default function SwapOutStatus() {
  const { state } = useLocation();
  const { toast } = useToast();
  useSyncWallet(); // ensure funds show up on node page after swap completes

  const swapOutResponse = state.swapOutResponse as SwapOutResponse;
  const amount = state.amount as number;
  const { data: lightningPayment } = useTransaction(
    swapOutResponse.paymentHash,
    true
  );

  const copy = () => {
    copyToClipboard(swapOutResponse.txId as string, toast);
  };

  const swapStatus =
    lightningPayment?.state === "settled"
      ? "success"
      : lightningPayment?.state === "failed"
        ? "failed"
        : "pending";

  return (
    <div className="grid gap-5">
      <AppHeader title="Swap Out" />
      <div className="w-full max-w-lg">
        <Card className="w-full">
          <CardHeader>
            <CardTitle className="text-center">
              Swap{" "}
              {swapStatus === "success"
                ? "Successful"
                : swapStatus === "failed"
                  ? "Failed"
                  : "Initiated"}
            </CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col items-center gap-4">
            {swapStatus === "success" ? (
              <CircleCheckIcon className="w-32 h-32 mb-2" />
            ) : (
              <Badge> {lightningPayment?.state || "Loading..."}</Badge>
            )}

            {/*  */}
            <div className="flex flex-col gap-2 items-center">
              <p className="text-xl font-bold slashed-zero">
                {new Intl.NumberFormat().format(amount)} sats
              </p>
              <FormattedFiatAmount amount={amount} />
            </div>
            {swapOutResponse.txId && (
              <div className="flex items-center gap-4">
                <ExternalLink
                  to={`https://mempool.space/tx/${swapOutResponse.txId}`}
                  className="flex items-center"
                >
                  <Button variant="outline" className="w-full sm:w-auto">
                    View on Mempool
                    <ExternalLinkIcon className="w-4 h-4 ml-2" />
                  </Button>
                </ExternalLink>
                <Button onClick={copy} variant="outline">
                  <CopyIcon className="w-4 h-4 mr-2" />
                  Copy TxId
                </Button>
              </div>
            )}
          </CardContent>
        </Card>
        <Link to={`/wallet/swap?type=out`}>
          <Button className="mt-4 w-full">Make Another Swap</Button>
        </Link>
        <Link to="/wallet">
          <Button className="mt-4 w-full" variant="secondary">
            Back To Wallet
          </Button>
        </Link>
      </div>
    </div>
  );
}
