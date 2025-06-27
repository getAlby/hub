import { CircleCheckIcon } from "lucide-react";
import { Link, useLocation } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";

export default function AutoSwapSuccess() {
  const { state } = useLocation();

  return (
    <div className="grid gap-5">
      <AppHeader title="Auto Swap" />
      <div className="w-full max-w-lg">
        <Card className="w-full">
          <CardHeader>
            <CardTitle className="text-center">Auto swaps enabled</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col items-center gap-4">
            <CircleCheckIcon className="w-32 h-32 mb-2" />
            <div className="flex flex-col gap-2 items-center">
              <p className="text-xl font-bold slashed-zero">
                {new Intl.NumberFormat().format(state.amount)} sats
              </p>
              <FormattedFiatAmount amount={state.amount} />
              <div className="text-sm">
                Will be swapped everytime balance reaches{" "}
                <span className="font-bold slashed-zero">
                  {new Intl.NumberFormat().format(state.balanceThreshold)} sats
                </span>
              </div>
            </div>
          </CardContent>
        </Card>
        <Link to="/wallet">
          <Button className="mt-4 w-full" variant="secondary">
            Back To Wallet
          </Button>
        </Link>
      </div>
    </div>
  );
}
