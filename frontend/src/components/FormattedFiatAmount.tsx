import { Skeleton } from "src/components/ui/skeleton";
import { useBitcoinRate } from "src/hooks/useBitcoinRate";
import { useInfo } from "src/hooks/useInfo";
import { cn } from "src/lib/utils";

type FormattedFiatAmountProps = {
  amount: number;
  className?: string;
};

export default function FormattedFiatAmount({
  amount,
  className,
}: FormattedFiatAmountProps) {
  const { data: info } = useInfo();
  const { data: bitcoinRate } = useBitcoinRate();

  if (info?.currency === "SATS") {
    return null;
  }

  return (
    <div className={cn("text-sm", className)}>
      {!bitcoinRate ? (
        <Skeleton className="w-20">&nbsp;</Skeleton>
      ) : (
        new Intl.NumberFormat("en-US", {
          style: "currency",
          currency: info?.currency || "usd",
        }).format((amount / 100_000_000) * bitcoinRate.rate_float)
      )}
    </div>
  );
}
