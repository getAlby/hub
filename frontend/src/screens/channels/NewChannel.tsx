import { Box, Zap } from "lucide-react";
import React, { FormEvent } from "react";
import { Link, useNavigate } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import Loading from "src/components/Loading";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "src/components/ui/breadcrumb";
import { Button } from "src/components/ui/button";
import { Checkbox } from "src/components/ui/checkbox";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "src/components/ui/select";
import { useChannelPeerSuggestions } from "src/hooks/useChannelPeerSuggestions";
import { useInfo } from "src/hooks/useInfo";
import { usePeers } from "src/hooks/usePeers";
import { cn, formatAmount } from "src/lib/utils";
import useChannelOrderStore from "src/state/ChannelOrderStore";
import {
  Network,
  NewChannelOrder,
  Node,
  RecommendedChannelPeer,
} from "src/types";
import { request } from "src/utils/request";

function getPeerKey(peer: RecommendedChannelPeer) {
  return JSON.stringify(peer);
}

export default function NewChannel() {
  const { data: info } = useInfo();

  if (!info?.network) {
    return <Loading />;
  }

  return <NewChannelInternal network={info.network} />;
}

function NewChannelInternal({ network }: { network: Network }) {
  const { data: channelPeerSuggestions } = useChannelPeerSuggestions();
  const navigate = useNavigate();

  const [order, setOrder] = React.useState<Partial<NewChannelOrder>>({
    paymentMethod: "onchain",
    status: "pay",
  });

  const [selectedPeer, setSelectedPeer] = React.useState<
    RecommendedChannelPeer | undefined
  >();

  function setPaymentMethod(paymentMethod: "onchain" | "lightning") {
    setOrder({
      ...order,
      paymentMethod,
    });
  }

  const setAmount = React.useCallback((amount: string) => {
    setOrder((current) => ({
      ...current,
      amount,
    }));
  }, []);

  React.useEffect(() => {
    if (!channelPeerSuggestions) {
      return;
    }
    const recommendedPeer = channelPeerSuggestions.find(
      (peer) =>
        peer.network === network && peer.paymentMethod === order.paymentMethod
    );

    setSelectedPeer(recommendedPeer);
  }, [network, order.paymentMethod, channelPeerSuggestions]);

  React.useEffect(() => {
    if (selectedPeer) {
      if (
        selectedPeer.paymentMethod === "onchain" &&
        order.paymentMethod === "onchain"
      ) {
        setOrder((current) => ({
          ...current,
          pubkey: selectedPeer.pubkey,
          host: selectedPeer.host,
        }));
      }
      if (
        selectedPeer.paymentMethod === "lightning" &&
        order.paymentMethod === "lightning"
      ) {
        setOrder((current) => ({
          ...current,
          lsp: selectedPeer.lsp,
        }));
      }
      setAmount(selectedPeer.minimumChannelSize.toString());
    }
  }, [order.paymentMethod, selectedPeer, setAmount]);

  const selectedCardStyles = "border-primary border-2 font-medium";
  const presetAmounts = [250_000, 500_000, 1_000_000];

  function onSubmit(e: FormEvent) {
    e.preventDefault();
    useChannelOrderStore.getState().setOrder(order as NewChannelOrder);
    navigate("/channels/order");
  }

  if (!channelPeerSuggestions) {
    return <Loading />;
  }

  return (
    <>
      <Breadcrumb>
        <BreadcrumbList>
          <BreadcrumbItem>
            <BreadcrumbLink asChild>
              <Link to="/channels">Liquidity</Link>
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbPage>Open Channel</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <AppHeader
        title="Open a channel"
        description="Funds used to open a channel minus fees will be added to your spending balance"
      />
      <form
        onSubmit={onSubmit}
        className="md:max-w-md max-w-full flex flex-col gap-5"
      >
        <div className="grid gap-1.5">
          <Label htmlFor="amount">Channel size (sats)</Label>
          {order.amount && +order.amount < 200_000 && (
            <p className="text-muted-foreground text-xs">
              For a smooth experience consider a opening a channel of 200k sats
              in size or more.{" "}
              <ExternalLink
                to="https://guides.getalby.com/user-guide/v/alby-account-and-browser-extension/alby-hub/liquidity"
                className="underline"
              >
                Learn more
              </ExternalLink>
            </p>
          )}
          <Input
            id="amount"
            type="number"
            required
            min={selectedPeer?.minimumChannelSize || 100000}
            value={order.amount}
            onChange={(e) => {
              setAmount(e.target.value.trim());
            }}
          />
          <div className="grid grid-cols-3 gap-1.5 text-muted-foreground text-xs">
            {presetAmounts.map((amount) => (
              <div
                key={amount}
                className={cn(
                  "text-center border rounded p-2 cursor-pointer hover:border-muted-foreground",
                  +(order.amount || "0") === amount &&
                    "border-primary hover:border-primary"
                )}
                onClick={() => setAmount(amount.toString())}
              >
                {formatAmount(amount * 1000, 0)}
              </div>
            ))}
          </div>
        </div>
        <div className="grid gap-3">
          <Label htmlFor="amount">Payment method</Label>
          <div className="grid grid-cols-2 gap-3">
            <Link
              to="#"
              onClick={() => setPaymentMethod("onchain")}
              className="flex-1"
            >
              <div
                className={cn(
                  "rounded-xl border bg-card text-card-foreground shadow p-5 flex flex-col items-center gap-3",
                  order.paymentMethod === "onchain"
                    ? selectedCardStyles
                    : undefined
                )}
              >
                <Box className="w-4 h-4" />
                Onchain
              </div>
            </Link>
            <Link to="#" onClick={() => setPaymentMethod("lightning")}>
              <div
                className={cn(
                  "rounded-xl border bg-card text-card-foreground shadow p-5 flex flex-col items-center gap-3",
                  order.paymentMethod === "lightning"
                    ? selectedCardStyles
                    : undefined
                )}
              >
                <Zap className="w-4 h-4" />
                Lightning
              </div>
            </Link>
          </div>
        </div>
        <div className="flex flex-col gap-3">
          {selectedPeer &&
            (selectedPeer.paymentMethod === "lightning" ||
              (order.paymentMethod === "onchain" &&
                selectedPeer.pubkey === order.pubkey)) && (
              <div className="grid gap-1.5">
                <Label>Channel peer</Label>
                <Select
                  value={getPeerKey(selectedPeer)}
                  onValueChange={(value) =>
                    setSelectedPeer(
                      channelPeerSuggestions.find(
                        (x) => getPeerKey(x) === value
                      )
                    )
                  }
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select channel peer" />
                  </SelectTrigger>
                  <SelectContent>
                    {channelPeerSuggestions
                      .filter(
                        (peer) =>
                          peer.network === network &&
                          peer.paymentMethod === order.paymentMethod
                      )
                      .map((peer) => (
                        <SelectItem
                          value={getPeerKey(peer)}
                          key={getPeerKey(peer)}
                        >
                          <div className="flex items-center gap-3">
                            <div className="flex items-center gap-3">
                              {peer.name !== "Custom" && (
                                <img
                                  src={peer.image}
                                  className="w-8 h-8 object-contain"
                                />
                              )}
                              <div>
                                {peer.name}
                                {peer.minimumChannelSize > 0 && (
                                  <span className="ml-4 text-xs text-muted-foreground">
                                    Min.{" "}
                                    {new Intl.NumberFormat().format(
                                      peer.minimumChannelSize
                                    )}{" "}
                                    sats
                                  </span>
                                )}
                              </div>
                            </div>
                          </div>
                        </SelectItem>
                      ))}
                  </SelectContent>
                </Select>
                {selectedPeer.name === "Custom" && (
                  <>
                    <div className="grid gap-1.5"></div>
                  </>
                )}
              </div>
            )}
        </div>
        {order.paymentMethod === "onchain" && (
          <NewChannelOnchain
            order={order}
            setOrder={setOrder}
            showCustomOptions={selectedPeer?.name === "Custom"}
          />
        )}
        {order.paymentMethod === "lightning" && (
          <NewChannelLightning order={order} setOrder={setOrder} />
        )}
        <Button size="lg">Next</Button>
      </form>
    </>
  );
}

type NewChannelLightningProps = {
  order: Partial<NewChannelOrder>;
  setOrder(order: Partial<NewChannelOrder>): void;
};

function NewChannelLightning(props: NewChannelLightningProps) {
  if (props.order.paymentMethod !== "lightning") {
    throw new Error("unexpected payment method");
  }
  return null;
}

type NewChannelOnchainProps = {
  order: Partial<NewChannelOrder>;
  setOrder: React.Dispatch<React.SetStateAction<Partial<NewChannelOrder>>>;
  showCustomOptions: boolean;
};

function NewChannelOnchain(props: NewChannelOnchainProps) {
  const [nodeDetails, setNodeDetails] = React.useState<Node | undefined>();
  const { data: peers } = usePeers();
  // const { data: csrf } = useCSRF();
  if (props.order.paymentMethod !== "onchain") {
    throw new Error("unexpected payment method");
  }
  const { pubkey, host, isPublic } = props.order;
  const { setOrder } = props;
  const isAlreadyPeered =
    pubkey && peers?.some((peer) => peer.nodeId === pubkey);

  function setPubkey(pubkey: string) {
    props.setOrder((current) => ({
      ...current,
      paymentMethod: "onchain",
      pubkey,
    }));
  }
  const setHost = React.useCallback(
    (host: string) => {
      setOrder((current) => ({
        ...current,
        paymentMethod: "onchain",
        host,
      }));
    },
    [setOrder]
  );
  function setPublic(isPublic: boolean) {
    props.setOrder((current) => ({
      ...current,
      paymentMethod: "onchain",
      isPublic,
    }));
  }

  const fetchNodeDetails = React.useCallback(async () => {
    if (!pubkey) {
      setNodeDetails(undefined);
      return;
    }
    try {
      const data = await request<Node>(
        `/api/mempool?endpoint=/v1/lightning/nodes/${pubkey}`
      );

      setNodeDetails(data);
      const socketAddress = data?.sockets?.split(",")?.[0];
      if (socketAddress) {
        setHost(socketAddress);
      }
    } catch (error) {
      console.error(error);
      setNodeDetails(undefined);
    }
  }, [pubkey, setHost]);

  React.useEffect(() => {
    fetchNodeDetails();
  }, [fetchNodeDetails]);

  return (
    <>
      <div className="flex flex-col gap-5">
        {props.showCustomOptions && (
          <>
            <div className="grid gap-1.5">
              <Label htmlFor="pubkey">Peer</Label>
              <Input
                id="pubkey"
                type="text"
                value={pubkey}
                placeholder="Pubkey of the peer"
                onChange={(e) => {
                  setPubkey(e.target.value.trim());
                }}
              />
              {nodeDetails && (
                <div className="ml-2 text-muted-foreground text-sm">
                  <span
                    className="mr-2"
                    style={{ color: `${nodeDetails.color}` }}
                  >
                    ⬤
                  </span>
                  {nodeDetails.alias && (
                    <>
                      {nodeDetails.alias} ({nodeDetails.active_channel_count}{" "}
                      channels)
                    </>
                  )}
                </div>
              )}
            </div>

            {!isAlreadyPeered && /*!nodeDetails && */ pubkey && (
              <div className="grid gap-1.5">
                <Label htmlFor="host">Host:Port</Label>
                <Input
                  id="host"
                  type="text"
                  value={host}
                  placeholder="0.0.0.0:9735 or [2600::]:9735"
                  onChange={(e) => {
                    setHost(e.target.value.trim());
                  }}
                />
              </div>
            )}
          </>
        )}

        <div className="mt-2 flex items-top space-x-2">
          <Checkbox
            id="public-channel"
            defaultChecked={isPublic}
            onCheckedChange={() => setPublic(!isPublic)}
            className="mr-2"
          />
          <div className="grid gap-1.5 leading-none">
            <Label htmlFor="public-channel" className="flex items-center gap-2">
              Public Channel
            </Label>
            <p className="text-xs text-muted-foreground">
              Enable if you want to receive keysend payments. (e.g. podcasting)
            </p>
          </div>
        </div>
      </div>
    </>
  );
}
