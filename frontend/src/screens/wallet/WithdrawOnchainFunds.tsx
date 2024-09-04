import { AlertTriangleIcon, CopyIcon, ExternalLinkIcon } from "lucide-react";
import React from "react";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import Loading from "src/components/Loading";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "src/components/ui/alert-dialog";
import { Button } from "src/components/ui/button";
import { Checkbox } from "src/components/ui/checkbox";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";
import { ONCHAIN_DUST_SATS } from "src/constants";
import { useBalances } from "src/hooks/useBalances";

import { copyToClipboard } from "src/lib/clipboard";
import { RedeemOnchainFundsResponse } from "src/types";
import { request } from "src/utils/request";

export default function WithdrawOnchainFunds() {
  const [isLoading, setLoading] = React.useState(false);
  const { toast } = useToast();
  const { data: balances } = useBalances();
  const [onchainAddress, setOnchainAddress] = React.useState("");
  const [amount, setAmount] = React.useState("");
  const [sendAll, setSendAll] = React.useState(false);
  const [transactionId, setTransactionId] = React.useState("");
  const [confirmDialogOpen, setConfirmDialogOpen] = React.useState(false);

  const copy = (text: string) => {
    copyToClipboard(text, toast);
  };

  const redeemFunds = React.useCallback(async () => {
    setLoading(true);
    try {
      await new Promise((resolve) => setTimeout(resolve, 100));
      if (!onchainAddress) {
        throw new Error("No onchain address");
      }
    } catch (error) {
      console.error(error);
      toast({
        title: "Something went wrong",
        description: "" + error,
        variant: "destructive",
      });
      setLoading(false);
      return;
    }

    try {
      const response = await request<RedeemOnchainFundsResponse>(
        "/api/wallet/redeem-onchain-funds",
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            toAddress: onchainAddress,
            amount: +amount,
            sendAll,
          }),
        }
      );
      console.info("Redeemed onchain funds", response);
      if (!response?.txId) {
        throw new Error("No address in response");
      }
      setTransactionId(response.txId);
    } catch (error) {
      console.error(error);
      toast({
        variant: "destructive",
        title: "Failed to redeem onchain funds",
        description: "" + error,
      });
    }
    setLoading(false);
  }, [amount, onchainAddress, sendAll, toast]);

  if (transactionId) {
    return (
      <div className="grid gap-5">
        <AppHeader
          title="Withdrawal Transaction Broadcasted"
          description={
            "You will receive the funds at the destination after the transaction is confirmed"
          }
        />
        <p className="text-primary">Withdrawal Transaction Id</p>
        <div className="flex items-center justify-between gap-4 max-w-sm">
          <p className="break-all font-semibold">{transactionId}</p>
          <CopyIcon
            className="cursor-pointer text-muted-foreground w-4 h-4 flex-shrink-0"
            onClick={() => {
              copy(transactionId);
            }}
          />
        </div>
        <ExternalLink
          to={`https://mempool.space/tx/${transactionId}`}
          className="underline flex items-center mt-2"
        >
          View on Mempool
          <ExternalLinkIcon className="w-4 h-4 ml-2" />
        </ExternalLink>
        <p>Your savings balance in Alby Hub may take some time to update.</p>
      </div>
    );
  }

  if (!balances) {
    return <Loading />;
  }

  if (balances.onchain.spendable <= ONCHAIN_DUST_SATS) {
    return (
      <p>
        You currently don't have enough sats to pay for an onchain transaction.
      </p>
    );
  }

  return (
    <div className="grid gap-5">
      <AppHeader
        title="Withdraw Savings Balance"
        description="Withdraw your onchain funds to another bitcoin wallet"
      />

      <div className="max-w-lg">
        <p>
          Your savings balance will be withdrawn to the onchain bitcoin wallet
          address you specify below.
        </p>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            setConfirmDialogOpen(true);
          }}
          className="grid gap-5 mt-4"
        >
          <div className="">
            <Label htmlFor="amount">Amount</Label>
            <div className="flex justify-between items-center mb-1">
              <p className="text-sm text-muted-foreground">
                Current onchain balance:{" "}
                {new Intl.NumberFormat().format(balances.onchain.spendable)}{" "}
                sats
              </p>
              <div className="flex items-center gap-1">
                <Checkbox
                  id="send-all"
                  onCheckedChange={() => setSendAll(!sendAll)}
                />
                <Label htmlFor="send-all" className="text-xs">
                  Send All
                </Label>
              </div>
            </div>
            {!sendAll && (
              <Input
                id="amount"
                type="number"
                value={amount}
                required
                onChange={(e) => {
                  setAmount(e.target.value);
                }}
              />
            )}
            {sendAll && (
              <Alert className="mt-4">
                <AlertTriangleIcon className="h-4 w-4" />
                <AlertTitle>Entire wallet balance will be sent</AlertTitle>
                <AlertDescription>
                  Your entire wallet balance
                  {balances.onchain.reserved > 0 && (
                    <>
                      {" "}
                      including reserves (
                      {new Intl.NumberFormat().format(
                        balances.onchain.reserved
                      )}{" "}
                      sats)
                    </>
                  )}{" "}
                  will be sent minus onchain transaction fees. The exact amount
                  cannot be determined until the payment is made.
                  {balances.onchain.reserved && (
                    <>
                      {" "}
                      You have channels open and this withdrawal will deplete
                      your anchor reserves, which may make it harder to close
                      channels without depositing additional onchain funds to
                      your savings balance.
                    </>
                  )}
                </AlertDescription>
              </Alert>
            )}
            {!!balances?.onchain.reserved &&
              !sendAll &&
              +amount > balances.onchain.spendable * 0.9 && (
                <Alert className="mt-4">
                  <AlertTriangleIcon className="h-4 w-4" />
                  <AlertTitle>
                    Channel Anchor Reserves may be depleted
                  </AlertTitle>
                  <AlertDescription>
                    You have channels open and this withdrawal may deplete your
                    anchor reserves, which may make it harder to close channels
                    without depositing additional onchain funds to your savings
                    balance.
                  </AlertDescription>
                </Alert>
              )}
          </div>
          <div className="">
            <Label htmlFor="onchain-address">Onchain Address</Label>
            <Input
              id="onchain-address"
              type="text"
              value={onchainAddress}
              required
              onChange={(e) => {
                setOnchainAddress(e.target.value);
              }}
            />
          </div>

          <p className="text-sm text-muted-foreground">
            Please double-check the destination address. This transaction cannot
            be reversed.
          </p>

          <div>
            <AlertDialog
              onOpenChange={setConfirmDialogOpen}
              open={confirmDialogOpen}
            >
              <Button>Withdraw</Button>

              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>
                    Confirm Onchain Transaction
                  </AlertDialogTitle>
                  <AlertDialogDescription>
                    <p>
                      Please confirm your payment to{" "}
                      <span className="font-bold">{onchainAddress}</span>
                    </p>
                    <p className="mt-4">
                      Amount:{" "}
                      <span className="font-bold">
                        {sendAll ? (
                          "entire savings balance"
                        ) : (
                          <>{new Intl.NumberFormat().format(+amount)} sats</>
                        )}
                      </span>
                    </p>
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>Cancel</AlertDialogCancel>

                  <LoadingButton
                    loading={isLoading}
                    onClick={() => redeemFunds()}
                  >
                    Confirm
                  </LoadingButton>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          </div>
        </form>
      </div>
    </div>
  );
}
