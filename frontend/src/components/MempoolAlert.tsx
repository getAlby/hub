import { TriangleAlertIcon } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { ExternalLinkButton } from "src/components/ui/button";
import { useMempoolApi } from "src/hooks/useMempoolApi";

export function MempoolAlert() {
  const { data: recommendedFees } = useMempoolApi<{ fastestFee: number }>(
    "/v1/fees/recommended",
    true
  );

  const fees = {
    "EXTREMELY HIGH": 200,
    "very high": 150,
    high: 100,
    "moderately high": 50,
  };

  const fee = recommendedFees?.fastestFee;
  if (!fee) {
    return null;
  }
  const matchedFee = Object.entries(fees).find((entry) => fee >= entry[1]);

  if (!matchedFee) {
    return null;
  }
  return (
    <Alert>
      <TriangleAlertIcon className="h-4 w-4" />
      <AlertTitle>
        Mempool Fees are currently{" "}
        <span className="font-semibold">{matchedFee[0]}</span>
      </AlertTitle>
      <AlertDescription>
        <p>Bitcoin transactions may be uneconomical at this time.</p>
        <div className="flex gap-2 mt-2">
          <ExternalLinkButton
            to="https://guides.getalby.com/user-guide/v/alby-account-and-browser-extension/alby-hub/faq-alby-hub/what-to-do-during-times-of-high-onchain-fees"
            size={"sm"}
          >
            Learn more
          </ExternalLinkButton>
          <ExternalLinkButton
            to="https://mempool.space"
            size={"sm"}
            variant="secondary"
          >
            View fees on mempool
          </ExternalLinkButton>
        </div>
      </AlertDescription>
    </Alert>
  );
}
