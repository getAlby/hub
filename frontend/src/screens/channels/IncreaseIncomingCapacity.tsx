import { Box, Zap } from "lucide-react";
import React, { FormEvent } from "react";
import { Link, useNavigate } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import Loading from "src/components/Loading";
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
import { cn, formatAmount } from "src/lib/utils";
import useChannelOrderStore from "src/state/ChannelOrderStore";
import { Network, NewChannelOrder, RecommendedChannelPeer } from "src/types";

function getPeerKey(peer: RecommendedChannelPeer) {
  return JSON.stringify(peer);
}

export default function IncreaseIncomingCapacity() {
  const { data: info } = useInfo();

  if (!info?.network) {
    return <Loading />;
  }

  return <NewChannelInternal network={info.network} />;
}

function NewChannelInternal({ network }: { network: Network }) {
  const { data: _channelPeerSuggestions } = useChannelPeerSuggestions();
  const navigate = useNavigate();

  const [order, setOrder] = React.useState<Partial<NewChannelOrder>>({
    paymentMethod: "lightning",
    status: "pay",
  });

  const [selectedPeer, setSelectedPeer] = React.useState<
    RecommendedChannelPeer | undefined
  >();

  const channelPeerSuggestions = React.useMemo(() => {
    return _channelPeerSuggestions
      ? [
          ..._channelPeerSuggestions.filter(
            (peer) =>
              peer.paymentMethod === "lightning" && peer.lspType === "LSPS1"
          ),
        ]
      : undefined;
  }, [_channelPeerSuggestions]);

  function setPublic(isPublic: boolean) {
    setOrder((current) => ({
      ...current,
      isPublic,
    }));
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
        selectedPeer.paymentMethod === "lightning" &&
        order.paymentMethod === "lightning"
      ) {
        setOrder((current) => ({
          ...current,
          lspType: selectedPeer.lspType,
          lspUrl: selectedPeer.lspUrl,
        }));
      }
      setAmount(selectedPeer.minimumChannelSize.toString());
    }
  }, [order.paymentMethod, selectedPeer, setAmount]);

  const selectedCardStyles = "border-primary border-2 font-medium";
  const presetAmounts = [1_000_000, 2_000_000, 3_000_000];

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
      <AppHeader
        title="Increase Receiving Capacity"
        description="Purchase a channel with incoming capacity to receive payments"
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
            <div
              //onClick={() => setPaymentMethod("onchain")}
              className="flex-1 opacity-50 select-none"
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
            </div>
            <Link
              to="#"
              //onClick={() => setPaymentMethod("lightning")}
            >
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
          {selectedPeer && (
            <div className="grid gap-1.5">
              <Label>Channel peer</Label>
              <Select
                value={getPeerKey(selectedPeer)}
                onValueChange={(value) =>
                  setSelectedPeer(
                    channelPeerSuggestions.find((x) => getPeerKey(x) === value)
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
        {order.paymentMethod === "lightning" && (
          <NewChannelLightning order={order} setOrder={setOrder} />
        )}

        <div className="mt-2 flex items-top space-x-2">
          <Checkbox
            id="public-channel"
            defaultChecked={order.isPublic}
            onCheckedChange={() => setPublic(!order.isPublic)}
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
