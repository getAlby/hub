import { AlertTriangle } from "lucide-react";
import { Link } from "react-router-dom";
import Loading from "src/components/Loading";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { MIN_ALBY_BALANCE } from "src/constants";
import { useAlbyBalance } from "src/hooks/useAlbyBalance";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useInfo } from "src/hooks/useInfo";

export default function LightningOnboarding() {
  const { data: info } = useInfo();
  const { data: albyMe } = useAlbyMe();
  const { data: albyBalance } = useAlbyBalance();

  if (!info || !albyMe || !albyBalance) {
    return <Loading />;
  }

  return (
    <div className="flex flex-col gap-5 p-5">
      <div className="grid gap-2 text-center">
        <h1 className="text-2xl font-semibold">Open a Channel</h1>
        <p className="text-muted-foreground">
          You will now connect your node to the lightning network.
        </p>
      </div>
      <div className="flex flex-col items-center gap-5 justify-center max-w-md">
        {albyMe && albyBalance && albyBalance.sats >= MIN_ALBY_BALANCE && (
          <>
            {albyBalance.sats >= MIN_ALBY_BALANCE ? (
              <>
                <Card className="w-full">
                  <CardHeader>
                    <CardTitle>Alby Account Balance</CardTitle>
                  </CardHeader>
                  <CardContent>{albyBalance.sats} sats</CardContent>
                </Card>
                <Link to="migrate-alby">
                  <Button>Continue</Button>
                </Link>
              </>
            ) : (
              <>
                <Alert>
                  <AlertTriangle className="h-4 w-4" />
                  <AlertTitle>Not enough funds available!</AlertTitle>
                  <AlertDescription>
                    You don't have enough funds in your Alby account to fund a
                    new channel right now. Top up your Alby Account to proceed.
                  </AlertDescription>
                </Alert>
              </>
            )}
          </>
        )}

        {/* TODO: Enable this link as soon as we have the flow ready */}
        {/* <Link to="channels/new">
          <Button variant="link">Open a Channel manually</Button>
        </Link> */}
      </div>
    </div>
  );
}
