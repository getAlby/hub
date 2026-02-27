import {
  AlertTriangleIcon,
  ChevronDown,
  CopyIcon,
  ExternalLinkIcon,
  InfoIcon,
} from "lucide-react";
import React from "react";
import { toast } from "sonner";
import { AnchorReserveAlert } from "src/components/AnchorReserveAlert";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import Loading from "src/components/Loading";
import { MempoolAlert } from "src/components/MempoolAlert";
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
import { ExternalLinkButton } from "src/components/ui/custom/external-link-button";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { Separator } from "src/components/ui/separator";
import { ONCHAIN_DUST_SATS } from "src/constants";
import { useBalances } from "src/hooks/useBalances";
import { useBitcoinMaxiMode } from "src/hooks/useBitcoinMaxiMode";
import { useInfo } from "src/hooks/useInfo";
import { useMempoolApi } from "src/hooks/useMempoolApi";

import { copyToClipboard } from "src/lib/clipboard";
import { RedeemOnchainFundsResponse } from "src/types";
import { request } from "src/utils/request";

export default function WithdrawOnchainFunds() {
  const { bitcoinMaxiMode } = useBitcoinMaxiMode();
  const { data: info } = useInfo();
  const { data: balances } = useBalances();
  const { data: recommendedFees, error: mempoolError } = useMempoolApi<{
    fastestFee: number;
    halfHourFee: number;
    economyFee: number;
    minimumFee: number;
  }>("/v1/fees/recommended");
  const [isLoading, setLoading] = React.useState(false);
  const [onchainAddress, setOnchainAddress] = React.useState("");
  const [amount, setAmount] = React.useState("");
  const [feeRate, setFeeRate] = React.useState("");
  const [sendAll, setSendAll] = React.useState(false);
  const [showAdvanced, setShowAdvanced] = React.useState(false);
  const [transactionId, setTransactionId] = React.useState("");
  const [confirmDialogOpen, setConfirmDialogOpen] = React.useState(false);

  React.useEffect(() => {
    if (mempoolError) {
      setShowAdvanced(true);
    }
  }, [mempoolError]);

  React.useEffect(() => {
    if (recommendedFees?.fastestFee) {
      setFeeRate(recommendedFees.fastestFee.toString());
    }
  }, [recommendedFees]);

  const copy = (text: string) => {
    copyToClipboard(text);
  };

  const redeemFunds = React.useCallback(async () => {
    setLoading(true);
    try {
      if (!onchainAddress) {
        throw new Error("No onchain address");
      }
      if (!feeRate) {
        throw new Error("No fee rate set");
      }
    } catch (error) {
      console.error(error);
      toast.error("Something went wrong", {
        description: "" + error,
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
            feeRate: +feeRate,
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
      toast.error("Failed to redeem onchain funds", {
        description: "" + error,
      });
    }
    setLoading(false);
  }, [amount, feeRate, onchainAddress, sendAll]);

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
            className="cursor-pointer text-muted-foreground size-4 shrink-0"
            onClick={() => {
              copy(transactionId);
            }}
          />
        </div>
        <ExternalLink
          to={`${info?.mempoolUrl}/tx/${transactionId}`}
          className="underline flex items-center mt-2"
        >
          View on Mempool
          <ExternalLinkIcon className="size-4 ml-2" />
        </ExternalLink>
        <p>Your on-chain balance in Alby Hub may take some time to update.</p>
      </div>
    );
  }

  if (!info || !balances || (!recommendedFees && !mempoolError)) {
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
        title="Withdraw On-Chain Balance"
        description="Withdraw your onchain funds to another bitcoin wallet"
      />

      <div className="max-w-lg">
        <p>
          Your on-chain balance will be withdrawn to the onchain bitcoin wallet
          address you specify below.
        </p>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            setConfirmDialogOpen(true);
          }}
          className="grid gap-5 mt-4"
        >
          <div>
            <div className="grid gap-2">
              <Label htmlFor="amount">Amount</Label>
              <div className="flex justify-between items-center">
                <p className="text-sm text-muted-foreground sensitive slashed-zero">
                  Current onchain balance:{" "}
                  <FormattedBitcoinAmount
                    amount={balances.onchain.spendable * 1000}
                  />
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
            </div>
            <MempoolAlert className="mt-4" />
            {sendAll && (
              <Alert className="mt-4" variant="warning">
                <AlertTriangleIcon />
                <AlertTitle>Entire wallet balance will be sent</AlertTitle>
                <AlertDescription>
                  Your entire wallet balance will be sent minus onchain
                  transaction fees. The exact amount cannot be determined until
                  the payment is made.
                </AlertDescription>
              </Alert>
            )}
            <AnchorReserveAlert
              amount={sendAll ? balances.onchain.spendable : +amount}
              className="mt-4"
            />
          </div>
          <div className="grid gap-2">
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
            <p className="text-sm text-muted-foreground">
              Please double-check the destination address. This transaction
              cannot be reversed.
            </p>
          </div>
          {(info?.backendType === "LDK" || info?.backendType === "LND") && (
            <>
              {showAdvanced && (
                <div className="grid gap-2">
                  <Label htmlFor="fee-rate">Fee Rate (Sat/vB)</Label>
                  {mempoolError && (
                    <div className="text-muted-foreground text-xs flex gap-1 items-center">
                      <AlertTriangleIcon className="h-3 w-3" />
                      Failed to fetch fee estimates. Try refreshing the page.
                    </div>
                  )}
                  <Input
                    id="fee-rate"
                    type="number"
                    value={feeRate}
                    step={1}
                    required
                    min={recommendedFees?.minimumFee || 1}
                    onChange={(e) => {
                      setFeeRate(e.target.value);
                    }}
                  />
                  {recommendedFees && (
                    <div className="flex items-center mt-2 gap-4">
                      <Button
                        variant="positive"
                        className="rounded-full"
                        type="button"
                        onClick={() =>
                          setFeeRate(recommendedFees.economyFee.toString())
                        }
                      >
                        Low priority: {recommendedFees.economyFee}
                      </Button>{" "}
                      <Button
                        variant="positive"
                        className="rounded-full"
                        type="button"
                        onClick={() =>
                          setFeeRate(recommendedFees.fastestFee.toString())
                        }
                      >
                        High priority: {recommendedFees.fastestFee}
                      </Button>{" "}
                      <ExternalLink
                        to={info?.mempoolUrl}
                        className="text-sm text-muted-foreground underline flex items-center gap-2"
                      >
                        View on Mempool
                        <ExternalLinkIcon className="w-4 h-4" />
                      </ExternalLink>
                    </div>
                  )}
                </div>
              )}
              {!showAdvanced && (
                <Button
                  type="button"
                  variant="link"
                  className="text-muted-foreground text-xs"
                  onClick={() => setShowAdvanced((current) => !current)}
                >
                  <ChevronDown />
                  Advanced Options
                </Button>
              )}
            </>
          )}

          <div>
            <AlertDialog
              onOpenChange={setConfirmDialogOpen}
              open={confirmDialogOpen}
            >
              <Button className="w-full">Withdraw</Button>
              {feeRate && (
                <div className="mt-2 text-muted-foreground text-sm flex gap-1 items-center justify-center">
                  <InfoIcon className="h-4 w-4" />
                  On-chain payment will be made with{" "}
                  <span className="font-semibold">{feeRate} sat/vB</span> fee
                </div>
              )}

              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>
                    Confirm Onchain Transaction
                  </AlertDialogTitle>
                  <AlertDialogDescription>
                    <div>
                      <p>Please confirm your payment to</p>
                      <p className="font-bold max-w-md break-anywhere">
                        {onchainAddress}
                      </p>
                      <p className="mt-4">
                        Amount:{" "}
                        <span className="font-bold slashed-zero">
                          {sendAll ? (
                            "entire on-chain balance"
                          ) : (
                            <>
                              <FormattedBitcoinAmount amount={+amount * 1000} />
                            </>
                          )}
                        </span>
                      </p>
                      {feeRate && (
                        <p className="mt-4">
                          Fee Rate:{" "}
                          <span className="font-bold slashed-zero">
                            {feeRate}
                          </span>{" "}
                          sat/vB
                        </p>
                      )}
                    </div>
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

            {!bitcoinMaxiMode && (
              <>
                <Separator className="my-4" />
                <ExternalLinkButton
                  to={`https://ff.io/?from=BTC&ref=qnnjvywb`}
                  className="w-full"
                  variant="secondary"
                >
                  <ExternalLinkIcon className="size-4" />
                  Withdraw to other Cryptocurrency
                </ExternalLinkButton>
              </>
            )}
          </div>
        </form>
      </div>
    </div>
  );
}
