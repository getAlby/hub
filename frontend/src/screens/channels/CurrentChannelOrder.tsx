import React from "react";
import { localStorageKeys } from "src/constants";
import {
  Channel,
  ConnectPeerRequest,
  GetOnchainAddressResponse,
  NewChannelOrder,
  Node,
  OpenChannelRequest,
  OpenChannelResponse,
} from "src/types";

import { Payment, init } from "@getalby/bitcoin-connect-react";
import { Copy, QrCode, RefreshCw } from "lucide-react";
import { Link } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import Loading from "src/components/Loading";
import QRCode from "src/components/QRCode";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from "src/components/ui/dialog";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { Table, TableBody, TableCell, TableRow } from "src/components/ui/table";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "src/components/ui/tooltip";
import { toast, useToast } from "src/components/ui/use-toast";
import { useBalances } from "src/hooks/useBalances";
import { useCSRF } from "src/hooks/useCSRF";
import { useChannels } from "src/hooks/useChannels";
import { useMempoolApi } from "src/hooks/useMempoolApi";
import { copyToClipboard } from "src/lib/clipboard";
import { Success } from "src/screens/onboarding/Success";
import useChannelOrderStore from "src/state/ChannelOrderStore";
import {
  NewInstantChannelInvoiceRequest,
  NewInstantChannelInvoiceResponse,
} from "src/types";
import { request } from "src/utils/request";
init({
  showBalance: false,
});

export function CurrentChannelOrder() {
  const order = useChannelOrderStore((store) => store.order);
  if (!order) {
    return (
      <p>
        No pending channel order.{" "}
        <Link to="/channels" className="underline">
          Return to channels page
        </Link>
      </p>
    );
  }
  return <ChannelOrderInternal order={order} />;
}

function ChannelOrderInternal({ order }: { order: NewChannelOrder }) {
  switch (order.status) {
    case "pay":
      switch (order.paymentMethod) {
        case "onchain":
          return <PayBitcoinChannelOrder order={order} />;
        case "lightning":
          return <PayLightningChannelOrder order={order} />;
        default:
          break;
      }
      break;
    case "opening":
      return <ChannelOpening fundingTxId={order.fundingTxId} />;
    case "success":
      return <Success />;
    default:
      break;
  }

  return (
    <p>
      TODO: {order.status} {order.paymentMethod}
    </p>
  );
}

function ChannelOpening({ fundingTxId }: { fundingTxId: string | undefined }) {
  const { data: channels } = useChannels(true);
  const channel = fundingTxId
    ? channels?.find((channel) => channel.fundingTxId === fundingTxId)
    : undefined;

  React.useEffect(() => {
    if (channel?.active) {
      useChannelOrderStore.getState().updateOrder({
        status: "success",
      });
    }
  }, [channel]);

  return (
    <div className="flex flex-col justify-center gap-2">
      <Card>
        <CardHeader>
          <CardTitle>Your channel is being opened</CardTitle>
          <CardDescription>
            Waiting for {channel?.confirmationsRequired ?? "unknown"}{" "}
            confirmations
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-row gap-2">
            <Loading />
            {channel?.confirmations ?? "0"} /{" "}
            {channel?.confirmationsRequired ?? "unknown"} confirmations
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function useEstimatedTransactionFee() {
  const { data: recommendedFees } = useMempoolApi<{ fastestFee: number }>(
    "/v1/fees/recommended",
    true
  );
  if (recommendedFees?.fastestFee) {
    // estimated transaction size: 200 vbytes
    return 200 * recommendedFees.fastestFee;
  }
}

// TODO: move these to new files
function PayBitcoinChannelOrder({ order }: { order: NewChannelOrder }) {
  if (order.paymentMethod !== "onchain") {
    throw new Error("incorrect payment method");
  }
  const { data: balances } = useBalances(true);
  const estimatedTransactionFee = useEstimatedTransactionFee();

  if (!balances || !estimatedTransactionFee) {
    return <Loading />;
  }

  const requiredAmount = +order.amount + estimatedTransactionFee;
  if (balances.onchain.spendable >= requiredAmount) {
    return <PayBitcoinChannelOrderWithSpendableFunds order={order} />;
  }
  if (balances.onchain.total >= requiredAmount) {
    return <PayBitcoinChannelOrderWaitingDepositConfirmation />;
  }
  return <PayBitcoinChannelOrderTopup order={order} />;
}

function PayBitcoinChannelOrderWaitingDepositConfirmation() {
  return (
    <>
      <p>Bitcoin deposited</p>
      <div className="flex items-center gap-2">
        <Loading />
        <p>Waiting for one block confirmation</p>
      </div>

      <p className="text-muted-foreground">estimated time: 10 minutes</p>
    </>
  );
}

function PayBitcoinChannelOrderTopup({ order }: { order: NewChannelOrder }) {
  if (order.paymentMethod !== "onchain") {
    throw new Error("incorrect payment method");
  }

  const { data: csrf } = useCSRF();
  const { data: balances } = useBalances();
  const [onchainAddress, setOnchainAddress] = React.useState<string>();
  const [isLoading, setLoading] = React.useState(false);
  const { data: mempoolAddressUtxos } = useMempoolApi<{ value: number }[]>(
    onchainAddress ? `/address/${onchainAddress}/utxo` : undefined,
    true
  );
  const estimatedTransactionFee = useEstimatedTransactionFee();

  const getNewAddress = React.useCallback(async () => {
    if (!csrf) {
      return;
    }
    setLoading(true);
    try {
      const response = await request<GetOnchainAddressResponse>(
        "/api/wallet/new-address",
        {
          method: "POST",
          headers: {
            "X-CSRF-Token": csrf,
            "Content-Type": "application/json",
          },
          //body: JSON.stringify({}),
        }
      );
      if (!response?.address) {
        throw new Error("No address in response");
      }
      localStorage.setItem(localStorageKeys.onchainAddress, response.address);
      setOnchainAddress(response.address);
    } catch (error) {
      alert("Failed to request a new address: " + error);
    } finally {
      setLoading(false);
    }
  }, [csrf]);

  React.useEffect(() => {
    const existingAddress = localStorage.getItem(
      localStorageKeys.onchainAddress
    );
    if (existingAddress) {
      setOnchainAddress(existingAddress);
      return;
    }
    getNewAddress();
  }, [getNewAddress]);

  if (!onchainAddress || !balances || !estimatedTransactionFee) {
    return (
      <div className="flex justify-center">
        <Loading />
      </div>
    );
  }

  const requiredAmount = +order.amount + estimatedTransactionFee;
  const unspentAmount =
    (mempoolAddressUtxos
      ?.map((utxo) => utxo.value)
      .reduce((a, b) => a + b, 0) || 0) - balances.onchain.reserved;

  if (unspentAmount >= requiredAmount) {
    return <PayBitcoinChannelOrderWaitingDepositConfirmation />;
  }

  return (
    <div className="grid gap-5">
      <AppHeader
        title="Deposit bitcoin"
        description="You don't have enough Bitcoin to open your intended channel"
      />
      <div className="grid gap-5 max-w-lg">
        {unspentAmount > 0 && <p>{unspentAmount} sats deposited</p>}

        <div className="grid gap-1.5">
          <Label htmlFor="text">On-Chain Address</Label>
          <p className="text-xs text-muted-foreground">
            You currently have {balances.onchain.total} sats. You need to deposit at
            least another {requiredAmount - balances.onchain.total} sats to cover
            channel opening fees.
          </p>
          <div className="flex flex-row gap-2 items-center">
            <Input type="text" value={onchainAddress} readOnly className="flex-1" />
            <Button
              variant="secondary"
              size="icon"
              onClick={() => { copyToClipboard(onchainAddress); toast({ title: "Copied to clipboard." }) }}>
              <Copy className="w-4 h-4" />
            </Button>
            <Dialog>
              <DialogTrigger asChild>
                <Button variant="secondary" size="icon">
                  <QrCode className="w-4 h-4" />
                </Button>
              </DialogTrigger>

              <DialogContent>
                <DialogHeader>
                  <DialogTitle>
                    Deposit bitcoin
                  </DialogTitle>
                  <DialogDescription>
                    Scan this QR code with your wallet to send funds.
                  </DialogDescription>
                </DialogHeader>
                <div className="flex flex-row justify-center p-3">
                  <a
                    href={`bitcoin:${onchainAddress}`}
                    target="_blank"
                  >
                    <QRCode value={onchainAddress} />
                  </a>
                </div>
              </DialogContent>
            </Dialog>
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <LoadingButton
                    variant="secondary"
                    size="icon"
                    onClick={getNewAddress}
                    loading={isLoading}>
                    <RefreshCw className="w-4 h-4" />
                  </LoadingButton>
                </TooltipTrigger>
                <TooltipContent>
                  Generate a new address
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </div>
        </div>


        <Card>
          <CardHeader>
            <CardTitle className="flex flex-row items-center gap-2">
              <Loading /> Waiting for your transaction
            </CardTitle>
            <CardDescription>
              Send a bitcoin transaction to the address provided above. You'll be redirected as soon as the transaction is seen in the mempool.
            </CardDescription>
          </CardHeader>
        </Card>
      </div>
    </div>
  );
}

function PayBitcoinChannelOrderWithSpendableFunds({
  order,
}: {
  order: NewChannelOrder;
}) {
  if (order.paymentMethod !== "onchain") {
    throw new Error("incorrect payment method");
  }
  const [nodeDetails, setNodeDetails] = React.useState<Node | undefined>();
  const { data: csrf } = useCSRF();
  const { toast } = useToast();
  const [, setHasCalledOpenChannel] = React.useState(false);

  const { pubkey, host } = order;

  const fetchNodeDetails = React.useCallback(async () => {
    if (!pubkey) {
      return;
    }
    try {
      const data = await request<Node>(
        `/api/mempool?endpoint=/v1/lightning/nodes/${pubkey}`
      );
      setNodeDetails(data);
    } catch (error) {
      console.error(error);
      setNodeDetails(undefined);
    }
  }, [pubkey]);

  React.useEffect(() => {
    fetchNodeDetails();
  }, [fetchNodeDetails]);

  const connectPeer = React.useCallback(async () => {
    if (!csrf) {
      throw new Error("csrf not loaded");
    }
    if (!nodeDetails && !host) {
      throw new Error("node details not found");
    }
    const _host = nodeDetails?.sockets
      ? nodeDetails.sockets.split(",")[0]
      : host;
    const [address, port] = _host.split(":");
    if (!address || !port) {
      throw new Error("host not found");
    }
    console.log(`🔌 Peering with ${pubkey}`);
    const connectPeerRequest: ConnectPeerRequest = {
      pubkey,
      address,
      port: +port,
    };
    await request("/api/peers", {
      method: "POST",
      headers: {
        "X-CSRF-Token": csrf,
        "Content-Type": "application/json",
      },
      body: JSON.stringify(connectPeerRequest),
    });
  }, [csrf, nodeDetails, pubkey, host]);

  const openChannel = React.useCallback(async () => {
    try {
      if (order.paymentMethod !== "onchain") {
        throw new Error("incorrect payment method");
      }
      if (!csrf) {
        throw new Error("csrf not loaded");
      }

      await connectPeer();

      console.log(`🎬 Opening channel with ${pubkey}`);

      const openChannelRequest: OpenChannelRequest = {
        pubkey,
        amount: +order.amount,
        public: order.isPublic,
      };
      const openChannelResponse = await request<OpenChannelResponse>(
        "/api/channels",
        {
          method: "POST",
          headers: {
            "X-CSRF-Token": csrf,
            "Content-Type": "application/json",
          },
          body: JSON.stringify(openChannelRequest),
        }
      );

      if (!openChannelResponse?.fundingTxId) {
        throw new Error("No funding txid in response");
      }
      toast({
        title: "Channel opening transaction published!",
      });
      useChannelOrderStore.getState().updateOrder({
        fundingTxId: openChannelResponse.fundingTxId,
        status: "opening",
      });
    } catch (error) {
      console.error(error);
      alert("Something went wrong: " + error);
    }
  }, [connectPeer, csrf, order, pubkey, toast]);

  React.useEffect(() => {
    setHasCalledOpenChannel((hasCalledOpenChannel) => {
      if (!hasCalledOpenChannel) {
        openChannel();
      }
      return true;
    });
  }, [openChannel, order.amount, pubkey]);

  return (
    <div className="flex flex-col gap-5">
      <AppHeader
        title="Opening channel"
        description="Your funds have been successfully deposited"
      />

      <div className="flex flex-col gap-5">
        <Loading />
        <p>Please wait...</p>
      </div>
    </div>
  );
}

function PayLightningChannelOrder({ order }: { order: NewChannelOrder }) {
  if (order.paymentMethod !== "lightning") {
    throw new Error("incorrect payment method");
  }
  const { data: csrf } = useCSRF();
  const { toast } = useToast();
  const { data: channels } = useChannels(true);
  const [, setRequestedInvoice] = React.useState(false);
  const [prevChannels, setPrevChannels] = React.useState<
    Channel[] | undefined
  >();
  const [wrappedInvoiceResponse, setWrappedInvoiceResponse] = React.useState<
    NewInstantChannelInvoiceResponse | undefined
  >();

  // This is not a good check if user already has enough inbound liquidity
  // - check balance instead or how else to check the invoice is paid?
  const newChannel =
    channels && prevChannels
      ? channels.find(
        (newChannel) =>
          !prevChannels.some((current) => current.id === newChannel.id)
      )
      : undefined;

  React.useEffect(() => {
    if (newChannel) {
      (async () => {
        toast({ title: "Channel opened!" });
        setTimeout(() => {
          useChannelOrderStore.getState().updateOrder({
            status: "opening",
            fundingTxId: newChannel.fundingTxId,
          });
        }, 3000);
      })();
    }
  }, [newChannel, toast]);

  React.useEffect(() => {
    // TODO: move fetching to NewChannel page otherwise fee cannot be retrieved
    if (!channels || !csrf) {
      return;
    }
    setRequestedInvoice((current) => {
      if (!current) {
        (async () => {
          try {
            setPrevChannels(channels);
            if (!order.lsp) {
              throw new Error("no lsp selected");
            }
            const newJITChannelRequest: NewInstantChannelInvoiceRequest = {
              lsp: order.lsp,
              amount: parseInt(order.amount),
            };
            const response = await request<NewInstantChannelInvoiceResponse>(
              "/api/instant-channel-invoices",
              {
                method: "POST",
                headers: {
                  "X-CSRF-Token": csrf,
                  "Content-Type": "application/json",
                },
                body: JSON.stringify(newJITChannelRequest),
              }
            );
            if (!response?.invoice) {
              throw new Error("No invoice in response");
            }
            setWrappedInvoiceResponse(response);
          } catch (error) {
            alert("Failed to connect to request wrapped invoice: " + error);
          }
        })();
      }
      return true;
    });
  }, [channels, csrf, order.amount, order.lsp]);

  return (
    <div className="flex flex-col gap-5">
      <AppHeader
        title={"Buy an Instant Channel"}
        description={
          wrappedInvoiceResponse
            ? "Complete Payment to open an instant channel to your node"
            : "Please wait, loading..."
        }
      />
      {!wrappedInvoiceResponse && <Loading />}

      {wrappedInvoiceResponse && (
        <>
          <div className="max-w-md flex flex-col gap-5">
            <div className="border rounded-lg">
              <Table>
                <TableBody>
                  <TableRow>
                    <TableCell className="font-medium p-3 flex flex-row gap-1.5 items-center">
                      Fee
                    </TableCell>
                    <TableCell className="text-right p-3">
                      {new Intl.NumberFormat().format(
                        wrappedInvoiceResponse.fee
                      )}{" "}
                      sats
                    </TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell className="font-medium p-3">
                      Amount to pay
                    </TableCell>
                    <TableCell className="font-semibold text-right p-3">
                      {new Intl.NumberFormat().format(parseInt(order.amount))}{" "}
                      sats
                    </TableCell>
                  </TableRow>
                </TableBody>
              </Table>
            </div>
            <Payment
              invoice={wrappedInvoiceResponse.invoice}
              payment={newChannel ? { preimage: "dummy preimage" } : undefined}
              paymentMethods="external"
            />
          </div>
        </>
      )}
    </div>
  );
}
