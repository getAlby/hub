import {
  ExternalLink
} from "lucide-react";
import { Link } from "react-router-dom";
import AlbyHead from "src/assets/images/alby-head.svg";
import AppHeader from "src/components/AppHeader";
import BreezRedeem from "src/components/BreezRedeem";
import Loading from "src/components/Loading";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { useBalances } from "src/hooks/useBalances";
import { useInfo } from "src/hooks/useInfo";

function Wallet() {
  const { data: info } = useInfo();
  const { data: balances } = useBalances();

  if (!info || !balances) {
    return <Loading />;
  }

  return (
    <>
      <AppHeader
        title="Wallet"
        description="Send and receive transactions"
      />

      <div className="flex flex-col lg:flex-row justify-between items-start lg:items-end gap-5">
        <div className="text-5xl font-semibold">
          {new Intl.NumberFormat().format(
            Math.floor(balances.lightning.totalSpendable / 1000)
          )}{" "}
          sats
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-5">
        <Link to={`https://www.getalby.com/dashboard`} target="_blank">
          <Card>
            <CardHeader>
              <div className="flex flex-row items-center">
                <img
                  src={AlbyHead}
                  className="w-12 h-12 rounded-xl p-1 border"
                />
                <div>
                  <CardTitle>
                    <div className="flex-1 leading-5 font-semibold text-xl whitespace-nowrap text-ellipsis overflow-hidden ml-4">
                      Alby Web
                    </div>
                  </CardTitle>
                  <CardDescription className="ml-4">
                    Install Alby Web on your phone and use your Hub on the go.
                  </CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent className="text-right">
              <Button variant="outline">
                Open Alby Web
                <ExternalLink className="w-4 h-4 ml-2" />
              </Button>
            </CardContent>
          </Card>
        </Link>
        <Link to={`https://www.getalby.com`} target="_blank">
          <Card>
            <CardHeader>
              <div className="flex flex-row items-center">
                <img
                  src={AlbyHead}
                  className="w-12 h-12 rounded-xl p-1 border bg-[#FFDF6F]"
                />
                <div>
                  <CardTitle>
                    <div className="flex-1 leading-5 font-semibold text-xl whitespace-nowrap text-ellipsis overflow-hidden ml-4">
                      Alby Browser Extension
                    </div>
                  </CardTitle>
                  <CardDescription className="ml-4">
                    Best to use when using your favourite Internet browser
                  </CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent className="text-right">
              <Button variant="outline">
                Install Alby Extension
                <ExternalLink className="w-4 h-4 ml-2" />
              </Button>
            </CardContent>
          </Card>
        </Link>
      </div>

      <BreezRedeem />
    </>
  );
}

export default Wallet;
