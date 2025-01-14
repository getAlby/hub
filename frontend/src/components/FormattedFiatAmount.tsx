import { useBitcoinRate } from "src/hooks/useBitcoinRate";

type FormattedFiatAmountProps = {
  amount: number;
};

export default function FormattedFiatAmount({
  amount,
}: FormattedFiatAmountProps) {
  const { data: bitcoinRate } = useBitcoinRate();

  return (
    <div>
      <div className="text-2xl font-bold balance sensitive">
        {new Intl.NumberFormat().format(Math.floor(amount))} sats
      </div>
      {bitcoinRate && (
        <div className="text-sm text-muted-foreground">
          {new Intl.NumberFormat("en-US", {
            style: "currency",
            currency: "USD",
          }).format((amount / 100_000_000) * bitcoinRate.rate_float)}
        </div>
      )}
    </div>
  );
}
