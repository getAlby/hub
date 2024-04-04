import { ChevronRight, WalletMinimal } from "lucide-react";
import React from "react";
import { Link } from "react-router-dom";

import Container from "src/components/Container";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { Button } from "src/components/ui/button";
import { useInfo } from "src/hooks/useInfo";

export function SetupWallet() {
  const { data: info } = useInfo();
  const [showOtherOptions, setShowOtherOptions] = React.useState(false);
  return (
    <>
      <Container>
        <div className="grid gap-5">
          <TwoColumnLayoutHeader
            title="Create Your Wallet"
            description="Alby Hub requires a wallet to connect to your apps."
          />
          {info?.backendType && (
            <>
              <Link
                to="setup/finish"
                className="shadow rounded-md p-4 bg-white dark:bg-surface-01dp hover:bg-gray-50 dark:hover:bg-surface-02dp text-gray-800 dark:text-neutral-200 cursor-pointer flex flex-row items-center gap-3 w-full"
              >
                <div
                  className={`flex-shrink-0 flex justify-center md:p-1 rounded ${"bg-green-100"}`}
                >
                  <WalletMinimal className="w-8 h-8 p-1 text-green-500" />
                </div>
                <div className="flex-grow">
                  <div className="font-medium leading-5 text-sm md:text-base capitalize">
                    {info.backendType} Wallet
                  </div>
                  <div className="text-gray-600 dark:text-neutral-400 text-xs leading-4 md:text-sm">
                    Connect to preconfigured {info.backendType} Wallet
                  </div>
                </div>
                <div className="flex-shrink-0 flex justify-end ">
                  <ChevronRight className="w-8 h-8 text-gray-400" />
                </div>
              </Link>

              {!showOtherOptions && (
                <button
                  className={`mt-4 items-center justify-center px-3 py-2 cursor-pointer duration-150 transition border dark:border-white/10 bg-white dark:bg-surface-02dp text-purple-700 dark:text-neutral-200 hover:bg-gray-50  dark:hover:bg-surface-16dp bg-origin-border rounded-md focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-purple-700`}
                  onClick={() => setShowOtherOptions(true)}
                >
                  See other options
                </button>
              )}
            </>
          )}

          {(showOtherOptions || !info?.backendType) && (
            <div className="grid gap-1.5 justify-center text-center">
              <Link to="/setup/node?wallet=new">
                <Button>New Wallet</Button>
              </Link>
              <Link to="/setup/node?wallet=import">
                <Button variant="link">Import Existing Wallet</Button>
              </Link>
            </div>
          )}
        </div>
      </Container>
    </>
  );
}
