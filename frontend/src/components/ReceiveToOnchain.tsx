import {
  ArrowLeftIcon,
  CopyIcon,
  ExternalLinkIcon,
  HandCoinsIcon,
  RefreshCwIcon,
} from "lucide-react";
import TickSVG from "public/images/illustrations/tick.svg";
import { useEffect, useRef, useState } from "react";
import { FixedFloatButton } from "src/components/FixedFloatButton";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import Loading from "src/components/Loading";
import LottieLoading from "src/components/LottieLoading";
import OnchainAddressDisplay from "src/components/OnchainAddressDisplay";
import QRCode from "src/components/QRCode";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { ExternalLinkButton } from "src/components/ui/custom/external-link-button";
import { LinkButton } from "src/components/ui/custom/link-button";
import { Separator } from "src/components/ui/separator";
import { useInfo } from "src/hooks/useInfo";
import { useMempoolApi } from "src/hooks/useMempoolApi";
import { useOnchainAddress } from "src/hooks/useOnchainAddress";
import { copyToClipboard } from "src/lib/clipboard";
import { MempoolUtxo } from "src/types";

export function ReceiveToOnchain() {
  const { data: onchainAddress, getNewAddress } = useOnchainAddress();
  const { data: mempoolAddressUtxos } = useMempoolApi<MempoolUtxo[]>(
    onchainAddress ? `/address/${onchainAddress}/utxo` : undefined,
    3000
  );

  const [txId, setTxId] = useState("");
  const [confirmedAmountSat, setConfirmedAmountSat] = useState<number | null>(
    null
  );
  const [pendingAmountSat, setPendingAmountSat] = useState<number | null>(null);
  const startTimeRef = useRef(0);

  useEffect(() => {
    if (startTimeRef.current === 0) {
      startTimeRef.current = Math.floor(Date.now() / 1000);
    }
  }, []);

  const receiveAnother = async () => {
    setTxId("");
    setConfirmedAmountSat(null);
    setPendingAmountSat(null);
    startTimeRef.current = Math.floor(Date.now() / 1000);
    await getNewAddress();
  };

  useEffect(() => {
    if (
      !mempoolAddressUtxos ||
      mempoolAddressUtxos.length === 0 ||
      startTimeRef.current === 0
    ) {
      return;
    }

    if (txId) {
      const utxo = mempoolAddressUtxos.find((utxo) => utxo.txid === txId);
      if (utxo?.status.confirmed) {
        setConfirmedAmountSat(utxo.value);
        setPendingAmountSat(null);
      }
    } else {
      const unconfirmed = mempoolAddressUtxos.find(
        (utxo) => !utxo.status.confirmed
      );
      if (unconfirmed) {
        setTxId(unconfirmed.txid);
        setPendingAmountSat(unconfirmed.value);
        return;
      }

      const confirmed = mempoolAddressUtxos.find(
        (utxo) =>
          utxo.status.confirmed &&
          !!utxo.status.block_time &&
          utxo.status.block_time >= startTimeRef.current
      );
      if (confirmed) {
        setTxId(confirmed.txid);
        setConfirmedAmountSat(confirmed.value);
        setPendingAmountSat(null);
      }
    }
  }, [mempoolAddressUtxos, txId]);

  if (!onchainAddress) {
    return <Loading />;
  }

  return (
    <>
      {confirmedAmountSat ? (
        <DepositSuccess
          amountSat={confirmedAmountSat}
          txId={txId}
          onReceiveAnother={receiveAnother}
        />
      ) : txId ? (
        <DepositPending amountSat={pendingAmountSat} txId={txId} />
      ) : (
        <Card>
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
                copyToClipboard(onchainAddress);
              }}
              variant="secondary"
            >
              <CopyIcon className="w-4 h-4" />
              Copy Address
            </Button>
            <Button
              className="w-full"
              variant="outline"
              onClick={getNewAddress}
            >
              <RefreshCwIcon className="h-4 w-4" />
              New Address
            </Button>
            <Separator className="my-4" />
            <FixedFloatButton
              to="BTC"
              address={onchainAddress}
              className="w-full"
              variant="secondary"
            >
              <ExternalLinkIcon className="size-4" />
              Top up using other Cryptocurrency
            </FixedFloatButton>
          </CardFooter>
        </Card>
      )}
    </>
  );
}

function DepositPending({
  amountSat,
  txId,
}: {
  amountSat: number | null;
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
        {amountSat && (
          <div className="flex flex-col gap-1 items-center">
            <p className="text-2xl font-medium slashed-zero">
              <FormattedBitcoinAmount amountMsat={amountSat * 1000} />
            </p>
            <FormattedFiatAmount amountSat={amountSat} className="text-xl" />
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

function DepositSuccess({
  amountSat,
  txId,
  onReceiveAnother,
}: {
  amountSat: number;
  txId: string;
  onReceiveAnother: () => void;
}) {
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
            <FormattedBitcoinAmount amountMsat={amountSat * 1000} />
          </p>
          <FormattedFiatAmount amountSat={amountSat} className="text-xl" />
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
        <Button
          type="button"
          variant="outline"
          className="w-full"
          onClick={onReceiveAnother}
        >
          <HandCoinsIcon className="w-4 h-4 mr-2" />
          Receive Another Payment
        </Button>
        <LinkButton to="/wallet" variant="link" className="w-full">
          <ArrowLeftIcon className="w-4 h-4 mr-2" />
          Back to Wallet
        </LinkButton>
      </CardFooter>
    </Card>
  );
}
