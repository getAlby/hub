import clsx from "clsx";
import { useBitcoinRate } from "src/hooks/useBitcoinRate";
import { useInfo } from "src/hooks/useInfo";

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

  if (!bitcoinRate) {
    return (
      <div className="animate-pulse h-2.5 bg-muted-foreground rounded-full w-16 my-1 inline-block"></div>
    );
  }

  return (
    <div className={clsx("text-sm text-muted-foreground", className)}>
      {new Intl.NumberFormat("en-US", {
        style: "currency",
        currency: info?.currency || "usd",
      }).format((amount / 100_000_000) * bitcoinRate.rate_float)}
    </div>
  );
}
