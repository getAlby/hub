import { Link } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import { NostrWalletConnectIcon } from "src/components/icons/NostrWalletConnectIcon";
import { suggestedApps } from "src/components/SuggestedAppData";
import { Button } from "src/components/ui/button";

const BITREFILL_APP_ID = "bitrefill";

export function Bitrefill() {
  const app = suggestedApps.find((x) => x.id === BITREFILL_APP_ID);

  return (
    <div className="grid gap-5">
      <AppHeader
        title={
          <>
            <div className="flex flex-row items-center">
              <img src={app?.logo} className="w-14 h-14 rounded-lg mr-4" />
              <div className="flex flex-col">
                <div>{app?.title}</div>
                <div className="text-sm font-normal text-muted-foreground">
                  {app?.description}
                </div>
              </div>
            </div>
          </>
        }
        description=""
        contentRight={
          <Link to={`/apps/new?app=${app?.id}`}>
            <Button variant="outline">
              <NostrWalletConnectIcon className="w-4 h-4 mr-2" />
              Connect to {app?.title}
            </Button>
          </Link>
        }
      />
      <iframe
        className="w-full rounded-lg flex-1"
        width="375"
        height="667"
        src="https://embed.bitrefill.com/?showPaymentInfo=true"
        sandbox="allow-same-origin allow-popups allow-scripts allow-forms"
      ></iframe>
    </div>
  );
}
