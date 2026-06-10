import { AlertTriangleIcon, ExternalLinkIcon } from "lucide-react";
import ExternalLink from "src/components/ExternalLink";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";

type RecommendedFees = {
  fastestFee: number;
  halfHourFee: number;
  economyFee: number;
  minimumFee: number;
};

type FeeRateFieldProps = {
  feeRate: string;
  onFeeRateChange: (value: string) => void;
  recommendedFees?: RecommendedFees;
  hasMempoolError?: boolean;
  mempoolUrl?: string;
};

export function FeeRateField({
  feeRate,
  onFeeRateChange,
  recommendedFees,
  hasMempoolError,
  mempoolUrl,
}: FeeRateFieldProps) {
  return (
    <div className="grid gap-2">
      <Label htmlFor="fee-rate">On-chain Fee Rate (sat/vB)</Label>
      {hasMempoolError && (
        <div className="text-muted-foreground text-xs flex gap-1 items-center">
          <AlertTriangleIcon className="h-3 w-3" />
          Failed to fetch fee estimates. Try refreshing the page.
        </div>
      )}
      <Input
        id="fee-rate"
        type="number"
        value={feeRate}
        step={1}
        required
        min={recommendedFees?.minimumFee || 1}
        onChange={(e) => {
          onFeeRateChange(e.target.value);
        }}
      />
      {recommendedFees && (
        <div className="flex items-center mt-2 gap-4">
          <Button
            variant="positive"
            className="rounded-full"
            type="button"
            onClick={() =>
              onFeeRateChange(recommendedFees.economyFee.toString())
            }
          >
            Low priority: {recommendedFees.economyFee}
          </Button>
          <Button
            variant="positive"
            className="rounded-full"
            type="button"
            onClick={() =>
              onFeeRateChange(recommendedFees.fastestFee.toString())
            }
          >
            High priority: {recommendedFees.fastestFee}
          </Button>
          {mempoolUrl && (
            <ExternalLink
              to={mempoolUrl}
              className="text-muted-foreground text-sm underline flex items-center gap-2"
            >
              View on Mempool
              <ExternalLinkIcon className="w-4 h-4" />
            </ExternalLink>
          )}
        </div>
      )}
    </div>
  );
}
