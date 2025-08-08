import {
  ArrowLeftIcon,
  CopyIcon,
  ExternalLinkIcon,
  HandCoinsIcon,
  RefreshCwIcon,
} from "lucide-react";
import { useEffect, useState } from "react";
import animationDataDark from "src/assets/lotties/loading-dark.json";
import animationDataLight from "src/assets/lotties/loading-light.json";
import AppHeader from "src/components/AppHeader";
import Loading from "src/components/Loading";
import QRCode from "src/components/QRCode";
import {
  Button,
  ExternalLinkButton,
  LinkButton,
} from "src/components/ui/button";
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
import { MempoolUtxo } from "src/types";

import TickSVG from "public/images/illustrations/tick.svg";
import Lottie from "react-lottie";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import { useTheme } from "src/components/ui/theme-provider";

export default function ReceiveOnchain() {
  const [receiveType, setReceiveType] = useState("onchain");

  useEffect(() => {
    const queryParams = new URLSearchParams(location.search);
    const receiveType = queryParams.get("type");
    if (receiveType) {
      setReceiveType(receiveType);
    }
  }, []);

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
              {onchainAddress.match(/.{1,4}/g)?.map((word, index) => {
                if (index % 2 === 0) {
                  return (
                    <span key={index} className="text-foreground">
                      {word}
                    </span>
                  );
                } else {
                  return (
                    <span key={index} className="text-muted-foreground">
                      {word}
                    </span>
                  );
                }
              })}
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
  const { isDarkMode } = useTheme();

  const defaultOptions = {
    loop: true,
    autoplay: true,
    animationData: isDarkMode ? animationDataDark : animationDataLight,
    rendererSettings: {
      preserveAspectRatio: "xMidYMid slice",
    },
  };

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle className="flex items-center justify-center gap-2">
          <Loading className="w-4 h-4" />
          Waiting for On-chain Confirmation...
        </CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col items-center gap-4">
        <Lottie options={defaultOptions} height={288} width={288} />
        {amount && (
          <div className="flex flex-col gap-1 items-center">
            <p className="text-2xl font-medium slashed-zero">
              {new Intl.NumberFormat().format(amount)} sats
            </p>
            <FormattedFiatAmount amount={amount} className="text-xl" />
          </div>
        )}
        <p className="text-muted-foreground text-center mt-6">
          You can now leave this page.
          <br />
          You'll be notified when transaction arrives.
        </p>
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
  return <></>;
}
