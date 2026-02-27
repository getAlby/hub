import {
  ArrowLeftIcon,
  CopyIcon,
  ExternalLinkIcon,
  HandCoinsIcon,
} from "lucide-react";
import TickSVG from "public/images/illustrations/tick.svg";
import { useEffect, useState } from "react";
import { FixedFloatButton } from "src/components/FixedFloatButton";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import LottieLoading from "src/components/LottieLoading";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { LinkButton } from "src/components/ui/custom/link-button";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Label } from "src/components/ui/label";
import { useTransaction } from "src/hooks/useTransaction";
import { copyToClipboard } from "src/lib/clipboard";
import { Transaction } from "src/types";

export function FixedFloatSwapInFlow({
  loading,
  transaction,
  onReset,
  resetLabel,
}: {
  loading: boolean;
  transaction: Transaction | null;
  onReset: () => void;
  resetLabel: string;
  feeLabel?: string;
}) {
  const { data: invoiceData } = useTransaction(
    transaction ? transaction.paymentHash : "",
    true
  );
  const [paymentDone, setPaymentDone] = useState(false);

  useEffect(() => {
    if (invoiceData?.settledAt) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setPaymentDone(true);
    }
  }, [invoiceData]);

  if (!transaction) {
    return (
      <>
        <div className="border-t pt-4 text-sm grid gap-2">
          <div className="flex items-center justify-between">
            <Label>Swap Fee</Label>
            <p className="text-muted-foreground">1% + on-chain fees</p>
          </div>
        </div>

        <div className="grid gap-2">
          <LoadingButton className="w-full" loading={loading}>
            Continue
            <ExternalLinkIcon className="size-4" />
          </LoadingButton>
          <p className="text-xs text-muted-foreground text-center">
            powered by{" "}
            <span className="font-medium text-foreground">Fixed Float</span>
          </p>
        </div>
      </>
    );
  }

  if (!paymentDone) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="text-center">Waiting for Payment</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col items-center gap-6">
          <LottieLoading size={288} />
          <div className="flex flex-col gap-1 items-center">
            <p className="text-2xl font-medium slashed-zero">
              <FormattedBitcoinAmount amount={transaction.amount} />
            </p>
            <FormattedFiatAmount
              amount={Math.floor(transaction.amount / 1000)}
              className="text-xl"
            />
          </div>
        </CardContent>
        <CardFooter className="flex flex-col gap-2 pt-2">
          <Button
            type="button"
            className="w-full"
            onClick={() => {
              copyToClipboard(transaction.invoice);
            }}
            variant="secondary"
          >
            <CopyIcon className="w-4 h-4" />
            Copy Invoice
          </Button>
          <FixedFloatButton
            to="BTCLN"
            address={transaction.invoice}
            variant="outline"
            className="w-full"
          >
            <ExternalLinkIcon className="size-4" />
            Open Fixed Float
          </FixedFloatButton>
        </CardFooter>
      </Card>
    );
  }

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle className="text-center">Transaction Received!</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col items-center gap-6">
        <img src={TickSVG} className="w-48" />
        <div className="flex flex-col gap-1 items-center">
          <p className="text-2xl font-medium slashed-zero">
            <FormattedBitcoinAmount amount={transaction.amount} />
          </p>
          <FormattedFiatAmount
            amount={Math.floor(transaction.amount / 1000)}
            className="text-xl"
          />
        </div>
      </CardContent>
      <CardFooter className="flex flex-col gap-2 pt-2">
        <Button
          type="button"
          onClick={() => {
            onReset();
            setPaymentDone(false);
          }}
          variant="outline"
          className="w-full"
        >
          <HandCoinsIcon className="w-4 h-4 mr-2" />
          {resetLabel}
        </Button>
        <LinkButton to="/wallet" variant="link" className="w-full">
          <ArrowLeftIcon className="w-4 h-4 mr-2" />
          Back to Wallet
        </LinkButton>
      </CardFooter>
    </Card>
  );
}
