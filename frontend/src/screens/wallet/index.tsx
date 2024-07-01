import {
  ArrowDown,
  ArrowDownIcon,
  ArrowUp,
  ArrowUpIcon,
  CreditCard,
} from "lucide-react";
import { Link } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import BreezRedeem from "src/components/BreezRedeem";
import ExternalLink from "src/components/ExternalLink";
import Loading from "src/components/Loading";
import TransactionsList from "src/components/TransactionsList";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { useBalances } from "src/hooks/useBalances";
import { useInfo } from "src/hooks/useInfo";

function Wallet() {
  const { data: info, hasChannelManagement } = useInfo();
  const { data: balances } = useBalances();

  if (!info || !balances) {
    return <Loading />;
  }

  return (
    <>
      <AppHeader title="Wallet" description="Send and receive transactions" />
      <BreezRedeem />
      <div className="flex flex-col lg:flex-row justify-between lg:items-center gap-5">
        <div className="text-5xl font-semibold">
          {new Intl.NumberFormat().format(
            Math.floor(balances.lightning.totalSpendable / 1000)
          )}{" "}
          sats
        </div>
        <div className="flex items-center gap-4">
          <ExternalLink to="https://www.getalby.com/topup">
            <Button size="lg" variant="secondary">
              <CreditCard className="h-4 w-4 mr-2" />
              Buy Bitcoin
            </Button>
          </ExternalLink>
          <Link to="/wallet/receive">
            <Button size="lg">
              <ArrowDownIcon className="h-4 w-4 mr-2" />
              Receive
            </Button>
          </Link>
          <Link to="/wallet/send">
            <Button size="lg">
              <ArrowUpIcon className="h-4 w-4 mr-2" />
              Send
            </Button>
          </Link>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-5">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Spending Balance
            </CardTitle>
            <ArrowUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            {!balances && (
              <div>
                <div className="animate-pulse d-inline ">
                  <div className="h-2.5 bg-primary rounded-full w-12 my-2"></div>
                </div>
              </div>
            )}
            {balances && (
              <div className="text-2xl font-bold">
                {new Intl.NumberFormat(undefined, {}).format(
                  Math.floor(balances.lightning.totalSpendable / 1000)
                )}{" "}
                sats
              </div>
            )}
          </CardContent>
          {hasChannelManagement && (
            <CardFooter className="flex justify-end">
              <Link to="/channels/outgoing">
                <Button variant="outline">Top Up</Button>
              </Link>
            </CardFooter>
          )}
        </Card>
        {hasChannelManagement && (
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                Receiving Capacity
              </CardTitle>
              <ArrowDown className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              {!balances && (
                <div>
                  <div className="animate-pulse d-inline ">
                    <div className="h-2.5 bg-primary rounded-full w-12 my-2"></div>
                  </div>
                </div>
              )}
              <div className="text-2xl font-bold">
                {balances && (
                  <>
                    {new Intl.NumberFormat().format(
                      Math.floor(balances.lightning.totalReceivable / 1000)
                    )}{" "}
                    sats
                  </>
                )}
              </div>
            </CardContent>
            <CardFooter className="flex justify-end">
              <Link to="/channels/incoming">
                <Button variant="outline">Increase</Button>
              </Link>
            </CardFooter>
          </Card>
        )}
      </div>

      <TransactionsList />
    </>
  );
}

export default Wallet;
