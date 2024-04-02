import { Bird, Crown, Landmark, Rss } from "lucide-react";
import { Link } from "react-router-dom";
import Loading from "src/components/Loading";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import {
  Card,
  CardContent,
  CardDescription,
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
    <div className="flex flex-col gap-4 p-4">
      <Alert>
        <Crown className="h-4 w-4" />
        <AlertTitle>Your Alby has grown up now</AlertTitle>
        <AlertDescription>
          You're running your own node! Now let's connect to the lightning
          network.
        </AlertDescription>
      </Alert>

      <div className="grid sm:grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
        {albyMe && albyBalance && albyBalance.sats >= MIN_ALBY_BALANCE && (
          <>
            <Link to="migrate-alby">
              <Card>
                <CardHeader>
                  <CardTitle>
                    <div className="flex flex-row items-center gap-2">
                      <Bird className="w-10 h-10" />
                      Migrate
                    </div>
                  </CardTitle>
                  <CardDescription>
                    Use your existing Alby account funds to open a channel to
                    Alby on the lightning network.
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  You have {albyBalance.sats} sats to migrate
                </CardContent>
              </Card>
            </Link>
          </>
        )}
        <Link to="channels/new" className="h-full">
          <Card className="h-full">
            <CardHeader>
              <CardTitle>
                <div className="flex flex-row items-center gap-2">
                  <Rss className="w-10 h-10" />
                  Open Channel
                </div>
              </CardTitle>
              <CardDescription>
                Deposit Bitcoin or pay with lightning to open a channel on the
                lightning network.
              </CardDescription>
            </CardHeader>
          </Card>
        </Link>
        <Card className="cursor-not-allowed">
          <CardHeader>
            <CardTitle>
              <div className="flex flex-row items-center gap-2">
                <Landmark className="w-10 h-10" />
                Connect to a Mint
              </div>
            </CardTitle>
            <CardDescription>
              Connect to a custodial Cashu mint to quickly get started with Alby
              Hub.
            </CardDescription>
          </CardHeader>
        </Card>
      </div>
    </div>
  );
}
