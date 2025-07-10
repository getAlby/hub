import { ChevronDownIcon } from "lucide-react";
import React from "react";
import { Link, useNavigate } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import Loading from "src/components/Loading";
import { Button } from "src/components/ui/button";
import { Checkbox } from "src/components/ui/checkbox";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { Separator } from "src/components/ui/separator";
import { useToast } from "src/components/ui/use-toast";
import { useChannels } from "src/hooks/useChannels";

import { useInfo } from "src/hooks/useInfo";
import { AutoChannelRequest, AutoChannelResponse } from "src/types";
import { request } from "src/utils/request";

import { MempoolAlert } from "src/components/MempoolAlert";
import { PayLightningInvoice } from "src/components/PayLightningInvoice";
import { ChannelPublicPrivateAlert } from "src/components/channels/ChannelPublicPrivateAlert";

import LightningNetworkDarkSVG from "public/images/illustrations/lightning-network-dark.svg";
import LightningNetworkLightSVG from "public/images/illustrations/lightning-network-light.svg";

export function AutoChannel() {
  const { data: info } = useInfo();
  const { data: channels } = useChannels(true);
  const [isLoading, setLoading] = React.useState(false);
  const [showAdvanced, setShowAdvanced] = React.useState(false);
  const [isPublic, setPublic] = React.useState(false);

  const navigate = useNavigate();
  const { toast } = useToast();
  const [invoice, setInvoice] = React.useState<string>();
  const [channelSize, setChannelSize] = React.useState<number>();
  const [, setPrevChannelIds] = React.useState<string[]>();

  React.useEffect(() => {
    if (channels) {
      setPrevChannelIds((current) => {
        if (current) {
          const newChannelId = channels.find(
            (channel) => !current?.includes(channel.id) && channel.fundingTxId
          )?.id;

          if (newChannelId) {
            console.info("Found new channel", newChannelId);
            navigate("/channels/auto/opening", {
              state: {
                newChannelId,
              },
            });
          }

          return current;
        }

        return channels.map((channel) => channel.id);
      });
    }
  }, [channels, navigate]);

  if (!info || !channels) {
    return <Loading />;
  }

  async function openChannel() {
    if (!info || !channels) {
      return;
    }
    setLoading(true);
    try {
      const newInstantChannelInvoiceRequest: AutoChannelRequest = {
        isPublic,
      };
      const autoChannelResponse = await request<AutoChannelResponse>(
        "/api/alby/auto-channel",
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(newInstantChannelInvoiceRequest),
        }
      );
      if (!autoChannelResponse) {
        throw new Error("unexpected auto channel response");
      }

      setInvoice(autoChannelResponse.invoice);
      setChannelSize(autoChannelResponse.channelSize);
    } catch (error) {
      setLoading(false);
      console.error(error);
      toast({
        title: "Something went wrong. Please try again",
        variant: "destructive",
      });
    }
  }

  return (
    <>
      <AppHeader
        title="Open a lightning channel"
        description="Open a channel to another node on the lightning network"
      />
      <MempoolAlert />
      {invoice && channelSize && (
        <div className="flex flex-col gap-4 items-center justify-center max-w-md">
          <p className="text-muted-foreground slashed-zero">
            Please pay the lightning invoice below which will cover the costs of
            opening your channel. You will receive a channel with{" "}
            {new Intl.NumberFormat().format(channelSize)} sats of receiving
            capacity.
          </p>
          <PayLightningInvoice invoice={invoice} />

          <Separator className="mt-8" />
          <p className="mt-8 text-sm mb-2 text-muted-foreground">
            Other options
          </p>
          <Link to="/channels/outgoing" className="w-full">
            <Button className="w-full" variant="secondary">
              Open Channel with On-Chain Bitcoin
            </Button>
          </Link>
          <ExternalLink to="https://www.getalby.com/topup" className="w-full">
            <Button className="w-full" variant="secondary">
              Buy Bitcoin
            </Button>
          </ExternalLink>
        </div>
      )}
      {!invoice && (
        <>
          <div className="flex flex-col gap-6 max-w-md text-muted-foreground">
            <img
              src={LightningNetworkDarkSVG}
              className="w-full hidden dark:block"
            />
            <img
              src={LightningNetworkLightSVG}
              className="w-full dark:hidden"
            />

            <>
              <p>
                You're now going to open a new lightning channel that you can
                use to send and receive payments using your Hub in the booming
                bitcoin economy! To make things easy, Alby has picked a channel
                partner for you from one of our recommended channel partners.
              </p>
              <p>
                After paying a lightning invoice to cover on-chain fees, you'll
                immediately be able to receive and send bitcoin through this
                channel with your Hub.
              </p>
              <p className="text-muted-foreground">
                Alby Hub works with selected service providers (LSPs) which
                provide the best network connectivity and liquidity to receive
                payments.{" "}
                <ExternalLink
                  className="underline"
                  to="https://guides.getalby.com/user-guide/alby-hub/faq/how-to-open-a-payment-channel"
                >
                  Learn more
                </ExternalLink>
              </p>
            </>
            {showAdvanced && (
              <>
                <div className="mt-2 flex items-top space-x-2">
                  <Checkbox
                    id="public-channel"
                    onCheckedChange={() => setPublic(!isPublic)}
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
            )}
            {!showAdvanced && (
              <div>
                <Button
                  type="button"
                  variant="link"
                  className="text-muted-foreground text-xs px-0"
                  onClick={() => setShowAdvanced((current) => !current)}
                >
                  Advanced Options
                  <ChevronDownIcon className="size-4 ml-1" />
                </Button>
              </div>
            )}
            {channels?.some((channel) => channel.public !== !!isPublic) && (
              <ChannelPublicPrivateAlert />
            )}
            <LoadingButton loading={isLoading} onClick={openChannel}>
              Open Channel
            </LoadingButton>
          </div>
        </>
      )}
    </>
  );
}
