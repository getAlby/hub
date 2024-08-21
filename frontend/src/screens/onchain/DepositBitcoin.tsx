import { Copy, CreditCard, RefreshCw } from "lucide-react";
import { Link } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import Loading from "src/components/Loading";
import QRCode from "src/components/QRCode";
import { Button } from "src/components/ui/button";
import { Card, CardContent } from "src/components/ui/card";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";
import { useOnchainAddress } from "src/hooks/useOnchainAddress";
import { copyToClipboard } from "src/lib/clipboard";

export default function DepositBitcoin() {
  const {
    data: onchainAddress,
    getNewAddress,
    loadingAddress,
  } = useOnchainAddress();
  const { toast } = useToast();

  if (!onchainAddress) {
    return (
      <div className="flex justify-center">
        <Loading />
      </div>
    );
  }

  return (
    <div className="grid gap-5">
      <AppHeader
        title="Deposit Bitcoin to Savings Balance"
        description="Deposit bitcoin to your on-chain address which then can be used to open new lightning channels."
        contentRight={
          <Link to="/channels/onchain/buy-bitcoin">
            <Button>
              <CreditCard className="h-4 w-4 mr-2" />
              Buy Bitcoin
            </Button>
          </Link>
        }
      />
      <div className="w-80">
        <Card>
          <CardContent className="grid gap-6 p-8 justify-center border border-muted">
            <a
              href={`bitcoin:${onchainAddress}`}
              target="_blank"
              className="flex justify-center"
            >
              <QRCode value={onchainAddress} />
            </a>

            <div className="text-center">
              {onchainAddress.match(/.{1,4}/g)?.map((word, index) => {
                if (index % 2 === 0) {
                  return <span className="text-foreground">{word} </span>;
                } else {
                  return <span className="text-muted-foreground">{word} </span>;
                }
              })}
            </div>

            <div className="flex flex-row gap-4 justify-center">
              <LoadingButton
                variant="outline"
                onClick={getNewAddress}
                className="w-28"
                loading={loadingAddress}
              >
                {!loadingAddress && <RefreshCw className="w-4 h-4 mr-2" />}
                Change
              </LoadingButton>
              <Button
                variant="secondary"
                className="w-28"
                onClick={() => {
                  copyToClipboard(onchainAddress);
                  toast({ title: "Copied to clipboard." });
                }}
              >
                <Copy className="w-4 h-4 mr-2" />
                Copy
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
