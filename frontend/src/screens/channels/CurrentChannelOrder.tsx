import React from "react";
import {
  Channel,
  ConnectPeerRequest,
  NewChannelOrder,
  Node,
  OpenChannelRequest,
  OpenChannelResponse,
  PayInvoiceResponse,
} from "src/types";

import { Payment, init } from "@getalby/bitcoin-connect-react";
import { Copy, QrCode, RefreshCw } from "lucide-react";
import { Link } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import Loading from "src/components/Loading";
import QRCode from "src/components/QRCode";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "src/components/ui/dialog";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { Table, TableBody, TableCell, TableRow } from "src/components/ui/table";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "src/components/ui/tooltip";
import { useToast } from "src/components/ui/use-toast";
import { useBalances } from "src/hooks/useBalances";
import { useCSRF } from "src/hooks/useCSRF";
import { useChannels } from "src/hooks/useChannels";
import { useMempoolApi } from "src/hooks/useMempoolApi";
import { useOnchainAddress } from "src/hooks/useOnchainAddress";
import { usePeers } from "src/hooks/usePeers";
import { useSyncWallet } from "src/hooks/useSyncWallet";
import { copyToClipboard } from "src/lib/clipboard";
import { splitSocketAddress } from "src/lib/utils";
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
let hasStartedOpenedChannel = false;

export function CurrentChannelOrder() {
  React.useEffect(() => {
    hasStartedOpenedChannel = false;
  }, []);
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
  useSyncWallet();
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
      <div className="w-full mt-40 gap-20 flex flex-col items-center justify-center">
        <p>Feel free to leave this page or browse around Alby Hub!</p>
        <p>We'll send you an email as soon as your channel is active.</p>
      </div>
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

  // expect at least the user to have more funds than the channel size, hopefully enough to cover mempool fees.
  if (balances.onchain.spendable > +order.amount) {
    return <PayBitcoinChannelOrderWithSpendableFunds order={order} />;
  }
  if (balances.onchain.total > +order.amount) {
    return <PayBitcoinChannelOrderWaitingDepositConfirmation />;
  }
  return <PayBitcoinChannelOrderTopup order={order} />;
}

function PayBitcoinChannelOrderWaitingDepositConfirmation() {
  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle className="flex flex-row items-center gap-2">
            Bitcoin deposited
          </CardTitle>
        </CardHeader>
        <CardContent className="flex items-center gap-2">
          <Loading /> Waiting for one block confirmation
        </CardContent>
        <CardFooter className="text-muted-foreground">
          estimated time: 10 minutes
        </CardFooter>
      </Card>
    </>
  );
}

function PayBitcoinChannelOrderTopup({ order }: { order: NewChannelOrder }) {
  if (order.paymentMethod !== "onchain") {
    throw new Error("incorrect payment method");
  }

  const { data: channels } = useChannels();

  const { data: balances } = useBalances();
  const {
    data: onchainAddress,
    getNewAddress,
    loadingAddress,
  } = useOnchainAddress();

  const { data: mempoolAddressUtxos } = useMempoolApi<{ value: number }[]>(
    onchainAddress ? `/address/${onchainAddress}/utxo` : undefined,
    true
  );
  const estimatedTransactionFee = useEstimatedTransactionFee();

  if (!onchainAddress || !balances || !estimatedTransactionFee) {
    return (
      <div className="flex justify-center">
        <Loading />
      </div>
    );
  }

  // expect at least the user to have more funds than the channel size, hopefully enough to cover mempool fees.
  // This only considers one UTXO and will not work well if the user generates a new address.
  // However, this is just a fallback because LDK only updates onchain balances ~ once per minute.
  const unspentAmount =
    mempoolAddressUtxos?.map((utxo) => utxo.value).reduce((a, b) => a + b, 0) ||
    0;

  if (unspentAmount > +order.amount) {
    return <PayBitcoinChannelOrderWaitingDepositConfirmation />;
  }

  const num0ConfChannels =
    channels?.filter((c) => c.confirmationsRequired === 0).length || 0;

  const estimatedAnchorReserve = Math.max(
    num0ConfChannels * 25000 - balances.onchain.reserved,
    0
  );

  const missingAmount =
    +order.amount +
    estimatedTransactionFee +
    estimatedAnchorReserve -
    balances.onchain.total;

  const recommendedAmount = Math.ceil(missingAmount / 10000) * 10000;
  const topupLink = `https://getalby.com/topup?address=${onchainAddress}&receive_amount=${recommendedAmount}`;

  return (
    <div className="grid gap-5">
      <AppHeader
        title="Deposit bitcoin"
        description="You don't have enough Bitcoin to open your intended channel"
      />
      <div className="grid gap-5 max-w-lg">
        <div className="grid gap-1.5">
          <Label htmlFor="text">On-Chain Address</Label>
          <p className="text-xs">
            You currently have{" "}
            {new Intl.NumberFormat().format(balances.onchain.total)} sats. We
            recommend to deposit another{" "}
            {new Intl.NumberFormat().format(recommendedAmount)} sats to open a
            channel.{" "}
          </p>
          <p className="text-xs text-muted-foreground">
            ~{new Intl.NumberFormat().format(+missingAmount)} sats are missing
            to cover the cost of opening the channel, including onchain fees and
            potential onchain channel reserves.
          </p>
          <div className="flex flex-row gap-2 items-center">
            <Input
              type="text"
              value={onchainAddress}
              readOnly
              className="flex-1"
            />
            <Button
              variant="secondary"
              size="icon"
              onClick={() => {
                copyToClipboard(onchainAddress);
              }}
            >
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
                  <DialogTitle>Deposit bitcoin</DialogTitle>
                  <DialogDescription>
                    Scan this QR code with your wallet to send funds.
                  </DialogDescription>
                </DialogHeader>
                <div className="flex flex-row justify-center p-3">
                  <a href={`bitcoin:${onchainAddress}`} target="_blank">
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
                    loading={loadingAddress}
                    className="w-9 h-9"
                  >
                    {!loadingAddress && <RefreshCw className="w-4 h-4" />}
                  </LoadingButton>
                </TooltipTrigger>
                <TooltipContent>Generate a new address</TooltipContent>
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
              Send a bitcoin transaction to the address provided above. You'll
              be redirected as soon as the transaction is seen in the mempool.
            </CardDescription>
          </CardHeader>
          {unspentAmount > 0 && (
            <CardContent>{unspentAmount} sats deposited</CardContent>
          )}
        </Card>

        <ExternalLink to={topupLink} className="w-full">
          <Button className="w-full">
            Topup with your credit card or bank account
          </Button>
        </ExternalLink>
        <Link to="/channels/incoming" className="w-full">
          <Button className="w-full" variant="secondary">
            Need receiving capacity?
          </Button>
        </Link>
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
  const { data: peers } = usePeers();
  const [nodeDetails, setNodeDetails] = React.useState<Node | undefined>();
  const [hasLoadedNodeDetails, setLoadedNodeDetails] = React.useState(false);
  const { data: csrf } = useCSRF();
  const { toast } = useToast();

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
    setLoadedNodeDetails(true);
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
    const socketAddress = nodeDetails?.sockets
      ? nodeDetails.sockets.split(",")[0]
      : host;

    const { address, port } = splitSocketAddress(socketAddress);

    if (!address || !port) {
      throw new Error("host not found");
    }
    console.info(`🔌 Peering with ${pubkey}`);
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

      if (!peers) {
        throw new Error("peers not loaded");
      }

      // only pair if necessary
      // also allows to open channel to existing peer without providing a socket address.
      if (!peers.some((peer) => peer.nodeId === pubkey)) {
        await connectPeer();
      }

      console.info(`🎬 Opening channel with ${pubkey}`);

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
      console.info(
        "Channel opening transaction published",
        openChannelResponse.fundingTxId
      );
      toast({
        title: "Successfully published channel opening transaction",
      });
      useChannelOrderStore.getState().updateOrder({
        fundingTxId: openChannelResponse.fundingTxId,
        status: "opening",
      });
    } catch (error) {
      console.error(error);
      alert("Something went wrong: " + error);
    }
  }, [
    connectPeer,
    csrf,
    order.amount,
    order.isPublic,
    order.paymentMethod,
    peers,
    pubkey,
    toast,
  ]);

  React.useEffect(() => {
    if (!peers || !csrf || !hasLoadedNodeDetails || hasStartedOpenedChannel) {
      return;
    }

    hasStartedOpenedChannel = true;
    openChannel();
  }, [csrf, hasLoadedNodeDetails, openChannel, order.amount, peers, pubkey]);

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
            !prevChannels.some((current) => current.id === newChannel.id) &&
            newChannel.fundingTxId
        )
      : undefined;

  React.useEffect(() => {
    if (newChannel) {
      (async () => {
        toast({ title: "Successfully opened channel" });
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
            if (!order.lspType || !order.lspUrl) {
              throw new Error("missing lsp info in order");
            }
            const newInstantChannelInvoiceRequest: NewInstantChannelInvoiceRequest =
              {
                lspType: order.lspType,
                lspUrl: order.lspUrl,
                amount: parseInt(order.amount),
                public: order.isPublic,
              };
            const response = await request<NewInstantChannelInvoiceResponse>(
              "/api/instant-channel-invoices",
              {
                method: "POST",
                headers: {
                  "X-CSRF-Token": csrf,
                  "Content-Type": "application/json",
                },
                body: JSON.stringify(newInstantChannelInvoiceRequest),
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
  }, [
    channels,
    csrf,
    order.amount,
    order.isPublic,
    order.lspType,
    order.lspUrl,
  ]);

  const canPayInternally =
    channels &&
    wrappedInvoiceResponse &&
    channels.some(
      (channel) =>
        channel.localSpendableBalance / 1000 >
        wrappedInvoiceResponse.invoiceAmount
    );
  const [isPaying, setPaying] = React.useState(false);
  const [paid, setPaid] = React.useState(false);
  const [payExternally, setPayExternally] = React.useState(false);

  return (
    <div className="flex flex-col gap-5">
      <AppHeader
        title={"Buy Channel"}
        description={
          wrappedInvoiceResponse
            ? "Complete Payment to open a channel to your node"
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
                  {wrappedInvoiceResponse.outgoingLiquidity > 0 && (
                    <TableRow>
                      <TableCell className="font-medium p-3">
                        Spending Balance
                      </TableCell>
                      <TableCell className="text-right p-3">
                        {new Intl.NumberFormat().format(
                          wrappedInvoiceResponse.outgoingLiquidity
                        )}{" "}
                        sats
                      </TableCell>
                    </TableRow>
                  )}
                  {wrappedInvoiceResponse.incomingLiquidity > 0 && (
                    <TableRow>
                      <TableCell className="font-medium p-3">
                        Incoming Liquidity
                      </TableCell>
                      <TableCell className="text-right p-3">
                        {new Intl.NumberFormat().format(
                          wrappedInvoiceResponse.incomingLiquidity
                        )}{" "}
                        sats
                      </TableCell>
                    </TableRow>
                  )}
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
                      {new Intl.NumberFormat().format(
                        wrappedInvoiceResponse.invoiceAmount
                      )}{" "}
                      sats
                    </TableCell>
                  </TableRow>
                </TableBody>
              </Table>
            </div>
            {paid ? (
              <div className="flex gap-2 items-center justify-center">
                <Loading /> <p>Waiting for channel to be opened...</p>
              </div>
            ) : (
              <>
                {canPayInternally && (
                  <>
                    <LoadingButton
                      loading={isPaying}
                      className="mt-4"
                      onClick={async () => {
                        try {
                          if (!csrf) {
                            throw new Error("csrf not loaded");
                          }
                          setPaying(true);
                          const payInvoiceResponse =
                            await request<PayInvoiceResponse>(
                              `/api/payments/${wrappedInvoiceResponse.invoice}`,
                              {
                                method: "POST",
                                headers: {
                                  "X-CSRF-Token": csrf,
                                  "Content-Type": "application/json",
                                },
                              }
                            );
                          if (payInvoiceResponse) {
                            setPaid(true);
                            toast({
                              title: "Channel successfully requested",
                            });
                          }
                          setPaid(true);
                        } catch (e) {
                          toast({
                            variant: "destructive",
                            title: "Failed to send: " + e,
                          });
                          console.error(e);
                        }
                        setPaying(false);
                      }}
                    >
                      Pay and open channel
                    </LoadingButton>
                    {!payExternally && (
                      <Button
                        type="button"
                        variant="link"
                        className="text-muted-foreground text-xs"
                        onClick={() => setPayExternally(true)}
                      >
                        Pay with another wallet
                      </Button>
                    )}
                  </>
                )}

                {(payExternally || !canPayInternally) && (
                  <Payment
                    invoice={wrappedInvoiceResponse.invoice}
                    payment={
                      newChannel ? { preimage: "dummy preimage" } : undefined
                    }
                    paymentMethods="external"
                  />
                )}
              </>
            )}
          </div>
        </>
      )}
    </div>
  );
}
