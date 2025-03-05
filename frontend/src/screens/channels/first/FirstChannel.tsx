import { ChevronDown } from "lucide-react";
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
import { useAlbyBalance } from "src/hooks/useAlbyBalance";
import { useChannels } from "src/hooks/useChannels";

import { useInfo } from "src/hooks/useInfo";
import { AutoChannelRequest, AutoChannelResponse } from "src/types";
import { request } from "src/utils/request";

import { MempoolAlert } from "src/components/MempoolAlert";
import { PayLightningInvoice } from "src/components/PayLightningInvoice";
import { ALBY_MIN_HOSTED_BALANCE_FOR_FIRST_CHANNEL } from "src/constants";

export function FirstChannel() {
  const { data: info } = useInfo();
  const { data: channels } = useChannels(true);
  const [isLoading, setLoading] = React.useState(false);
  const [showAdvanced, setShowAdvanced] = React.useState(false);
  const [isPublic, setPublic] = React.useState(false);

  const navigate = useNavigate();
  const { toast } = useToast();
  const [invoice, setInvoice] = React.useState<string>();
  const [channelSize, setChannelSize] = React.useState<number>();
  const { data: albyBalance } = useAlbyBalance();

  React.useEffect(() => {
    if (channels?.length) {
      navigate("/channels/first/opening");
    }
  }, [channels, navigate]);

  React.useEffect(() => {
    if (info && !info.albyAccountConnected) {
      navigate("/channels/incoming");
    }
  }, [info, navigate]);

  if (!info?.albyAccountConnected || !channels) {
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

  const canPayForFirstChannel =
    albyBalance &&
    albyBalance.sats >= ALBY_MIN_HOSTED_BALANCE_FOR_FIRST_CHANNEL;

  return (
    <>
      <AppHeader
        title="Open Your First Channel"
        description="Open a channel to another lightning network node to join the lightning network"
      />
      <MempoolAlert />
      {invoice && channelSize && (
        <div className="flex flex-col gap-4 items-center justify-center max-w-md">
          <p className="text-muted-foreground slashed-zero">
            Please pay the lightning invoice below which will cover the costs of
            opening your first channel. You will receive a channel with{" "}
            {new Intl.NumberFormat().format(channelSize)} sats of incoming
            liquidity.
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
              src="/images/illustrations/lightning-network-dark.svg"
              className="w-full hidden dark:block"
            />
            <img
              src="/images/illustrations/lightning-network-light.svg"
              className="w-full dark:hidden"
            />
            {canPayForFirstChannel ? (
              <>
                <p>
                  You currently have{" "}
                  <span className="font-medium text-foreground sensitive slashed-zero">
                    {new Intl.NumberFormat().format(albyBalance?.sats)} Alby fee
                    credits.
                  </span>{" "}
                  <Link
                    to="https://guides.getalby.com/user-guide/alby-account-and-browser-extension/alby-account/faqs-alby-account/what-are-fee-credits-in-my-alby-account"
                    target="_blank"
                    className="underline"
                  >
                    Learn more
                  </Link>
                </p>
                <p>
                  These fee credits will be applied to open your first Lightning
                  channel.
                </p>
              </>
            ) : (
              <>
                <p>
                  You're now going to open your first lightning channel and can
                  begin using your Hub in the booming bitcoin economy!
                </p>
                <p>
                  After paying a lightning invoice to cover on-chain fees,
                  you'll immediately be able to receive and send bitcoin with
                  your Hub.
                </p>
              </>
            )}
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
                      Only enable if you want to receive keysend payments. (e.g.
                      podcasting)
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
                  <ChevronDown className="w-4 h-4 ml-1" />
                </Button>
              </div>
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
