import {
  AlertTriangleIcon,
  ArrowLeftIcon,
  CopyIcon,
  ExternalLinkIcon,
  HandCoinsIcon,
  RefreshCwIcon,
} from "lucide-react";
import { useEffect, useState } from "react";
import AppHeader from "src/components/AppHeader";
import Loading from "src/components/Loading";
import QRCode from "src/components/QRCode";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { useToast } from "src/components/ui/use-toast";
import { useInfo } from "src/hooks/useInfo";
import { useMempoolApi } from "src/hooks/useMempoolApi";
import { useOnchainAddress } from "src/hooks/useOnchainAddress";
import { copyToClipboard } from "src/lib/clipboard";
import { cn } from "src/lib/utils";
import { MempoolUtxo, SwapResponse } from "src/types";

import TickSVG from "public/images/illustrations/tick.svg";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import LottieLoading from "src/components/LottieLoading";
import OnchainAddressDisplay from "src/components/OnchainAddressDisplay";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { ExternalLinkButton } from "src/components/ui/custom/external-link-button";
import { InputWithAdornment } from "src/components/ui/custom/input-with-adornment";
import { LinkButton } from "src/components/ui/custom/link-button";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Label } from "src/components/ui/label";
import { useBalances } from "src/hooks/useBalances";
import { useSwapInfo } from "src/hooks/useSwaps";
import { request } from "src/utils/request";

export default function ReceiveOnchain() {
  const [receiveType, setReceiveType] = useState("spending");
  const [searchParams] = useSearchParams();

  useEffect(() => {
    const receiveType = searchParams.get("type");
    if (receiveType) {
      setReceiveType(receiveType);
    }
  }, [searchParams]);

  return (
    <div className="grid gap-5">
      <AppHeader title="Receive On-chain" />
      <div className="w-full max-w-lg grid gap-5">
        <div className="flex items-center text-center text-foreground font-medium rounded-xl bg-muted p-1">
          <div
            className={cn(
              "cursor-pointer rounded-lg flex-1 py-1.5 text-sm",
              receiveType == "onchain" &&
                "text-foreground bg-background shadow-md"
            )}
            onClick={() => setReceiveType("onchain")}
          >
            Receive to On-chain
          </div>
          <div
            className={cn(
              "cursor-pointer rounded-lg flex-1 py-1.5 text-sm",
              receiveType == "spending" &&
                "text-foreground bg-background shadow-md"
            )}
            onClick={() => setReceiveType("spending")}
          >
            Receive to Spending
          </div>
        </div>
        {receiveType == "onchain" ? (
          <ReceiveToOnchain />
        ) : (
          <ReceiveToSpending />
        )}
      </div>
    </div>
  );
}

function ReceiveToOnchain() {
  const { data: onchainAddress, getNewAddress } = useOnchainAddress();
  const { toast } = useToast();
  const { data: mempoolAddressUtxos } = useMempoolApi<MempoolUtxo[]>(
    onchainAddress ? `/address/${onchainAddress}/utxo` : undefined,
    3000
  );

  const [txId, setTxId] = useState("");
  const [confirmedAmount, setConfirmedAmount] = useState<number | null>(null);
  const [pendingAmount, setPendingAmount] = useState<number | null>(null);

  useEffect(() => {
    if (!mempoolAddressUtxos || mempoolAddressUtxos.length === 0) {
      return;
    }

    if (txId) {
      const utxo = mempoolAddressUtxos.find((utxo) => utxo.txid === txId);
      if (utxo?.status.confirmed) {
        setConfirmedAmount(utxo.value);
        setPendingAmount(null);
      }
    } else {
      const unconfirmed = mempoolAddressUtxos.find(
        (utxo) => !utxo.status.confirmed
      );
      if (unconfirmed) {
        setTxId(unconfirmed.txid);
        setPendingAmount(unconfirmed.value);
      }
    }
  }, [mempoolAddressUtxos, txId]);

  if (!onchainAddress) {
    return <Loading />;
  }

  return (
    <>
      {confirmedAmount ? (
        <DepositSuccess amount={confirmedAmount} txId={txId} />
      ) : txId ? (
        <DepositPending amount={pendingAmount} txId={txId} />
      ) : (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center justify-center gap-2">
              <Loading className="w-4 h-4" />
              Waiting for Payment...
            </CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col items-center gap-6">
            <a
              href={`bitcoin:${onchainAddress}`}
              target="_blank"
              className="flex justify-center"
            >
              <QRCode value={onchainAddress} />
            </a>
            <div className="flex flex-wrap max-w-64 gap-2 items-center justify-center">
              <OnchainAddressDisplay address={onchainAddress} />
            </div>
          </CardContent>
          <CardFooter className="flex flex-col gap-2 pt-2">
            <Button
              className="w-full"
              onClick={() => {
                copyToClipboard(onchainAddress, toast);
              }}
              variant="secondary"
            >
              <CopyIcon className="w-4 h-4 mr-2" />
              Copy Address
            </Button>
            <Button
              className="w-full"
              variant="outline"
              onClick={getNewAddress}
            >
              <RefreshCwIcon className="h-4 w-4 shrink-0 mr-2" />
              New Address
            </Button>
          </CardFooter>
        </Card>
      )}
    </>
  );
}

function DepositPending({
  amount,
  txId,
}: {
  amount: number | null;
  txId: string;
}) {
  const { data: info } = useInfo();

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle className="flex items-center justify-center gap-2">
          <Loading className="w-4 h-4" />
          Waiting for On-chain Confirmation...
        </CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col items-center gap-4">
        <LottieLoading size={288} />
        {amount && (
          <div className="flex flex-col gap-1 items-center">
            <p className="text-2xl font-medium slashed-zero">
              {new Intl.NumberFormat().format(amount)} sats
            </p>
            <FormattedFiatAmount amount={amount} className="text-xl" />
          </div>
        )}
      </CardContent>
      <CardFooter className="flex flex-col gap-2 pt-2">
        <ExternalLinkButton
          to={`${info?.mempoolUrl}/tx/${txId}`}
          variant="outline"
          className="w-full"
        >
          <ExternalLinkIcon className="w-4 h-4 mr-2" />
          View on Mempool
        </ExternalLinkButton>
      </CardFooter>
    </Card>
  );
}

function DepositSuccess({ amount, txId }: { amount: number; txId: string }) {
  const { data: info } = useInfo();

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle className="text-center">Transaction Received!</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col items-center gap-6">
        <img src={TickSVG} className="w-48" />
        <div className="flex flex-col gap-1 items-center">
          <p className="text-2xl font-medium slashed-zero">
            {new Intl.NumberFormat().format(amount)} sats
          </p>
          <FormattedFiatAmount amount={amount} className="text-xl" />
        </div>
      </CardContent>
      <CardFooter className="flex flex-col gap-2 pt-2">
        <ExternalLinkButton
          to={`${info?.mempoolUrl}/tx/${txId}`}
          variant="outline"
          className="w-full"
        >
          <ExternalLinkIcon className="w-4 h-4 mr-2" />
          View on Mempool
        </ExternalLinkButton>
        <LinkButton to="/wallet/send" variant="outline" className="w-full">
          <HandCoinsIcon className="w-4 h-4 mr-2" />
          Receive Another Payment
        </LinkButton>
        <LinkButton to="/wallet" variant="link" className="w-full">
          <ArrowLeftIcon className="w-4 h-4 mr-2" />
          Back to Wallet
        </LinkButton>
      </CardFooter>
    </Card>
  );
}

function ReceiveToSpending() {
  const { toast } = useToast();
  const { data: info, hasChannelManagement } = useInfo();
  const { data: balances } = useBalances();
  const { data: swapInfo } = useSwapInfo("in");
  const { data: recommendedFees, error: mempoolError } = useMempoolApi<{
    fastestFee: number;
    halfHourFee: number;
    economyFee: number;
    minimumFee: number;
  }>("/v1/fees/recommended");
  const navigate = useNavigate();

  const [swapAmount, setSwapAmount] = useState("");
  const [loading, setLoading] = useState(false);
  const [feeRate, setFeeRate] = useState("");

  useEffect(() => {
    if (recommendedFees?.fastestFee) {
      setFeeRate(recommendedFees.fastestFee.toString());
    }
  }, [recommendedFees]);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      setLoading(true);
      const swapInResponse = await request<SwapResponse>("/api/swaps/in", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          swapAmount: parseInt(swapAmount),
        }),
      });
      if (!swapInResponse) {
        throw new Error("Error swapping in");
      }
      navigate(`/wallet/swap/in/status/${swapInResponse.swapId}`);
      toast({ title: "Initiated swap" });
    } catch (error) {
      toast({
        title: "Failed to initiate swap",
        description: (error as Error).message,
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  if (!info || !balances || !swapInfo || (!recommendedFees && !mempoolError)) {
    return <Loading />;
  }

  return (
    <form onSubmit={onSubmit} className="flex flex-col gap-6">
      {hasChannelManagement &&
        parseInt(swapAmount || "0") * 1000 >=
          0.8 * balances.lightning.totalReceivable && (
          <Alert>
            <AlertTriangleIcon className="h-4 w-4" />
            <AlertTitle>Low receiving capacity</AlertTitle>
            <AlertDescription>
              You likely won't be able to receive payments until you{" "}
              <Link className="underline" to="/channels/incoming">
                increase your receiving capacity.
              </Link>
            </AlertDescription>
          </Alert>
        )}
      <div className="grid gap-1.5">
        <Label>Amount</Label>
        <InputWithAdornment
          type="number"
          autoFocus
          placeholder="Amount in satoshis"
          value={swapAmount}
          min={swapInfo.minAmount}
          max={Math.min(
            swapInfo.maxAmount,
            (balances.lightning.totalReceivable / 1000) * 0.99
          )}
          onChange={(e) => setSwapAmount(e.target.value)}
          required
          endAdornment={
            <FormattedFiatAmount amount={+swapAmount} className="mr-2" />
          }
        />
        <div className="grid">
          <div className="flex justify-between text-muted-foreground text-xs sensitive slashed-zero">
            <div>
              Receiving Capacity:{" "}
              {new Intl.NumberFormat().format(
                Math.floor(balances.lightning.totalReceivable / 1000)
              )}{" "}
              sats{" "}
              <Link className="underline" to="/channels/incoming">
                increase
              </Link>
            </div>
            <FormattedFiatAmount
              className="text-xs"
              amount={balances.lightning.totalReceivable / 1000}
            />
          </div>
          <div className="flex justify-between text-muted-foreground text-xs sensitive slashed-zero">
            <div>
              Minimum: {new Intl.NumberFormat().format(swapInfo.minAmount)} sats
            </div>
            <FormattedFiatAmount
              className="text-xs"
              amount={swapInfo.minAmount}
            />
          </div>
        </div>
      </div>

      <div className="border-t pt-4 text-sm grid gap-2">
        <div className="flex items-center justify-between">
          <p className="text-muted-foreground">On-chain Fee</p>
          {feeRate ? <p>{feeRate} sat/vB</p> : <Loading className="w-4 h-4" />}
        </div>
        <div className="flex items-center justify-between">
          <p className="text-muted-foreground">Swap Fee</p>
          <p>{swapInfo.albyServiceFee + swapInfo.boltzServiceFee}%</p>
        </div>
      </div>
      <div className="grid gap-2">
        <LoadingButton className="w-full" loading={loading}>
          Continue
        </LoadingButton>
      </div>
    </form>
  );
}
