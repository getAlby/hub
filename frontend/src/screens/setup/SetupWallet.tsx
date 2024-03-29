import { PopiconsChevronRightLine, PopiconsWalletLine } from "@popicons/react";
import React from "react";
import { Link } from "react-router-dom";

import Container from "src/components/Container";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { useInfo } from "src/hooks/useInfo";

export function SetupWallet() {
  const { data: info } = useInfo();
  const [showOtherOptions, setShowOtherOptions] = React.useState(false);
  return (
    <>
      <Container>
        <div className="grid gap-5">
          <div className="grid gap-2 text-center">
            <h1 className="font-semibold text-2xl font-headline">
              Create Your Wallet
            </h1>
            <p className="text-muted-foreground">
              Alby Hub requires a wallet to connect to your apps.
            </p>
          </div>
          {info?.backendType && (
            <>
              <Link
                to={`/setup/finish`}
                className="shadow rounded-md p-4 bg-white dark:bg-surface-01dp hover:bg-gray-50 dark:hover:bg-surface-02dp text-gray-800 dark:text-neutral-200 cursor-pointer flex flex-row items-center gap-3 w-full"
              >
                <div
                  className={`flex-shrink-0 flex justify-center md:p-1 rounded ${"bg-green-100"}`}
                >
                  <PopiconsWalletLine className="w-8 h-8 p-1 text-green-500" />
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
                  <PopiconsChevronRightLine className="w-8 h-8 text-gray-400" />
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

function WalletComponent({ walletType }: { walletType: string }) {
  return (
    <>
      <Link
        to={`/setup/node?wallet=${walletType}`}
        className="shadow rounded-md p-4 bg-white dark:bg-surface-01dp hover:bg-gray-50 dark:hover:bg-surface-02dp text-gray-800 dark:text-neutral-200 cursor-pointer flex flex-row items-center gap-3 mb-4"
      >
        <div
          className={`flex-shrink-0 flex justify-center md:p-1 rounded ${
            walletType == "new" ? "bg-amber-100" : "bg-violet-100"
          }`}
        >
          <PopiconsWalletLine
            className={`w-8 h-8 p-1 ${
              walletType == "new" ? "text-amber-500" : "text-violet-500"
            }`}
          />
        </div>
        <div className="flex-grow">
          <div className="font-medium leading-5 text-sm md:text-base capitalize">
            {walletType} Wallet
          </div>
          <div className="text-gray-600 dark:text-neutral-400 text-xs leading-4 md:text-sm">
            {walletType == "new"
              ? "Create a new wallet powered by the Breez SDK"
              : "Connect to an existing Breez or LND wallet"}
          </div>
        </div>
        <div className="flex-shrink-0 flex justify-end ">
          <PopiconsChevronRightLine className="w-8 h-8 text-gray-400" />
        </div>
      </Link>

      <Link to={`/setup/node?wallet=${walletType}`}>
        <Card>
          <CardHeader>
            <CardTitle>test</CardTitle>
          </CardHeader>
          <CardContent>
            <div
              className={`flex-shrink-0 flex justify-center md:p-1 rounded ${
                walletType == "new" ? "bg-amber-100" : "bg-violet-100"
              }`}
            >
              <PopiconsWalletLine
                className={`w-8 h-8 p-1 ${
                  walletType == "new" ? "text-amber-500" : "text-violet-500"
                }`}
              />
            </div>
            <div className="flex-grow">
              <div className="font-medium leading-5 text-sm md:text-base capitalize">
                {walletType} Wallet
              </div>
              <div className="text-gray-600 dark:text-neutral-400 text-xs leading-4 md:text-sm">
                {walletType == "new"
                  ? "Create a new wallet powered by the Breez SDK"
                  : "Connect to an existing Breez or LND wallet"}
              </div>
            </div>
            <div className="flex-shrink-0 flex justify-end ">
              <PopiconsChevronRightLine className="w-8 h-8 text-gray-400" />
            </div>
          </CardContent>
        </Card>
      </Link>
    </>
  );
}
