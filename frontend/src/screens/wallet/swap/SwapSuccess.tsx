import { CircleCheckIcon, CopyIcon, ExternalLinkIcon } from "lucide-react";
import { Link, useLocation } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { useToast } from "src/components/ui/use-toast";
import { copyToClipboard } from "src/lib/clipboard";

export default function SwapSuccess() {
  const { state } = useLocation();
  const { toast } = useToast();

  const copy = () => {
    copyToClipboard(state.txId as string, toast);
  };

  return (
    <div className="grid gap-5">
      <AppHeader title={state.isAutoSwap ? "Auto Swap" : "Swap"} />
      <div className="w-full max-w-lg">
        <Card className="w-full">
          <CardHeader>
            <CardTitle className="text-center">
              {state.isAutoSwap ? "Auto swaps enabled" : "Swap Initiated"}
            </CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col items-center gap-4">
            <CircleCheckIcon className="w-32 h-32 mb-2" />
            <div className="flex flex-col gap-2 items-center">
              <p className="text-xl font-bold slashed-zero">
                {new Intl.NumberFormat().format(state.amount)} sats
              </p>
              <FormattedFiatAmount amount={state.amount} />
              {state.isAutoSwap && (
                <div className="text-sm">
                  Will be swapped everytime balance reaches{" "}
                  <span className="font-bold slashed-zero">
                    {new Intl.NumberFormat().format(state.balanceThreshold)}{" "}
                    sats
                  </span>
                </div>
              )}
            </div>
            {state.txId && (
              <div className="flex items-center gap-4">
                <ExternalLink
                  to={`https://mempool.space/tx/${state.txId}`}
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
        <Link to={`/wallet/swap?type=${state.type}`}>
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
