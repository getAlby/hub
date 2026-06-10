import { AlertTriangleIcon, ExternalLinkIcon, PencilIcon } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import ExternalLink from "src/components/ExternalLink";
import Loading from "src/components/Loading";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { useInfo } from "src/hooks/useInfo";
import { useMempoolApi } from "src/hooks/useMempoolApi";

type RecommendedFees = {
  fastestFee: number;
  halfHourFee: number;
  economyFee: number;
  minimumFee: number;
};

type FeeRateFieldProps = {
  feeRate: string;
  onFeeRateChange: (value: string) => void;
};

export function FeeRateField({ feeRate, onFeeRateChange }: FeeRateFieldProps) {
  const { data: info } = useInfo();
  const { data: recommendedFees, error: mempoolError } =
    useMempoolApi<RecommendedFees>("/v1/fees/recommended");
  const [isEditing, setIsEditing] = useState(false);
  const hasInitializedDefaultFee = useRef(false);

  useEffect(() => {
    if (
      recommendedFees?.fastestFee &&
      !hasInitializedDefaultFee.current &&
      !feeRate
    ) {
      hasInitializedDefaultFee.current = true;
      onFeeRateChange(recommendedFees.fastestFee.toString());
    }
  }, [feeRate, onFeeRateChange, recommendedFees]);

  useEffect(() => {
    if (mempoolError) {
      setIsEditing(true);
    }
  }, [mempoolError]);

  if (!info || (!recommendedFees && !mempoolError)) {
    return (
      <div className="flex items-center justify-between">
        <Label>On-chain Fee Rate (sat/vB)</Label>
        <Loading className="w-4 h-4" />
      </div>
    );
  }

  if (!isEditing) {
    return (
      <div className="flex items-center justify-between">
        <Label>On-chain Fee Rate (sat/vB)</Label>
        {feeRate ? (
          <button
            type="button"
            className="flex items-center gap-2 cursor-pointer"
            onClick={() => setIsEditing(true)}
          >
            <p className="text-sm">{feeRate} sat/vB</p>
            <PencilIcon className="w-4 h-4" />
          </button>
        ) : (
          <Loading className="w-4 h-4" />
        )}
      </div>
    );
  }

  return (
    <div className="grid gap-2">
      <Label htmlFor="fee-rate">On-chain Fee Rate (sat/vB)</Label>
      {mempoolError && (
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
          {info.mempoolUrl && (
            <ExternalLink
              to={info.mempoolUrl}
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
