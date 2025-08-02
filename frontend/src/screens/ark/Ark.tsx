import { InfoIcon, LinkIcon, Settings2Icon, ZapIcon } from "lucide-react";
import React, { useState } from "react";
import { Link } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import { HealthCheckAlert } from "src/components/channels/HealthcheckAlert";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import Loading from "src/components/Loading";
import ResponsiveButton from "src/components/ResponsiveButton";
import {
  Card,
  CardContent,
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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "src/components/ui/table";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "src/components/ui/tooltip";
import { useToast } from "src/components/ui/use-toast";
import { useBalances } from "src/hooks/useBalances";
import { cn } from "src/lib/utils";
import { request } from "src/utils/request";

type VTXO = {
  point: {
    Txid: string;
    Vout: number;
  };
  amount_sat: number;
  user_pubkey: string;
  asp_pubkey: string;
  expiry_height: number;
  is_arkoor: boolean;
};

export function Ark() {
  const { toast } = useToast();
  const { data: balances, mutate: reloadBalances } = useBalances();
  const [vtxos, setVtxos] = useState<VTXO[]>(); // TODO: use proper hook
  const executeCommand = React.useCallback(
    async function <T>(command: string) {
      try {
        if (!command) {
          throw new Error("No command set");
        }
        const result = await request("/api/command", {
          method: "POST",
          body: JSON.stringify({ command }),
          headers: {
            "Content-Type": "application/json",
          },
        });

        return result as T;
      } catch (error) {
        console.error(error);
        toast({
          variant: "destructive",
          title: "Something went wrong: " + error,
        });
      }
    },
    [toast]
  );

  const board = React.useCallback(() => {
    executeCommand("board");
  }, [executeCommand]);

  // TODO: when should this be executed?
  React.useEffect(() => {
    (async () => {
      await executeCommand("maintenance");
      reloadBalances();
    })();
  }, [executeCommand, reloadBalances]);

  React.useEffect(() => {
    (async () => {
      const response = await executeCommand<{ vtxos: VTXO[] }>("list_vtxos");
      if (!response) {
        console.error("failed to fetch vtxos");
        return;
      }
      console.info("VTXOs response", response);
      setVtxos(response.vtxos);
    })();
  }, [executeCommand]);

  return (
    <>
      <AppHeader
        title="Ark"
        contentRight={
          <div className="flex gap-3 items-center justify-center">
            <DropdownMenu modal={false}>
              <DropdownMenuTrigger>
                <ResponsiveButton
                  icon={Settings2Icon}
                  text="Advanced"
                  variant="outline"
                />
              </DropdownMenuTrigger>
              <DropdownMenuContent className="w-56">
                <DropdownMenuGroup>
                  <DropdownMenuLabel>On-Chain</DropdownMenuLabel>
                  <DropdownMenuItem>
                    <Link
                      to="/channels/onchain/deposit-bitcoin"
                      className="w-full"
                    >
                      Deposit Bitcoin
                    </Link>
                  </DropdownMenuItem>
                  <DropdownMenuItem>
                    <div
                      className="flex flex-row gap-4 items-center w-full cursor-pointer"
                      onClick={() => {
                        board();
                      }}
                    >
                      Board
                    </div>
                  </DropdownMenuItem>
                  <DropdownMenuSeparator />
                  <DropdownMenuLabel>Off-Chain</DropdownMenuLabel>
                  <DropdownMenuItem>
                    <Link
                      to="#"
                      className="w-full"
                      onClick={async () => {
                        const result = await executeCommand<{
                          address: string;
                        }>("new_address");
                        prompt("Copy ark address", result?.address);
                      }}
                    >
                      Receive via Ark
                    </Link>
                  </DropdownMenuItem>
                </DropdownMenuGroup>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        }
      ></AppHeader>

      <HealthCheckAlert />

      <div
        className={cn("flex flex-col sm:flex-row flex-wrap gap-3 slashed-zero")}
      >
        <Card className="flex flex-1 sm:flex-[2] flex-col">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="font-semibold text-2xl">Off-Chain</CardTitle>
            <ZapIcon className="h-6 w-6 text-muted-foreground" />
          </CardHeader>

          <CardContent className="flex flex-col sm:flex-row pl-0 flex-wrap">
            <div className="flex flex-col flex-1">
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-1 pr-0">
                <CardTitle className="text-sm font-medium">
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger>
                        <div className="flex flex-row gap-1 items-center justify-start text-sm font-medium">
                          Spending Balance
                          <InfoIcon className="h-3 w-3 shrink-0 text-muted-foreground" />
                        </div>
                      </TooltipTrigger>
                      <TooltipContent className="w-[300px]">
                        Your spending balance is the funds in your VTXOs.
                      </TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                </CardTitle>
              </CardHeader>
              <CardContent className="flex-grow pb-0">
                {!balances && (
                  <div>
                    <div className="animate-pulse d-inline ">
                      <div className="h-2.5 bg-primary rounded-full w-12 my-2"></div>
                    </div>
                  </div>
                )}
                {balances && (
                  <>
                    <div className="text-xl font-medium balance sensitive mb-1">
                      {new Intl.NumberFormat().format(
                        Math.floor(balances.lightning.totalSpendable / 1000)
                      )}{" "}
                      sats
                    </div>
                    <FormattedFiatAmount
                      amount={balances.lightning.totalSpendable / 1000}
                    />
                  </>
                )}
              </CardContent>
            </div>
          </CardContent>
        </Card>
        <Card className="flex flex-1 flex-col">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-2xl font-semibold">On-Chain</CardTitle>
            <LinkIcon className="h-6 w-6 text-muted-foreground" />
          </CardHeader>
          <CardContent className="flex-grow">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-1 pl-0">
              <CardTitle className="text-sm font-medium">
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger>
                      <div className="flex flex-row gap-1 items-center text-sm font-medium">
                        Balance
                        <InfoIcon className="h-3 w-3 shrink-0 text-muted-foreground" />
                      </div>
                    </TooltipTrigger>
                    <TooltipContent className="w-[300px]">
                      Your on-chain balance can be used to board and off-board
                      the Ark.
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              </CardTitle>
            </CardHeader>
            {!balances && (
              <div>
                <div className="animate-pulse d-inline ">
                  <div className="h-2.5 bg-primary rounded-full w-12 my-2"></div>
                </div>
              </div>
            )}
            <div>
              {balances && (
                <>
                  <div className="mb-1">
                    <span className="text-xl font-medium balance sensitive mb-1 mr-1">
                      {new Intl.NumberFormat().format(
                        Math.floor(balances.onchain.spendable)
                      )}{" "}
                      sats
                    </span>
                  </div>
                  <FormattedFiatAmount
                    amount={balances.onchain.spendable}
                    className="mb-1"
                  />
                  {balances &&
                    balances.onchain.spendable !== balances.onchain.total && (
                      <p className="text-xs text-muted-foreground animate-pulse">
                        +
                        {new Intl.NumberFormat().format(
                          balances.onchain.total - balances.onchain.spendable
                        )}{" "}
                        sats incoming
                      </p>
                    )}
                </>
              )}
            </div>
          </CardContent>
        </Card>
      </div>

      {!vtxos && <Loading />}
      {vtxos && (
        <Card className="">
          <CardHeader>
            <CardTitle className="text-2xl">VTXOs</CardTitle>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[160px] text-muted-foreground">
                    Amount (sats)
                  </TableHead>
                  <TableHead className="text-muted-foreground">
                    Status
                  </TableHead>
                  <TableHead className="text-muted-foreground">
                    Expires at block
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {vtxos.map((vtxo) => {
                  return (
                    <TableRow key={vtxo.point.Txid + vtxo.point.Vout}>
                      <TableCell>
                        {new Intl.NumberFormat().format(vtxo.amount_sat)}
                      </TableCell>
                      <TableCell>
                        {vtxo.is_arkoor ? "Pending" : "Refreshed"}
                      </TableCell>
                      <TableCell>{vtxo.expiry_height}</TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}
    </>
  );
}
