import {
  CirclePlusIcon,
  HandCoins,
  HelpCircle,
  TriangleAlert,
  Wallet2,
} from "lucide-react";
import { Link } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import ResponsiveButton from "src/components/ResponsiveButton";
import { Button } from "src/components/ui/button";

import SubWalletDarkSVG from "public/images/illustrations/sub-wallet-dark.svg";
import SubWalletLightSVG from "public/images/illustrations/sub-wallet-light.svg";

export function SubwalletIntro() {
  return (
    <div className="grid gap-4">
      <AppHeader
        title="Sub-wallets"
        description="Create sub-wallets for yourself, friends, family or coworkers"
        contentRight={
          <>
            <ExternalLink to="https://guides.getalby.com/user-guide/alby-hub/sub-wallets">
              <Button variant="outline" size="icon">
                <HelpCircle className="w-4 h-4" />
              </Button>
            </ExternalLink>
            <Link to="/sub-wallets/new">
              <ResponsiveButton icon={CirclePlusIcon} text="New Sub-wallet" />
            </Link>
          </>
        }
      />
      <div>
        <div className="flex flex-col gap-6 max-w-(--breakpoint-md)">
          <div className="mb-2">
            <img src={SubWalletDarkSVG} className="w-72 hidden dark:block" />
            <img src={SubWalletLightSVG} className="w-72 dark:hidden" />
          </div>
          <div>
            <div className="flex flex-row gap-3">
              <Wallet2 className="w-6 h-6" />
              <div className="font-medium">
                Sub-wallets are separate wallets hosted by your Alby Hub
              </div>
            </div>
            <div className="ml-9 text-muted-foreground text-sm">
              Each sub-wallet has its own balance and can be used as a separate
              wallet that can be connected to Alby Account or any app.
            </div>
          </div>
          <div>
            <div className="flex flex-row gap-3">
              <HandCoins className="w-6 h-6" />
              <div className="font-medium">
                Sub-wallets depend on your Alby Hub spending balance and receive
                limit
              </div>
            </div>
            <div className="ml-9 text-muted-foreground text-sm">
              Sub-wallets are using your Hubs node liquidity. They can receive
              funds as long as you have enough receive limit in your channels.
            </div>
          </div>
          <div>
            <div className="flex flex-row gap-3">
              <TriangleAlert className="w-6 h-6" />
              <div className="font-medium">
                Be wary of spending sub-wallets funds
              </div>
            </div>
            <div className="ml-9 text-muted-foreground text-sm">
              Make sure you always maintain enough funds in your spending
              balance to prevent sub-wallets becoming unspendable. Sub-wallet
              payments might fail if the amount isn't available in your spending
              balance.
            </div>
          </div>
          <div>
            <Link to="/sub-wallets/new">
              <Button className="mt-4">Create Sub-wallet</Button>
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}
