import { ArrowUpIcon, InfoIcon } from "lucide-react";
import { Outlet } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import BalanceCard from "src/components/BalanceCard";
import Loading from "src/components/Loading";
import { useBalances } from "src/hooks/useBalances";
import { useChannels } from "src/hooks/useChannels";
import { useTransactions } from "src/hooks/useTransactions";

import dayjs from "dayjs";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { useInfo } from "src/hooks/useInfo";
import { LinkButton } from "../ui/custom/link-button";

export default function SendLayout() {
  const { hasChannelManagement } = useInfo();
  const { data: balances } = useBalances();
  const { data: channels } = useChannels();
  const { data: transactionData } = useTransactions();

  if (!balances || !channels || !transactionData) {
    return <Loading />;
  }

  return (
    <div className="grid gap-5">
      <AppHeader title="Send" />
      {transactionData.transactions.some(
        (tx) =>
          tx.state === "pending" &&
          dayjs().diff(dayjs(tx.createdAt)) <
            1000 * 60 * 60 * 24 /* payment pending in last 24h */
        /* TODO: remove diff check when expired transactions are marked as failed */
      ) && (
        <Alert>
          <InfoIcon />
          <AlertTitle>Pending Payment</AlertTitle>
          <AlertDescription>
            You have one or more payments that have not settled.
            <LinkButton to={"/wallet"} size={"sm"}>
              View Payments
            </LinkButton>
          </AlertDescription>
        </Alert>
      )}
      <div className="flex gap-12 w-full">
        <div className="w-full max-w-lg">
          <Outlet />
        </div>
        <BalanceCard
          balance={balances.lightning.totalSpendable}
          title="Spending Balance"
          buttonTitle="Top Up"
          buttonLink="/channels/outgoing"
          BalanceCardIcon={ArrowUpIcon}
          hasChannelManagement={!!hasChannelManagement}
        />
      </div>
    </div>
  );
}
