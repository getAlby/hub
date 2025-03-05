import { ChevronDown, ChevronUp, InfoIcon, RefreshCw } from "lucide-react";
import React, { useState } from "react";
import { Link, useLocation } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import { ChannelPeerNote } from "src/components/channels/ChannelPeerNote";
import { ChannelPublicPrivateAlert } from "src/components/channels/ChannelPublicPrivateAlert";
import { DuplicateChannelAlert } from "src/components/channels/DuplicateChannelAlert";
import ExternalLink from "src/components/ExternalLink";
import Loading from "src/components/Loading";
import { MempoolAlert } from "src/components/MempoolAlert";
import { Button } from "src/components/ui/button";
import { Card } from "src/components/ui/card";
import { Checkbox } from "src/components/ui/checkbox";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "src/components/ui/sheet";
import StepButtons from "src/components/ui/stepButtons";
import { Step, StepItem, Stepper } from "src/components/ui/Stepper";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "src/components/ui/tooltip";
import { useToast } from "src/components/ui/use-toast";
import { useChannelPeerSuggestions } from "src/hooks/useChannelPeerSuggestions";
import { useChannels } from "src/hooks/useChannels";
import { useInfo } from "src/hooks/useInfo";
import { cn, formatAmount } from "src/lib/utils";
import { CurrentChannelOrder } from "src/screens/channels/CurrentChannelOrder";
import useChannelOrderStore from "src/state/ChannelOrderStore";
import {
  Channel,
  LightningOrder,
  Network,
  NewChannelOrder,
  RecommendedChannelPeer,
} from "src/types";

function getPeerKey(peer: RecommendedChannelPeer) {
  return JSON.stringify(peer);
}

export default function IncreaseIncomingCapacity() {
  const { data: info } = useInfo();
  const { data: channels } = useChannels();

  if (!info?.network || !channels) {
    return <Loading />;
  }

  return <NewChannelInternal network={info.network} channels={channels} />;
}

function NewChannelInternal({
  network,
  channels,
}: {
  network: Network;
  channels: Channel[];
}) {
  const { data: _channelPeerSuggestions } = useChannelPeerSuggestions();
  const location = useLocation();

  const steps = [
    { label: "configureChannel" },
    { label: "openChannel" },
  ] satisfies StepItem[];

  const { toast } = useToast();

  const [ChannelPartnersMenuOpen, setChannelPartnersMenuOpen] = useState(false);

  React.useEffect(() => {
    setChannelPartnersMenuOpen(false);
  }, [location]);

  const presetAmounts = [1_000_000, 2_000_000, 3_000_000];

  const [order, setOrder] = React.useState<Partial<LightningOrder>>({
    paymentMethod: "lightning",
    status: "pay",
    amount: presetAmounts[0].toString(),
    prevChannelIds: channels.map((channel) => channel.id),
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
          ...(!selectedPeer.publicChannelsAllowed && { isPublic: false }),
        }));
      }
    }
  }, [order.paymentMethod, selectedPeer]);

  function onSubmit() {
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
            amount <= partner.maximumChannelSize &&
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
        title="Open Channel with Lightning"
        description="Increase your receive limit by opening a new lightning channel for a small fee."
        contentRight={
          <div className="flex items-end">
            <Link
              to="/channels/outgoing"
              className="underline break-words text-sm"
            >
              Open Channel with On-chain
            </Link>
          </div>
        }
      />
      <Stepper initialStep={0} steps={steps} orientation="vertical">
        <Step key="configureChannel" label="Configure Channel">
          <div className="md:max-w-md max-w-full flex flex-col gap-5 flex-1">
            <p className="text-muted-foreground">
              Alby Hub works with selected service providers (LSPs) which
              provide the best network connectivity and liquidity to receive
              payments. The channel typically stays open as long as there is
              usage.{" "}
              <ExternalLink
                className="underline"
                to="https://guides.getalby.com/user-guide/alby-account-and-browser-extension/alby-hub/faq-alby-hub/how-to-open-a-channel"
              >
                Learn more
              </ExternalLink>
            </p>
            <div className="flex flex-col gap-5 flex-1">
              <div className="grid gap-2">
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger type="button">
                      <div className="flex flex-row gap-2 items-center justify-start text-sm">
                        <Label htmlFor="amount">Increase receive limit</Label>
                        <InfoIcon className="h-4 w-4 shrink-0 text-muted-foreground" />
                      </div>
                    </TooltipTrigger>
                    <TooltipContent className="w-[300px]">
                      Configure the amount of receiving capacity you need. You
                      will only pay for the liquidity fee which will be shown in
                      the next step.
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>

                {order.amount && +order.amount < 200_000 && (
                  <p className="text-muted-foreground text-xs">
                    For a smooth experience consider a opening a channel of 200k
                    sats in size or more.{" "}
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
              <div
                className="flex gap-1 cursor-pointer items-center"
                onClick={() => setShowAdvanced((current) => !current)}
              >
                Advanced
                {showAdvanced ? (
                  <ChevronUp className="w-4 h-4" />
                ) : (
                  <ChevronDown className="w-4 h-4" />
                )}
              </div>
              {showAdvanced && (
                <>
                  <div className="flex flex-col gap-3">
                    {selectedPeer && (
                      <div className="grid gap-1">
                        <p className="text-sm font-medium">Channel peer</p>
                        <Card className="p-4 shadow-none">
                          <div className="flex items-center gap-3 justify-between">
                            <div className="flex items-center gap-3">
                              {selectedPeer.name !== "Custom" && (
                                <img
                                  src={selectedPeer.image}
                                  className="w-8 h-8 object-contain"
                                />
                              )}
                              <div className="flex flex-col gap-1">
                                <p className="font-semibold">
                                  {selectedPeer.name}
                                </p>
                                <span className="text-sm slashed-zero">
                                  <span className="text-muted-foreground">
                                    Min. channel size:
                                  </span>
                                  {new Intl.NumberFormat().format(
                                    selectedPeer.minimumChannelSize
                                  )}{" "}
                                  sats
                                </span>
                              </div>
                            </div>
                            <Sheet
                              open={ChannelPartnersMenuOpen}
                              onOpenChange={setChannelPartnersMenuOpen}
                            >
                              <SheetTrigger>
                                <Button
                                  variant="secondary"
                                  type="button"
                                  className="flex gap-2 items-center justify-center"
                                >
                                  <RefreshCw className="w-4 h-4 mr-2" />
                                  Change
                                </Button>
                              </SheetTrigger>
                              <SheetContent>
                                <SheetHeader>
                                  <SheetTitle>Change channel peer</SheetTitle>
                                  <SheetDescription>
                                    <div className="grid gap-4">
                                      {channelPeerSuggestions
                                        .filter(
                                          (peer) =>
                                            peer.network === network &&
                                            peer.paymentMethod ===
                                              order.paymentMethod
                                        )
                                        .map((peer) => (
                                          <Card
                                            key={getPeerKey(peer)}
                                            className={`p-4 shadow-none cursor-pointer ${
                                              getPeerKey(selectedPeer) ===
                                              getPeerKey(peer)
                                                ? "border-primary"
                                                : ""
                                            }`}
                                            onClick={() =>
                                              setSelectedPeer(peer)
                                            }
                                          >
                                            <div className="flex items-center gap-3 justify-between">
                                              <div className="flex items-center gap-3">
                                                {peer.name !== "Custom" && (
                                                  <img
                                                    src={peer.image}
                                                    className="w-8 h-8 object-contain"
                                                    alt={peer.name}
                                                  />
                                                )}
                                                <div className="flex flex-col gap-1">
                                                  <p className="font-semibold">
                                                    {peer.name}
                                                  </p>
                                                  <span className="text-sm slashed-zero">
                                                    <span className="text-muted-foreground">
                                                      Min. channel size:{" "}
                                                    </span>
                                                    {new Intl.NumberFormat().format(
                                                      peer.minimumChannelSize
                                                    )}{" "}
                                                    sats
                                                  </span>
                                                </div>
                                              </div>
                                            </div>
                                          </Card>
                                        ))}
                                    </div>
                                  </SheetDescription>
                                </SheetHeader>
                              </SheetContent>
                            </Sheet>
                          </div>
                        </Card>
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
                      disabled={
                        selectedPeer && !selectedPeer.publicChannelsAllowed
                      }
                      title={
                        selectedPeer && !selectedPeer.publicChannelsAllowed
                          ? "This channel partner does not support public channels."
                          : undefined
                      }
                    />
                    <div className="grid gap-1 leading-none">
                      <Label
                        htmlFor="public-channel"
                        className="flex items-center gap-2"
                      >
                        Public Channel
                      </Label>
                      <p className="text-xs text-muted-foreground">
                        Enable if you need keysend payments, mostly applicable
                        for podcasters. Otherwise, it’s not recommended.
                      </p>
                    </div>
                  </div>
                </>
              )}
              {channels?.some(
                (channel) => channel.public !== !!order.isPublic
              ) && <ChannelPublicPrivateAlert />}
              {selectedPeer?.note && <ChannelPeerNote peer={selectedPeer} />}

              <DuplicateChannelAlert
                pubkey={selectedPeer?.pubkey}
                name={selectedPeer?.name}
              />

              <MempoolAlert />

              <StepButtons onNextClick={() => onSubmit()} />
            </div>
          </div>
        </Step>

        <Step key="openChannel" label="Open Channel">
          <CurrentChannelOrder />
          <StepButtons />
        </Step>
      </Stepper>
    </>
  );
}

type NewChannelLightningProps = {
  order: Partial<LightningOrder>;
  setOrder(order: Partial<LightningOrder>): void;
};

function NewChannelLightning(props: NewChannelLightningProps) {
  if (props.order.paymentMethod !== "lightning") {
    throw new Error("unexpected payment method");
  }
  return null;
}
