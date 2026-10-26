import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { useTransactionStats } from "src/hooks/useTransactionStats";

function formatFeeRate(rate: number): string {
  if (rate > 0 && rate < 0.01) {
    return "<0.01%";
  }
  return `${rate.toFixed(2)}%`;
}

export function FeeRateWidget() {
  const { data: stats } = useTransactionStats();

  // Only show once there's payment volume to talk about — the volume-weighted
  // rate is meaningless (and divides by zero) before the first payment.
  if (!stats || !stats.totalVolumeMsat || !stats.numPayments) {
    return null;
  }

  const feeRate = (stats.totalFeesPaidMsat / stats.totalVolumeMsat) * 100;

  return (
    <Card>
      <CardHeader>
        <CardTitle>Lightning fees</CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-muted-foreground text-xs">Average fee rate</p>
        <p className="text-3xl font-semibold">{formatFeeRate(feeRate)}</p>
        <p className="text-muted-foreground text-sm mt-3">
          You've sent{" "}
          <FormattedBitcoinAmount amountMsat={stats.totalVolumeMsat} /> across{" "}
          {stats.numPayments} payment{stats.numPayments === 1 ? "" : "s"} and
          paid only{" "}
          <FormattedBitcoinAmount amountMsat={stats.totalFeesPaidMsat} /> in
          fees.
        </p>
      </CardContent>
    </Card>
  );
}
