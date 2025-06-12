import {
  CircleCheckIcon,
  CircleEllipsisIcon,
  CircleXIcon,
  CopyIcon,
  ExternalLinkIcon,
  ZapIcon,
} from "lucide-react";
import { useState } from "react";
import { Link, useLocation } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import QRCode from "src/components/QRCode";
import { Badge } from "src/components/ui/badge";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";
import { useBalances } from "src/hooks/useBalances";
import { useMempoolApi } from "src/hooks/useMempoolApi";
import { useSyncWallet } from "src/hooks/useSyncWallet";
import { useTransaction } from "src/hooks/useTransaction";
import { copyToClipboard } from "src/lib/clipboard";
import { RedeemOnchainFundsResponse, SwapInResponse } from "src/types";
import { request } from "src/utils/request";

export default function SwapInStatus() {
  const { state } = useLocation();
  const { toast } = useToast();
  const [isPaying, setPaying] = useState(false);
  const [hasPaidWithHubFunds, setPaidWithHubFunds] = useState(false);
  useSyncWallet(); // ensure funds show up on node page after swap completes

  const swapInResponse = state.swapInResponse as SwapInResponse;
  const amount = swapInResponse.amountToDeposit;
  const { data: lightningPayment } = useTransaction(
    swapInResponse.paymentHash,
    true
  );

  const { data: balances } = useBalances();

  // TODO: fetch the onchain transaction to view status
  const { data: addressTransactions } = useMempoolApi<
    { txId: string; status: { confirmed: boolean } }[]
  >(`/address/${swapInResponse.onchainAddress}/txs`, true);

  async function payWithAlbyHub() {
    setPaying(true);

    try {
      const response = await request<RedeemOnchainFundsResponse>(
        "/api/wallet/redeem-onchain-funds",
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            toAddress: swapInResponse.onchainAddress,
            amount: swapInResponse.amountToDeposit,
          }),
        }
      );
      console.info("Redeemed onchain funds", response);
      if (!response?.txId) {
        throw new Error("No address in response");
      }
      //setTransactionId(response.txId);
      setPaidWithHubFunds(true);
    } catch (error) {
      console.error(error);
      toast({
        variant: "destructive",
        title: "Failed to redeem onchain funds",
        description: "" + error,
      });
    }
    setPaying(false);
  }

  const copyAddress = () => {
    copyToClipboard(swapInResponse.onchainAddress, toast);
  };
  const copyAmount = () => {
    copyToClipboard(swapInResponse.amountToDeposit.toString(), toast);
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
                  : "Pending"}
            </CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col items-center gap-4">
            {swapStatus === "success" ? (
              <CircleCheckIcon className="w-32 h-32 mb-2" />
            ) : swapStatus === "failed" ? (
              <CircleXIcon className="w-32 h-32 mb-2" />
            ) : (
              <CircleEllipsisIcon className="animate-pulse w-32 h-32 mb-2" />
            )}
            <div className="flex flex-col items-center">
              <p className="text-xl font-bold slashed-zero text-center">
                {new Intl.NumberFormat().format(amount)} sats
              </p>
              <FormattedFiatAmount amount={amount} />
            </div>

            {swapStatus === "pending" && (
              <>
                {isPaying && <p>Broadcasting On-Chain payment...</p>}
                {!addressTransactions?.length &&
                  !isPaying &&
                  !hasPaidWithHubFunds && (
                    <div className="flex flex-col gap-2 items-center">
                      <p className="slashed-zero text-center text-muted-foreground max-w-xs">
                        Please deposit {new Intl.NumberFormat().format(amount)}{" "}
                        sats to the address using your bitcoin on-chain wallet.
                      </p>
                      <a
                        href={`bitcoin:${swapInResponse.onchainAddress}`}
                        target="_blank"
                        className="flex justify-center"
                      >
                        <QRCode value={swapInResponse.onchainAddress} />
                      </a>
                    </div>
                  )}
                {!addressTransactions?.length && (
                  <div className="flex flex-col gap-2 items-center">
                    <Badge>Waiting for onchain payment... </Badge>
                  </div>
                )}
                {!!addressTransactions?.length && (
                  <div className="flex flex-col gap-2 items-center">
                    <Badge>
                      {addressTransactions.every((tx) => tx.status.confirmed)
                        ? "Transaction confirmed"
                        : "Transaction in mempool"}
                    </Badge>
                  </div>
                )}
              </>
            )}
            {swapInResponse.onchainAddress && (
              <div className="flex items-center gap-4 flex-wrap">
                <ExternalLink
                  to={`https://mempool.space/address/${swapInResponse.onchainAddress}`}
                  className="flex items-center"
                >
                  <Button variant="outline" className="w-full sm:w-auto">
                    View on Mempool
                    <ExternalLinkIcon className="w-4 h-4 ml-2" />
                  </Button>
                </ExternalLink>
                <Button onClick={copyAddress} variant="outline">
                  <CopyIcon className="w-4 h-4 mr-2" />
                  Copy Address
                </Button>
                {!isPaying && !hasPaidWithHubFunds && (
                  <Button onClick={copyAmount} variant="outline">
                    <CopyIcon className="w-4 h-4 mr-2" />
                    Copy Amount
                  </Button>
                )}
                {!isPaying &&
                  !hasPaidWithHubFunds &&
                  balances &&
                  balances.onchain.spendable - 25000 /* anchor reserve */ >
                    swapInResponse.amountToDeposit && (
                    <LoadingButton onClick={payWithAlbyHub} variant="outline">
                      <ZapIcon className="w-4 h-4 mr-2" />
                      Use Hub On-Chain Funds
                    </LoadingButton>
                  )}
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
