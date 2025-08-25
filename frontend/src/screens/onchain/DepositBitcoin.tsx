import {
  CircleCheckIcon,
  CopyIcon,
  CreditCardIcon,
  ExternalLinkIcon,
  RefreshCwIcon,
} from "lucide-react";
import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import Loading from "src/components/Loading";
import LottieLoading from "src/components/LottieLoading";
import { MempoolAlert } from "src/components/MempoolAlert";
import OnchainAddressDisplay from "src/components/OnchainAddressDisplay";
import QRCode from "src/components/QRCode";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { useInfo } from "src/hooks/useInfo";
import { useMempoolApi } from "src/hooks/useMempoolApi";
import { useOnchainAddress } from "src/hooks/useOnchainAddress";
import { useSyncWallet } from "src/hooks/useSyncWallet";
import { copyToClipboard } from "src/lib/clipboard";
import { MempoolUtxo } from "src/types";

export default function DepositBitcoin() {
  useSyncWallet();
  const {
    data: onchainAddress,
    getNewAddress,
    loadingAddress,
  } = useOnchainAddress();
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
    <div className="grid gap-5">
      <AppHeader
        title="Deposit Bitcoin to On-Chain Balance"
        description="Deposit bitcoin to your on-chain address which then can be used to open new lightning channels."
        contentRight={
          <Link to="/channels/onchain/buy-bitcoin">
            <Button>
              <CreditCardIcon />
              Buy Bitcoin
            </Button>
          </Link>
        }
      />
      <MempoolAlert />
      <div className="w-80">
        {confirmedAmount ? (
          <DepositSuccess amount={confirmedAmount} txId={txId} />
        ) : txId ? (
          <DepositPending amount={pendingAmount} txId={txId} />
        ) : (
          <Card>
            <CardContent className="grid gap-6 justify-center">
              <a
                href={`bitcoin:${onchainAddress}`}
                target="_blank"
                className="flex justify-center"
              >
                <QRCode value={onchainAddress} />
              </a>

              <div className="flex flex-wrap gap-2 items-center justify-center">
                <OnchainAddressDisplay address={onchainAddress} />
              </div>

              <div className="flex flex-row gap-4 justify-center">
                <LoadingButton
                  variant="outline"
                  onClick={getNewAddress}
                  className="w-28"
                  loading={loadingAddress}
                >
                  {!loadingAddress && <RefreshCwIcon />}
                  Change
                </LoadingButton>
                <Button
                  variant="secondary"
                  className="w-28"
                  onClick={() => {
                    copyToClipboard(onchainAddress);
                  }}
                >
                  <CopyIcon />
                  Copy
                </Button>
              </div>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
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
        <CardTitle className="text-center">Awaiting Confirmation</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col items-center gap-4">
        <LottieLoading size={288} />
        {amount && (
          <div className="flex flex-col gap-2 items-center">
            <p className="text-xl font-semibold slashed-zero">
              {new Intl.NumberFormat().format(amount)} sats
            </p>
            <FormattedFiatAmount amount={amount} />
          </div>
        )}
        <div>
          <Button asChild variant="outline">
            <ExternalLink
              to={`${info?.mempoolUrl}/tx/${txId}`}
              className="flex items-center mt-2"
            >
              View on Mempool
              <ExternalLinkIcon className="size-4 ml-2" />
            </ExternalLink>
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

function DepositSuccess({ amount, txId }: { amount: number; txId: string }) {
  const { data: info } = useInfo();

  return (
    <>
      <Card className="w-full">
        <CardHeader>
          <CardTitle className="text-center">Payment Received!</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col items-center gap-4">
          <CircleCheckIcon className="w-72 h-72 p-2" />
          <div className="flex flex-col gap-2 items-center">
            <p className="text-xl font-semibold slashed-zero">
              {new Intl.NumberFormat().format(amount)} sats
            </p>
            <FormattedFiatAmount amount={amount} />
          </div>
          <div>
            <Button asChild variant="outline">
              <ExternalLink
                to={`${info?.mempoolUrl}/tx/${txId}`}
                className="flex items-center mt-2"
              >
                View on Mempool
                <ExternalLinkIcon className="size-4 ml-2" />
              </ExternalLink>
            </Button>
          </div>
        </CardContent>
      </Card>
      <Link to="/channels">
        <Button className="mt-4 w-full">Back To Node</Button>
      </Link>
    </>
  );
}
