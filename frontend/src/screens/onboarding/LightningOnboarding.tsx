import { Landmark, Rss, Send } from "lucide-react";
import { Link } from "react-router-dom";
import Loading from "src/components/Loading";
import { Button } from "src/components/ui/button";
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
    <div className="flex flex-col gap-5 p-5">
      <div className="grid gap-2 text-center">
        <h1 className="text-2xl font-semibold">Connect to Lightning</h1>
        <p className="text-muted-foreground">
          Choose how you want to connect to the lightning network.
        </p>
      </div>
      <div className="grid sm:grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4 items-stretch">
        {albyMe && albyBalance && albyBalance.sats >= MIN_ALBY_BALANCE && (
          <>
            <Link to="migrate-alby" className="h-full">
              <Card className="h-full">
                <CardHeader>
                  <CardTitle>
                    <div className="flex flex-row items-center gap-3">
                      <Send className="w-6 h-6" />
                      Migrate
                    </div>
                  </CardTitle>
                  <CardDescription>
                    Use your existing Alby account funds to open a channel to
                    Alby on the lightning network.
                  </CardDescription>
                </CardHeader>
                <CardContent className="flex flex-col items-center">
                  <Button>Migrate {albyBalance.sats} sats</Button>
                </CardContent>
              </Card>
            </Link>
          </>
        )}
        <Link to="channels/new" className="h-full">
          <Card className="h-full">
            <CardHeader>
              <CardTitle>
                <div className="flex flex-row items-center gap-3">
                  <Rss className="w-6 h-6" />
                  Open Channel
                </div>
              </CardTitle>
              <CardDescription>
                Deposit Bitcoin or pay with lightning to open a channel on the
                lightning network.
              </CardDescription>
            </CardHeader>
            <CardContent className="flex flex-col items-center">
              <Button variant="secondary">Open Channel</Button>
            </CardContent>
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
          <CardContent className="flex flex-col items-center">
            <Button variant="secondary" disabled>
              Choose Mint
            </Button>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
