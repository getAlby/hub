import { InfoIcon } from "lucide-react";
import React, { FormEvent } from "react";
import { Link, useNavigate } from "react-router";
import { toast } from "sonner";
import AppHeader from "src/components/AppHeader";
import { ChannelPeerNote } from "src/components/channels/ChannelPeerNote";
import { ChannelPublicPrivateAlert } from "src/components/channels/ChannelPublicPrivateAlert";
import { DuplicateChannelAlert } from "src/components/channels/DuplicateChannelAlert";
import { SwapAlert } from "src/components/channels/SwapAlert";
import ExternalLink from "src/components/ExternalLink";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import Loading from "src/components/Loading";
import { MempoolAlert } from "src/components/MempoolAlert";
import { Alert, AlertDescription } from "src/components/ui/alert";
import { Button } from "src/components/ui/button";
import { Checkbox } from "src/components/ui/checkbox";
import { ExternalLinkButton } from "src/components/ui/custom/external-link-button";
import { LinkButton } from "src/components/ui/custom/link-button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "src/components/ui/dialog";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "src/components/ui/select";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "src/components/ui/tooltip";
import { useBalances } from "src/hooks/useBalances";
import { useChannelPeerSuggestions } from "src/hooks/useChannelPeerSuggestions";
import { useChannels } from "src/hooks/useChannels";
import { useInfo } from "src/hooks/useInfo";
import { usePeers } from "src/hooks/usePeers";
import { cn, formatAmount } from "src/lib/utils";
import useChannelOrderStore from "src/state/ChannelOrderStore";
import {
  Channel,
  Network,
  NewChannelOrder,
  OnchainOrder,
  RecommendedChannelPeer,
} from "src/types";

import LightningNetworkDarkSVG from "public/images/illustrations/lightning-network-dark.svg";
import LightningNetworkLightSVG from "public/images/illustrations/lightning-network-light.svg";
import { useNodeDetails } from "src/hooks/useNodeDetails";

function getPeerKey(peer: RecommendedChannelPeer) {
  return JSON.stringify(peer);
}

export default function IncreaseOutgoingCapacity() {
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
  const { data: balances } = useBalances();

  const navigate = useNavigate();

  const presetAmounts = [250_000, 500_000, 1_000_000];

  const [order, setOrder] = React.useState<Partial<OnchainOrder>>({
    paymentMethod: "onchain",
    status: "pay",
    amountSat: presetAmounts[0].toString(),
    isPublic: !!channels.length && channels.every((channel) => channel.public),
  });

  const [selectedPeer, setSelectedPeer] = React.useState<
    RecommendedChannelPeer | undefined
  >();

  const [showConfirmModal, setShowConfirmModal] = React.useState(false);

  const channelPeerSuggestions = React.useMemo(() => {
    const customOption: RecommendedChannelPeer = {
      name: "Custom",
      network,
      paymentMethod: "onchain",
      minimumChannelSizeSat: 0,
      maximumChannelSizeSat: 0,
      description: "",
      pubkey: "",
      host: "",
      image: "",
      note: "",
      publicChannelsAllowed: true,
    };
    return _channelPeerSuggestions
      ? [
          ..._channelPeerSuggestions.filter(
            (peer) => peer.paymentMethod !== "lightning"
          ),
          customOption,
        ]
      : [customOption];
  }, [_channelPeerSuggestions, network]);

  function setPublic(isPublic: boolean) {
    setOrder((current) => ({
      ...current,
      isPublic,
    }));
  }

  const setAmountSat = React.useCallback((amountSat: string) => {
    setOrder((current) => ({
      ...current,
      amountSat,
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
          ...(!selectedPeer.publicChannelsAllowed && { isPublic: false }),
        }));
      }
    }
  }, [order.paymentMethod, selectedPeer]);

  function onSubmit(e: FormEvent) {
    e.preventDefault();
    setShowConfirmModal(true);
  }

  function handleConfirmSubmit() {
    try {
      if (!channels) {
        throw new Error("Channels not loaded");
      }
      if (
        channels.some(
          (channel) =>
            channel.status === "opening" &&
            channel.isOutbound &&
            !channel.confirmations
        )
      ) {
        throw new Error(
          "You already are opening a channel which has not been confirmed yet. Please wait for one block confirmation."
        );
      }

      useChannelOrderStore.getState().setOrder(order as NewChannelOrder);
      setShowConfirmModal(false);
      navigate("/channels/order");
    } catch (error) {
      toast.error("Something went wrong", {
        description: `${error}`,
      });
      setShowConfirmModal(false);
    }
  }

  if (!channelPeerSuggestions || !balances) {
    return <Loading />;
  }

  const openImmediately =
    order.amountSat &&
    order.paymentMethod === "onchain" &&
    +order.amountSat < balances.onchain.spendableSat;

  return (
    <>
      <AppHeader
        pageTitle="Open Channel with On-Chain"
        title="Open Channel with On-Chain"
        description="Funds used to open a channel minus fees will be added to your spending balance"
        contentRight={
          <div className="flex items-end">
            <Link to="/channels/incoming" className="underline text-sm">
              Open Channel with Lightning
            </Link>
          </div>
        }
      />
      <div className="md:max-w-md max-w-full flex flex-col gap-5 flex-1">
        <img
          src={LightningNetworkDarkSVG}
          className="w-full hidden dark:block"
        />
        <img src={LightningNetworkLightSVG} className="w-full dark:hidden" />
        <p className="text-muted-foreground">
          Open a channel with on-chain funds. Both parties are free to close the
          channel at any time. However, by keeping more funds on your side of
          the channel and using it regularly, there is more chance the channel
          will stay open.{" "}
          <ExternalLink
            className="underline"
            to="https://guides.getalby.com/user-guide/alby-hub/node/advanced-increase-spending-balance-with-on-chain-bitcoin"
          >
            Learn more
          </ExternalLink>
          .
        </p>
        <form
          onSubmit={onSubmit}
          className="md:max-w-md max-w-full flex flex-col gap-5 flex-1"
        >
          <div className="grid gap-1.5">
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger type="button">
                  <div className="flex flex-row gap-2 items-center justify-start text-sm">
                    <Label htmlFor="amount">
                      Increase spending balance (sats)
                    </Label>
                    <InfoIcon className="h-4 w-4 shrink-0 text-muted-foreground" />
                  </div>
                </TooltipTrigger>
                <TooltipContent>
                  Configure the amount of spending capacity you need. You will
                  need to deposit on-chain bitcoin to cover the entire channel
                  size, plus on-chain fees.
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>

            {order.amountSat && +order.amountSat < 200_000 && (
              <p className="text-muted-foreground text-xs">
                For a smooth experience consider a opening a channel of{" "}
                <FormattedBitcoinAmount amountMsat={200_000 * 1000} /> in size
                or more.{" "}
                <ExternalLink
                  to="https://guides.getalby.com/user-guide/alby-hub/node"
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
              min={selectedPeer?.minimumChannelSizeSat || 100000}
              value={order.amountSat}
              onChange={(e) => {
                setAmountSat(e.target.value.trim());
              }}
            />
            <div className="text-muted-foreground text-sm sensitive slashed-zero">
              Current on-chain balance:{" "}
              <FormattedBitcoinAmount
                amountMsat={balances.onchain.spendableSat * 1000}
              />
            </div>
            <div className="grid grid-cols-3 gap-1.5 text-muted-foreground text-xs">
              {presetAmounts.map((presetAmountSat) => (
                <div
                  key={presetAmountSat}
                  className={cn(
                    "text-center border rounded p-2 cursor-pointer hover:border-muted-foreground",
                    +(order.amountSat || "0") === presetAmountSat &&
                      "border-primary hover:border-primary"
                  )}
                  onClick={() => setAmountSat(presetAmountSat.toString())}
                >
                  {formatAmount(presetAmountSat * 1000, 0)}
                </div>
              ))}
            </div>
          </div>
          <>
            <div className="flex flex-col gap-3">
              {selectedPeer &&
                order.paymentMethod === "onchain" &&
                selectedPeer.pubkey === order.pubkey && (
                  <div className="grid gap-1.5">
                    <Label>Choose your channel peer:</Label>
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
                                      className="size-8 object-contain"
                                    />
                                  )}
                                  <div>
                                    {peer.name}
                                    {peer.minimumChannelSizeSat > 0 && (
                                      <span className="ml-4 text-xs text-muted-foreground slashed-zero">
                                        Min.{" "}
                                        <FormattedBitcoinAmount
                                          amountMsat={
                                            peer.minimumChannelSizeSat * 1000
                                          }
                                        />
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

            <div className="mt-2 flex items-top space-x-2">
              <Checkbox
                id="public-channel"
                checked={order.isPublic}
                onCheckedChange={() => setPublic(!order.isPublic)}
                className="mr-2"
                disabled={selectedPeer && !selectedPeer.publicChannelsAllowed}
                title={
                  selectedPeer && !selectedPeer.publicChannelsAllowed
                    ? "This channel partner does not support public channels."
                    : undefined
                }
              />
              <div className="grid gap-1.5 leading-none">
                <Label htmlFor="public-channel" className="cursor-pointer">
                  Public Channel
                </Label>
                <p className="text-xs text-muted-foreground">
                  Not recommended for most users.{" "}
                  <ExternalLink
                    className="underline"
                    to="https://guides.getalby.com/user-guide/alby-hub/faq/should-i-open-a-private-or-public-channel"
                  >
                    Learn more
                  </ExternalLink>
                </p>
              </div>
            </div>
          </>
          <MempoolAlert />
          <SwapAlert swapType="in" />
          {channels?.some((channel) => channel.public !== !!order.isPublic) && (
            <ChannelPublicPrivateAlert />
          )}
          {selectedPeer?.note && <ChannelPeerNote peer={selectedPeer} />}
          <DuplicateChannelAlert
            pubkey={order?.pubkey}
            name={selectedPeer?.name}
          />
          <Button size="lg">{openImmediately ? "Open Channel" : "Next"}</Button>
        </form>

        <div className="flex-1 flex flex-col justify-end items-center gap-4">
          <p className="mt-32 text-sm text-muted-foreground text-center">
            Other options
          </p>
          <LinkButton
            to="/channels/incoming"
            className="w-full"
            variant="secondary"
          >
            Increase Receiving Capacity
          </LinkButton>
          <ExternalLinkButton
            to="https://www.getalby.com/topup"
            className="w-full"
            variant="secondary"
          >
            Buy Bitcoin
          </ExternalLinkButton>
        </div>
      </div>

      {/* Confirmation Modal */}
      <Dialog open={showConfirmModal} onOpenChange={setShowConfirmModal}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Confirm Channel Opening</DialogTitle>
            <DialogDescription>
              Are you sure you want to open a Lightning channel with the
              following details?
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <div className="font-medium text-muted-foreground">Peer</div>
                <div>{selectedPeer?.name || "Custom"}</div>
              </div>
              <div>
                <div className="font-medium text-muted-foreground">Amount</div>
                <div>
                  <FormattedBitcoinAmount
                    amountMsat={parseInt(order.amountSat || "0") * 1000}
                  />
                </div>
              </div>
              <div>
                <div className="font-medium text-muted-foreground">
                  Channel Type
                </div>
                <div>{order.isPublic ? "Public" : "Private"}</div>
              </div>
              <div>
                <div className="font-medium text-muted-foreground">
                  Payment Method
                </div>
                <div>On-chain</div>
              </div>
            </div>

            {selectedPeer?.name === "Custom" && order.pubkey && (
              <div className="text-sm">
                <div className="font-medium text-muted-foreground">
                  Node Public Key
                </div>
                <div className="font-mono text-xs break-all bg-muted p-2 rounded">
                  {order.pubkey}
                </div>
              </div>
            )}

            <Alert variant="warning">
              <InfoIcon />
              <AlertDescription>
                <strong>Important:</strong> Opening a channel requires an
                on-chain transaction and network fees. This action cannot be
                undone. Please verify all details before proceeding.
              </AlertDescription>
            </Alert>
          </div>

          <DialogFooter className="gap-2">
            <Button
              variant="outline"
              onClick={() => setShowConfirmModal(false)}
            >
              Cancel
            </Button>
            <Button onClick={handleConfirmSubmit}>
              Confirm & Open Channel
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}

type NewChannelOnchainProps = {
  order: Partial<OnchainOrder>;
  setOrder: React.Dispatch<React.SetStateAction<Partial<OnchainOrder>>>;
  showCustomOptions: boolean;
};

function NewChannelOnchain(props: NewChannelOnchainProps) {
  const { data: peers } = usePeers();

  if (props.order.paymentMethod !== "onchain") {
    throw new Error("unexpected payment method");
  }
  const { pubkey, host } = props.order;
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

  const { data: nodeDetails } = useNodeDetails(pubkey);

  React.useEffect(() => {
    const socketAddress = nodeDetails?.sockets?.split(",")?.[0];
    if (socketAddress) {
      setHost(socketAddress);
    }
  }, [nodeDetails, setHost]);

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
                required
                placeholder="Pubkey of the peer"
                onChange={(e) => {
                  const parts = e.target.value.trim().split("@");
                  setPubkey(parts[0]);
                  if (parts.length > 1) {
                    setHost(parts[1]);
                  }
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
                  required
                  placeholder="0.0.0.0:9735 or [2600::]:9735"
                  onChange={(e) => {
                    setHost(e.target.value.trim());
                  }}
                />
              </div>
            )}
          </>
        )}
      </div>
    </>
  );
}
