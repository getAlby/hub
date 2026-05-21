import { Skeleton } from "src/components/ui/skeleton";
import { useBitcoinRate } from "src/hooks/useBitcoinRate";
import { useInfo } from "src/hooks/useInfo";
import { cn } from "src/lib/utils";

type FormattedFiatAmountProps = {
  amountSat: number;
  className?: string;
  showApprox?: boolean;
};

export default function FormattedFiatAmount({
  amountSat,
  className,
  showApprox,
}: FormattedFiatAmountProps) {
  const { data: info } = useInfo();
  const { data: bitcoinRate, error: bitcoinRateError } = useBitcoinRate(
    info?.currency
  );

  if (info?.currency === "SATS" || bitcoinRateError) {
    return null;
  }

  return (
    <div className={cn("text-sm text-muted-foreground", className)}>
      {showApprox && bitcoinRate && "~"}
      {!bitcoinRate ? (
        <Skeleton className="w-20">&nbsp;</Skeleton>
      ) : (
        new Intl.NumberFormat("en-US", {
          style: "currency",
          currency: info?.currency || "usd",
        }).format((amountSat / 100_000_000) * bitcoinRate.rate_float)
      )}
    </div>
  );
}
