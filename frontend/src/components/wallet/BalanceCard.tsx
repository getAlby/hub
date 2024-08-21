import { ArrowUp } from "lucide-react";
import { Link } from "react-router-dom";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { BalancesResponse } from "src/types";

type BalanceCardProps = {
  balances: BalancesResponse;
  hasChannelManagement: boolean;
  isSpending?: boolean;
};

function BalanceCard({
  balances,
  hasChannelManagement,
  isSpending,
}: BalanceCardProps) {
  return (
    <Card className="w-full hidden md:block self-start">
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">
          {isSpending ? "Spending Balance" : "Receiving Capacity"}
        </CardTitle>
        <ArrowUp className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        {!balances && (
          <div>
            <div className="animate-pulse d-inline ">
              <div className="h-2.5 bg-primary rounded-full w-12 my-2"></div>
            </div>
          </div>
        )}
        {balances && (
          <div className="text-2xl font-bold balance sensitive">
            {new Intl.NumberFormat(undefined, {}).format(
              Math.floor(
                isSpending
                  ? balances.lightning.totalSpendable / 1000
                  : balances.lightning.totalReceivable / 1000
              )
            )}{" "}
            sats
          </div>
        )}
      </CardContent>
      {hasChannelManagement && (
        <CardFooter className="flex justify-end">
          <Link to={isSpending ? "/channels/outgoing" : "/channels/incoming"}>
            <Button variant="outline">
              {isSpending ? "Top Up" : "Increase"}
            </Button>
          </Link>
        </CardFooter>
      )}
    </Card>
  );
}

export default BalanceCard;
