import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { useMempoolApi } from "src/hooks/useMempoolApi";

export function OnchainFeesWidget() {
  const { data: recommendedFees } = useMempoolApi<{
    fastestFee: number;
    halfHourFee: number;
    economyFee: number;
    minimumFee: number;
  }>("/v1/fees/recommended");

  if (!recommendedFees) {
    return null;
  }

  const entries = [
    {
      title: "No priority",
      value: recommendedFees.minimumFee,
    },
    {
      title: "Low priority",
      value: recommendedFees.economyFee,
    },
    {
      title: "Medium priority",
      value: recommendedFees.halfHourFee,
    },
    {
      title: "High priority",
      value: recommendedFees.fastestFee,
    },
  ];

  return (
    <Card>
      <CardHeader>
        <CardTitle>On-Chain Fees</CardTitle>
      </CardHeader>
      <CardContent className="grid grid-cols-2 gap-6">
        {entries.map((entry) => (
          <div key={entry.title}>
            <p className="text-muted-foreground text-xs">{entry.title}</p>
            <p className="text-xl font-semibold">{entry.value} sat/vB</p>
          </div>
        ))}
      </CardContent>
    </Card>
  );
}
