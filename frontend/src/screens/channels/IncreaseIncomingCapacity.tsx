import { ChevronDown } from "lucide-react";
import React, { FormEvent } from "react";
import { useNavigate } from "react-router-dom";
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
import { useToast } from "src/components/ui/use-toast";
import { useChannelPeerSuggestions } from "src/hooks/useChannelPeerSuggestions";
import { useChannels } from "src/hooks/useChannels";
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
  const { data: channels } = useChannels();
  const { toast } = useToast();

  const presetAmounts = [1_000_000, 2_000_000, 3_000_000];

  const [order, setOrder] = React.useState<Partial<NewChannelOrder>>({
    paymentMethod: "lightning",
    status: "pay",
    amount: presetAmounts[0].toString(),
  });

  const [showAdvanced, setShowAdvanced] = React.useState(false);

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
    }
  }, [order.paymentMethod, selectedPeer]);

  function onSubmit(e: FormEvent) {
    e.preventDefault();
    try {
      if (!showAdvanced) {
        if (!channelPeerSuggestions) {
          throw new Error("Channel Peer suggestions not loaded");
        }
        if (!channels) {
          throw new Error("Channels not loaded");
        }
        const amount = parseInt(order.amount || "0");
        if (!amount) {
          throw new Error("No amount set");
        }

        // find the best channel partner
        const okPartners = channelPeerSuggestions.filter(
          (partner) =>
            amount >= partner.minimumChannelSize &&
            partner.network === network &&
            partner.paymentMethod === "lightning" &&
            partner.lspType === "LSPS1" &&
            partner.pubkey &&
            !channels.some((channel) => channel.remotePubkey === partner.pubkey)
        );

        const partner = okPartners[0];
        if (!partner) {
          toast({
            description:
              "No ideal channel partner found. Please choose from the advanced options to continue",
          });
          return;
        }
        order.paymentMethod = "lightning";
        if (
          order.paymentMethod !== "lightning" ||
          partner.paymentMethod !== "lightning"
        ) {
          throw new Error("Unexpected order or partner payment method");
        }
        order.lspType = partner.lspType;
        order.lspUrl = partner.lspUrl;
      }

      useChannelOrderStore.getState().setOrder(order as NewChannelOrder);
      navigate("/channels/order");
    } catch (error) {
      toast({
        variant: "destructive",
        description: "Something went wrong: " + error,
      });
      console.error(error);
    }
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
            min={
              showAdvanced
                ? selectedPeer?.minimumChannelSize || 100000
                : undefined
            }
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
        {showAdvanced && (
          <>
            <div className="flex flex-col gap-3">
              {selectedPeer && (
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
                <Label
                  htmlFor="public-channel"
                  className="flex items-center gap-2"
                >
                  Public Channel
                </Label>
                <p className="text-xs text-muted-foreground">
                  Enable if you want to receive keysend payments. (e.g.
                  podcasting)
                </p>
              </div>
            </div>
          </>
        )}
        {!showAdvanced && (
          <Button
            type="button"
            variant="link"
            className="text-muted-foreground text-xs"
            onClick={() => setShowAdvanced((current) => !current)}
          >
            <ChevronDown className="w-4 h-4 mr-2" />
            Advanced Options
          </Button>
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
