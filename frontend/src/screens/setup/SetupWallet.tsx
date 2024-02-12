import {
  CaretRightIcon,
  WalletIcon,
} from "@bitcoin-design/bitcoin-icons-react/outline";
import { Link } from "react-router-dom";

import { useInfo } from "src/hooks/useInfo";
import Container from "src/components/Container";
import Alert from "src/components/Alert";

export function SetupWallet() {
  const { data: info } = useInfo();

  return (
    <>
      <Container>
        <p className="text-center font-light text-md leading-relaxed dark:text-neutral-400 px-4 mb-4">
          Connect Wallet to NWC
        </p>
        {info?.setupCompleted && (
          <Alert type="warn">
            ⚠️ Your node is already setup! only continue if you actually want to
            change your connection settings.
          </Alert>
        )}
        <div className="w-full mt-4">
          <WalletComponent walletType="new" />
          <WalletComponent walletType="import" />
        </div>
      </Container>
    </>
  );
}

function WalletComponent({ walletType }: { walletType: string }) {
  return (
    <Link
      to={`/setup/node?wallet=${walletType}`}
      className="shadow rounded-md p-4 bg-white dark:bg-surface-01dp hover:bg-gray-50 dark:hover:bg-surface-02dp text-gray-800 dark:text-neutral-200 cursor-pointer flex flex-row items-center gap-3 mb-4"
    >
      <div
        className={`flex-shrink-0 flex justify-center md:p-1 rounded ${
          walletType == "new" ? "bg-amber-100" : "bg-violet-100"
        }`}
      >
        <WalletIcon
          className={`w-8 ${
            walletType == "new" ? "text-amber-500" : "text-violet-500"
          }`}
        />
      </div>
      <div className="flex-grow">
        <div className="font-medium leading-5 text-sm md:text-base capitalize">
          {walletType} Wallet
        </div>
        <div className="text-gray-600 dark:text-neutral-400 text-xs leading-4 md:text-sm">
          {walletType == "new" ? "Breez only" : "Breez or LND"}
        </div>
      </div>
      <div className="flex-shrink-0 flex justify-end ">
        <CaretRightIcon className="w-8" />
      </div>
    </Link>
  );
}
