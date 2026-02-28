import { CoinsIcon, ExternalLinkIcon } from "lucide-react";
import { FixedFloatButton } from "src/components/FixedFloatButton";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { useBitcoinMaxiMode } from "src/hooks/useBitcoinMaxiMode";

type CryptoSwapAlertProps = {
  className?: string;
};
export function CryptoSwapAlert({ className }: CryptoSwapAlertProps) {
  const { bitcoinMaxiMode } = useBitcoinMaxiMode();

  if (bitcoinMaxiMode) {
    return null;
  }

  return (
    <Alert className={className}>
      <AlertTitle className="flex items-center gap-2">
        <CoinsIcon className="h-4 w-4" />
        Looking to pay to other Cryptocurrency?
      </AlertTitle>
      <AlertDescription className="text-xs gap-2 mt-1">
        <p>
          If you are trying to pay a non-Bitcoin payment destination, use
          FixedFloat to swap and complete the payment across 70+ supported
          cryptocurrencies.
        </p>
        <FixedFloatButton
          from="BTCLN"
          variant="outline"
          className="text-foreground"
        >
          Pay with FixedFloat
          <ExternalLinkIcon className="size-4" />
        </FixedFloatButton>
      </AlertDescription>
    </Alert>
  );
}
