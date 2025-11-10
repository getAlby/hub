import React from "react";
import {
  ConnectPeerRequest,
  MempoolUtxo,
  NewChannelOrder,
  OpenChannelRequest,
  OpenChannelResponse,
  PayInvoiceResponse,
} from "src/types";

import { CopyIcon, QrCodeIcon, RefreshCwIcon } from "lucide-react";
import { Link } from "react-router-dom";
import { toast } from "sonner";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
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
import { LoadingButton } from "src/components/ui/custom/loading-button";
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
import { Separator } from "src/components/ui/separator";
import { Table, TableBody, TableCell, TableRow } from "src/components/ui/table";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "src/components/ui/tooltip";
import { useBalances } from "src/hooks/useBalances";

import { ChannelWaitingForConfirmations } from "src/components/channels/ChannelWaitingForConfirmations";
import { PayLightningInvoice } from "src/components/PayLightningInvoice";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { useChannels } from "src/hooks/useChannels";
import { useMempoolApi } from "src/hooks/useMempoolApi";
import { useNodeDetails } from "src/hooks/useNodeDetails";
import { useOnchainAddress } from "src/hooks/useOnchainAddress";
import { usePeers } from "src/hooks/usePeers";
import { useSyncWallet } from "src/hooks/useSyncWallet";
import { copyToClipboard } from "src/lib/clipboard";
import { splitSocketAddress } from "src/lib/utils";
import useChannelOrderStore from "src/state/ChannelOrderStore";
import { LSPOrderRequest, LSPOrderResponse } from "src/types";
import { request } from "src/utils/request";

// ensures React does not open a duplicate channel
// this is a hack and will break if the user tries to open
// 2 outbound channels without refreshing the page (I think an edge case)
let hasStartedOpenedChannel = false;

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
    case "paid":
      // LSPS1 only
      return <PaidLightningChannelOrder />;
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

function Success() {
  return (
    <div className="flex flex-col justify-center gap-5 p-5 max-w-md items-stretch">
      <TwoColumnLayoutHeader
        title="Channel Opened"
        description="Your new lightning channel is ready to use"
      />

      <p>
        Congratulations! Your channel is active and can be used to send and
        receive payments.
      </p>
      <p>
        To ensure you can both send and receive, make sure to balance your{" "}
        <ExternalLink
          to="https://guides.getalby.com/user-guide/alby-hub/node"
          className="underline"
        >
          channel's liquidity
        </ExternalLink>
        .
      </p>

      <Link to="/home" className="flex justify-center mt-8">
        <Button>Go to your dashboard</Button>
      </Link>
    </div>
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

  if (!channel) {
    return <Loading />;
  }

  return <ChannelWaitingForConfirmations channel={channel} />;
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

  if (!balances) {
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

  const { data: mempoolAddressUtxos } = useMempoolApi<MempoolUtxo[]>(
    onchainAddress ? `/address/${onchainAddress}/utxo` : undefined,
    3000
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
          <p className="text-xs slashed-zero">
            You currently have{" "}
            <span className="font-semibold sensitive">
              <FormattedBitcoinAmount amount={balances.onchain.total * 1000} />
            </span>
            . We recommend depositing an additional amount of{" "}
            <span className="font-semibold">
              <FormattedBitcoinAmount amount={recommendedAmount * 1000} />
            </span>{" "}
            to open this channel.
          </p>
          <p className="text-xs text-muted-foreground">
            This amount includes cost for the channel opening and potential
            channel onchain reserves.
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
              <CopyIcon className="size-4" />
            </Button>
            <Dialog>
              <DialogTrigger asChild>
                <Button variant="secondary" size="icon">
                  <QrCodeIcon className="size-4" />
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
                    {!loadingAddress && <RefreshCwIcon className="size-4" />}
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
            <CardContent className="slashed-zero">
              <FormattedBitcoinAmount amount={unspentAmount * 1000} /> deposited
            </CardContent>
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

  const { pubkey, host } = order;

  const { data: nodeDetails } = useNodeDetails(pubkey);

  const connectPeer = React.useCallback(async () => {
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
        "Content-Type": "application/json",
      },
      body: JSON.stringify(connectPeerRequest),
    });
  }, [nodeDetails, pubkey, host]);

  const openChannel = React.useCallback(async () => {
    try {
      if (order.paymentMethod !== "onchain") {
        throw new Error("incorrect payment method");
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
        amountSats: +order.amount,
        public: order.isPublic,
      };
      const openChannelResponse = await request<OpenChannelResponse>(
        "/api/channels",
        {
          method: "POST",
          headers: {
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
      toast("Successfully published channel opening transaction");
      useChannelOrderStore.getState().updateOrder({
        fundingTxId: openChannelResponse.fundingTxId,
        status: "opening",
      });
    } catch (error) {
      console.error(error);
      toast.error("Something went wrong", {
        description: "" + error,
      });
    }
  }, [
    connectPeer,
    order.amount,
    order.isPublic,
    order.paymentMethod,
    peers,
    pubkey,
  ]);

  React.useEffect(() => {
    if (!peers || hasStartedOpenedChannel) {
      return;
    }

    hasStartedOpenedChannel = true;
    openChannel();
  }, [openChannel, order.amount, peers, pubkey]);

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

function useWaitForNewChannel() {
  const order = useChannelOrderStore((store) => store.order);
  const { data: channels } = useChannels(true);

  const newChannel =
    channels && order?.prevChannelIds
      ? channels.find(
          (newChannel) =>
            !order.prevChannelIds.some(
              (current) => newChannel.id === current
            ) && newChannel.fundingTxId
        )
      : undefined;

  React.useEffect(() => {
    if (newChannel) {
      (async () => {
        toast("Successfully opened channel");
        setTimeout(() => {
          useChannelOrderStore.getState().updateOrder({
            status: "opening",
            fundingTxId: newChannel.fundingTxId,
          });
        }, 3000);
      })();
    }
  }, [newChannel]);
}

function PaidLightningChannelOrder() {
  useWaitForNewChannel();

  return (
    <div className="flex w-full h-full gap-2 items-center justify-center">
      <Loading /> <p>Waiting for channel to be opened...</p>
    </div>
  );
}

function PayLightningChannelOrder({ order }: { order: NewChannelOrder }) {
  if (order.paymentMethod !== "lightning") {
    throw new Error("incorrect payment method");
  }

  const { data: channels } = useChannels(true);
  const [, setRequestedInvoice] = React.useState(false);

  const [lspOrderResponse, setLspOrderResponse] = React.useState<
    LSPOrderResponse | undefined
  >();

  useWaitForNewChannel();

  React.useEffect(() => {
    if (!channels) {
      return;
    }
    setRequestedInvoice((current) => {
      if (!current) {
        (async () => {
          try {
            if (!order.lspType || !order.lspIdentifier) {
              throw new Error("missing lsp info in order");
            }
            const newLSPOrderRequest: LSPOrderRequest = {
              lspType: order.lspType,
              lspIdentifier: order.lspIdentifier,
              amount: parseInt(order.amount),
              public: order.isPublic,
            };
            const response = await request<LSPOrderResponse>(
              "/api/lsp-orders",
              {
                method: "POST",
                headers: {
                  "Content-Type": "application/json",
                },
                body: JSON.stringify(newLSPOrderRequest),
              }
            );
            if (!response) {
              throw new Error("no LSP order response");
            }

            if (!response.invoice) {
              // assume payment is handled by Alby Account
              // we will wait for a channel to be opened to us
              useChannelOrderStore.getState().updateOrder({
                status: "paid",
              });
            }
            setLspOrderResponse(response);
          } catch (error) {
            toast.error("Something went wrong", {
              description: "" + error,
            });
          }
        })();
      }
      return true;
    });
  }, [
    channels,
    order.amount,
    order.isPublic,
    order.lspType,
    order.lspIdentifier,
  ]);

  const canPayInternally =
    channels &&
    lspOrderResponse &&
    channels.some(
      (channel) =>
        channel.localSpendableBalance / 1000 > lspOrderResponse.invoiceAmount
    );
  const [isPaying, setPaying] = React.useState(false);
  const [payExternally, setPayExternally] = React.useState(false);

  return (
    <div className="flex flex-col gap-5">
      <AppHeader
        title={"Buy Channel"}
        description={
          lspOrderResponse
            ? "Complete Payment to open a channel to your node"
            : "Please wait, loading..."
        }
      />
      {!lspOrderResponse?.invoice && <Loading />}

      {lspOrderResponse?.invoice && (
        <>
          <div className="max-w-md flex flex-col gap-5">
            <div className="border rounded-lg slashed-zero">
              <Table>
                <TableBody>
                  {lspOrderResponse.outgoingLiquidity > 0 && (
                    <TableRow>
                      <TableCell className="font-medium p-3">
                        Spending Balance
                      </TableCell>
                      <TableCell className="text-right p-3">
                        <FormattedBitcoinAmount
                          amount={lspOrderResponse.outgoingLiquidity * 1000}
                        />
                      </TableCell>
                    </TableRow>
                  )}
                  {lspOrderResponse.incomingLiquidity > 0 && (
                    <TableRow>
                      <TableCell className="font-medium p-3">
                        Incoming Liquidity
                      </TableCell>
                      <TableCell className="text-right p-3">
                        <FormattedBitcoinAmount
                          amount={lspOrderResponse.incomingLiquidity * 1000}
                        />
                      </TableCell>
                    </TableRow>
                  )}
                  <TableRow>
                    <TableCell className="font-medium p-3">
                      Amount to pay
                    </TableCell>
                    <TableCell className="font-semibold text-right p-3">
                      <FormattedBitcoinAmount
                        amount={lspOrderResponse.invoiceAmount * 1000}
                      />
                    </TableCell>
                  </TableRow>
                </TableBody>
              </Table>
            </div>
            <>
              {canPayInternally && (
                <>
                  <LoadingButton
                    loading={isPaying}
                    className="mt-4"
                    onClick={async () => {
                      try {
                        setPaying(true);

                        await request<PayInvoiceResponse>(
                          `/api/payments/${lspOrderResponse.invoice}`,
                          {
                            method: "POST",
                            headers: {
                              "Content-Type": "application/json",
                            },
                          }
                        );

                        useChannelOrderStore.getState().updateOrder({
                          status: "paid",
                        });
                        toast("Channel successfully requested");
                      } catch (e) {
                        toast.error("Failed to send: ", {
                          description: "" + e,
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
                <PayLightningInvoice invoice={lspOrderResponse.invoice} />
              )}

              <div className="flex-1 flex flex-col justify-end items-center gap-4">
                <Separator className="my-16" />
                <p className="text-sm text-muted-foreground text-center">
                  Other options
                </p>
                <Link to="/channels/outgoing" className="w-full">
                  <Button className="w-full" variant="secondary">
                    Increase Spending Balance
                  </Button>
                </Link>
                <ExternalLink
                  to="https://www.getalby.com/topup"
                  className="w-full"
                >
                  <Button className="w-full" variant="secondary">
                    Buy Bitcoin
                  </Button>
                </ExternalLink>
              </div>
            </>
          </div>
        </>
      )}
    </div>
  );
}
