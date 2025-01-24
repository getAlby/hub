import { useBitcoinRate } from "src/hooks/useBitcoinRate";

type FormattedFiatAmountProps = {
  amount: number;
  className?: string;
};

export default function FormattedFiatAmount({
  amount,
  className,
}: FormattedFiatAmountProps) {
  const { data: bitcoinRate } = useBitcoinRate();

  return (
    <div>
      {bitcoinRate && (
        <div className={className}>
          {new Intl.NumberFormat("en-US", {
            style: "currency",
            currency: "USD",
          }).format((amount / 100_000_000) * bitcoinRate.rate_float)}
        </div>
      )}
    </div>
  );
}
