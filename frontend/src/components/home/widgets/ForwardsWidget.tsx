import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { ExternalLinkButton } from "src/components/ui/custom/external-link-button";
import { useForwards } from "src/hooks/useForwards";

export function ForwardsWidget() {
  const { data: forwards } = useForwards();

  if (!forwards) {
    return null;
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Routing</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-2 gap-6">
          <div>
            <p className="text-muted-foreground text-xs">Fees Earned</p>
            <p className="text-xl font-semibold">
              <FormattedBitcoinAmount amount={forwards.totalFeeEarnedMsat} />
              <FormattedFiatAmount
                amount={Math.floor(forwards.totalFeeEarnedMsat / 1000)}
              />
            </p>
          </div>
          <div>
            <p className="text-muted-foreground text-xs">Total Routed</p>
            <p className="text-xl font-semibold">
              <FormattedBitcoinAmount
                amount={forwards.outboundAmountForwardedMsat}
              />
              <FormattedFiatAmount
                amount={Math.floor(forwards.outboundAmountForwardedMsat / 1000)}
              />
            </p>
          </div>
          <div>
            <p className="text-muted-foreground text-xs">Transactions Routed</p>
            <p className="text-xl font-semibold">{forwards.numForwards}</p>
          </div>
        </div>
        <div className="flex justify-end mt-4 items-end">
          <p className="text-muted-foreground text-sm">
            Earn and support the lightning network by routing payments. To route
            payments you need public channels and set competitive fees.
          </p>
          <ExternalLinkButton
            variant="secondary"
            to="https://guides.getalby.com/user-guide/alby-hub/faq/how-to-earn-routing-fees"
          >
            Learn more
          </ExternalLinkButton>
        </div>
      </CardContent>
    </Card>
  );
}
