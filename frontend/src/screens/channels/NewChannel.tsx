import { Box, Zap } from "lucide-react";
import React, { FormEvent } from "react";
import { Link, useNavigate } from "react-router-dom";
import albyImage from "src/assets/images/peers/alby.svg";
import mutinynetImage from "src/assets/images/peers/mutinynet.jpeg";
import olympusImage from "src/assets/images/peers/olympus.svg";
import voltageImage from "src/assets/images/peers/voltage.webp";
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
import { useInfo } from "src/hooks/useInfo";
import { cn, formatAmount } from "src/lib/utils";
import useChannelOrderStore from "src/state/ChannelOrderStore";
import { Network, NewChannelOrder, Node } from "src/types";
import { request } from "src/utils/request";

type RecommendedPeer = {
  network: Network;
  image: string;
  name: string;
  minimumChannelSize: number;
} & (
  | {
      paymentMethod: "onchain";
      pubkey: string;
      host: string;
    }
  | {
      paymentMethod: "lightning";
      lsp: string;
    }
);

const recommendedPeers: RecommendedPeer[] = [
  {
    paymentMethod: "onchain",
    network: "bitcoin",
    pubkey:
      "029ca15ad2ea3077f5f0524c4c9bc266854c14b9fc81b9cc3d6b48e2460af13f65",
    host: "141.95.84.44:9735",
    minimumChannelSize: 250_000,
    name: "Alby",
    image: albyImage,
  },
  {
    paymentMethod: "onchain",
    network: "testnet",
    pubkey:
      "030f8fcc69816d90445f450e59304171fd805f4395a0f4950a5956ce3300463f5a",
    host: "209.38.178.74:9735",
    minimumChannelSize: 50_000,
    name: "Alby Testnet LND2",
    image: albyImage,
  },
  {
    paymentMethod: "onchain",
    network: "signet",
    pubkey:
      "02465ed5be53d04fde66c9418ff14a5f2267723810176c9212b722e542dc1afb1b",
    host: "45.79.52.207:9735",
    minimumChannelSize: 50_000,
    name: "Mutinynet Faucet",
    image: mutinynetImage,
  },
  {
    paymentMethod: "lightning",
    network: "signet",
    lsp: "OLYMPUS_MUTINYNET_LSPS1",
    minimumChannelSize: 1_000_000,
    name: "Olympus Mutinynet (LSPS1)",
    image: olympusImage,
  },
  {
    paymentMethod: "lightning",
    network: "bitcoin",
    lsp: "OLYMPUS_FLOW_2_0",
    minimumChannelSize: 20_000,
    name: "Olympus (Flow 2.0)",
    image: olympusImage,
  },
  {
    paymentMethod: "lightning",
    network: "bitcoin",
    lsp: "VOLTAGE",
    minimumChannelSize: 20_000,
    name: "Voltage (Flow 2.0)",
    image: voltageImage,
  },
  {
    paymentMethod: "lightning",
    network: "signet",
    lsp: "OLYMPUS_MUTINYNET_FLOW_2_0",
    minimumChannelSize: 20_000,
    name: "Olympus Mutinynet (Flow 2.0)",
    image: olympusImage,
  },
  {
    paymentMethod: "lightning",
    network: "signet",
    lsp: "ALBY_MUTINYNET",
    minimumChannelSize: 150_000,
    name: "Alby Mutinynet",
    image: albyImage,
  },
  {
    network: "signet",
    paymentMethod: "onchain",
    pubkey: "",
    host: "",
    minimumChannelSize: 0,
    name: "Custom",
    image: albyImage,
  },
];

export default function NewChannel() {
  const { data: info } = useInfo();

  if (!info?.network) {
    return <Loading />;
  }

  return <NewChannelInternal network={info.network} />;
}

function NewChannelInternal({ network }: { network: Network }) {
  const navigate = useNavigate();

  const [order, setOrder] = React.useState<Partial<NewChannelOrder>>({
    paymentMethod: "onchain",
    status: "pay",
  });

  const [selectedPeer, setSelectedPeer] = React.useState<
    RecommendedPeer | undefined
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
    const recommendedPeer = recommendedPeers.find(
      (peer) =>
        peer.network === network && peer.paymentMethod === order.paymentMethod
    );

    setSelectedPeer(recommendedPeer);
  }, [network, order.paymentMethod]);

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
                to="https://guides.getalby.com/user-guide/v/alby-account-and-browser-extension/alby-lightning-account/faqs-alby-account"
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
            min={selectedPeer?.minimumChannelSize}
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
                  value={selectedPeer.name}
                  onValueChange={(value) =>
                    setSelectedPeer(
                      recommendedPeers.find((x) => x.name === value)
                    )
                  }
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select channel peer" />
                  </SelectTrigger>
                  <SelectContent>
                    {recommendedPeers
                      .filter(
                        (peer) =>
                          peer.network === network &&
                          peer.paymentMethod === order.paymentMethod
                      )
                      .map((peer) => (
                        <SelectItem value={peer.name} key={peer.name}>
                          <div className="flex items-center space-between gap-3 w-full">
                            <div className="flex items-center gap-3">
                              <img
                                src={peer.image}
                                className="w-12 h-12 object-contain"
                              />
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
  setOrder(order: Partial<NewChannelOrder>): void;
  showCustomOptions: boolean;
};

function NewChannelOnchain(props: NewChannelOnchainProps) {
  const [nodeDetails, setNodeDetails] = React.useState<Node | undefined>();

  // const { data: csrf } = useCSRF();
  if (props.order.paymentMethod !== "onchain") {
    throw new Error("unexpected payment method");
  }
  const { pubkey, host, isPublic } = props.order;

  function setPubkey(pubkey: string) {
    props.setOrder({
      ...props.order,
      paymentMethod: "onchain",
      pubkey,
    });
  }
  function setHost(host: string) {
    props.setOrder({
      ...props.order,
      paymentMethod: "onchain",
      host,
    });
  }
  function setPublic(isPublic: boolean) {
    props.setOrder({
      ...props.order,
      paymentMethod: "onchain",
      isPublic,
    });
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
    } catch (error) {
      console.error(error);
      setNodeDetails(undefined);
    }
  }, [pubkey]);

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

            {!nodeDetails && pubkey && (
              <div className="grid gap-1.5">
                <Label htmlFor="host">Host:Port</Label>
                <Input
                  id="host"
                  type="text"
                  value={host}
                  placeholder="0.0.0.0:9735"
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
