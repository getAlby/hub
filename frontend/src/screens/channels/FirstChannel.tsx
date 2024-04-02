import { Crown } from "lucide-react";
import { Link } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import Loading from "src/components/Loading";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { Button } from "src/components/ui/button";
import { MIN_ALBY_BALANCE } from "src/constants";
import { useAlbyBalance } from "src/hooks/useAlbyBalance";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useInfo } from "src/hooks/useInfo";

export default function FirstChannel() {
  const { data: info } = useInfo();
  const { data: albyMe } = useAlbyMe();
  const { data: albyBalance } = useAlbyBalance();

  if (!info) {
    return <Loading />;
  }

  return (
    <>
      <AppHeader
        title="Connect to the lightning network"
        description="Get started and create your first channel in no time"
      />
      <div className="flex flex-col gap-4">
        <Alert>
          <Crown className="h-4 w-4" />
          <AlertTitle>Your Alby has grown up now</AlertTitle>
          <AlertDescription>
            You have your own node and you also need your own channels to send
            and receive payments on the lightning network.
          </AlertDescription>
        </Alert>

        {!albyMe && (
          <>
            <p className="mb-5">
              If you have funds on your Alby account you can use them to open
              your first channel.
            </p>

            <div className="flex flex-row items-center justify-center gap-3">
              <Link to={info.albyAuthUrl}>
                <Button>Connect your Alby Account</Button>
              </Link>
              <Link to="/channels/new">
                <Button variant="ghost">Open manually</Button>
              </Link>
            </div>
          </>
        )}
        {albyMe && albyBalance && (
          <div className="mt-8 border-2 p-8 rounded-lg flex flex-col justify-center items-center border-yellow-300 bg-yellow-100">
            <p className="mb-8">
              Logged in as{" "}
              <span className="font-bold">{albyMe.lightning_address}</span>
            </p>

            {albyBalance.sats >= MIN_ALBY_BALANCE && (
              <>
                <Link to="/channels/migrate-alby">
                  <button className="bg-yellow-400 border-8 rounded-lg border-yellow-500 p-4 shadow-lg font-mono text-lg font-black">
                    Migrate Funds 🚀
                  </button>
                </Link>
                <p className="text-sm italic mt-4">
                  You have {albyBalance.sats} sats to migrate
                </p>
              </>
            )}
            {albyBalance.sats < MIN_ALBY_BALANCE && (
              <>
                <p>
                  You don't have enough sats in your Alby account to open a
                  channel.
                </p>
              </>
            )}
          </div>
        )}
      </div>
    </>
  );
}
