import {
  ArrowDownToDot,
  ArrowUpFromDot,
  CircleDot,
  CopyIcon,
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
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "src/components/ui/dropdown-menu";
import { useToast } from "src/components/ui/use-toast";
import { useBalances } from "src/hooks/useBalances";
import { useInfo } from "src/hooks/useInfo";
import { useNodeConnectionInfo } from "src/hooks/useNodeConnectionInfo";
import { copyToClipboard } from "src/lib/clipboard";

function Wallet() {
  const { data: info } = useInfo();
  const { data: balances } = useBalances();
  const { toast } = useToast();
  const { data: nodeConnectionInfo } = useNodeConnectionInfo();

  if (!info || !balances) {
    return <Loading />;
  }

  const isWalletUsable =
    balances.lightning.totalReceivable > 0 ||
    balances.lightning.totalSpendable > 0;

  return (
    <>
      <AppHeader
        title="Wallet"
        description="Send and receive transactions"
        contentRight={
          isWalletUsable && (
            <>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="outline" size="default">
                    <CircleDot className="mr-2 h-4 w-4 text-primary" />
                    Online
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent className="w-72" align="end">
                  <DropdownMenuItem>
                    <div className="flex flex-row gap-10 items-center w-full">
                      <div className="whitespace-nowrap flex flex-row items-center gap-2">
                        Node
                      </div>
                      <div className="overflow-hidden text-ellipsis">
                        {/* TODO: replace with skeleton loader */}
                        {nodeConnectionInfo?.pubkey || "Loading..."}
                      </div>
                      {nodeConnectionInfo && (
                        <CopyIcon
                          className="shrink-0 w-4 h-4"
                          onClick={() => {
                            copyToClipboard(nodeConnectionInfo.pubkey);
                            toast({ title: "Copied to clipboard." });
                          }}
                        />
                      )}
                    </div>
                  </DropdownMenuItem>
                  <>
                    <DropdownMenuSeparator />
                    <DropdownMenuGroup>
                      <DropdownMenuLabel>Liquidity</DropdownMenuLabel>
                      <DropdownMenuItem>
                        <div className="flex flex-row gap-3 items-center justify-between w-full">
                          <div className="grid grid-flow-col gap-2 items-center">
                            <ArrowDownToDot className="w-4 h-4 " />
                            Receiving Capacity
                          </div>
                          <div className="text-muted-foreground text-right">
                            {new Intl.NumberFormat().format(
                              Math.floor(
                                balances.lightning.totalReceivable / 1000
                              )
                            )}{" "}
                            sats
                          </div>
                        </div>
                      </DropdownMenuItem>
                      <DropdownMenuItem>
                        <div className="flex flex-row gap-3 items-center justify-between w-full">
                          <div className="grid grid-flow-col gap-2 items-center">
                            <ArrowUpFromDot className="w-4 h-4 " />
                            Spending Balance
                          </div>
                          <div className="text-muted-foreground text-right">
                            {new Intl.NumberFormat().format(
                              Math.floor(
                                balances.lightning.totalSpendable / 1000
                              )
                            )}{" "}
                            sats
                          </div>
                        </div>
                      </DropdownMenuItem>
                    </DropdownMenuGroup>
                  </>
                </DropdownMenuContent>
              </DropdownMenu>
            </>
          )
        }
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
