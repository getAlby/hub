import { Payment } from "@getalby/bitcoin-connect-react";
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
import { ALBY_MIN_HOSTED_BALANCE_FOR_FIRST_CHANNEL } from "src/constants";
import { useAlbyBalance } from "src/hooks/useAlbyBalance";
import { useChannels } from "src/hooks/useChannels";
import { useCSRF } from "src/hooks/useCSRF";
import { useInfo } from "src/hooks/useInfo";
import { AutoChannelRequest, AutoChannelResponse } from "src/types";
import { request } from "src/utils/request";

import lightningNetworkDark from "/images/illustrations/lightning-network-dark.svg";
import lightningNetworkLight from "/images/illustrations/lightning-network-light.svg";

export function FirstChannel() {
  const { data: info } = useInfo();
  const { data: channels } = useChannels(true);
  const [isLoading, setLoading] = React.useState(false);
  const [showAdvanced, setShowAdvanced] = React.useState(false);
  const [isPublic, setPublic] = React.useState(false);
  const { data: csrf } = useCSRF();
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

  if (!info || !channels) {
    return <Loading />;
  }

  async function openChannel() {
    if (!info || !channels || !csrf) {
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
            "X-CSRF-Token": csrf,
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
      {invoice && channelSize && (
        <div className="flex flex-col gap-4 items-center justify-center max-w-md">
          <p className="text-muted-foreground">
            Please pay the lightning invoice below which will cover the costs of
            opening your first channel. You will receive a channel with{" "}
            {new Intl.NumberFormat().format(channelSize)} sats of incoming
            liquidity.
          </p>
          <Payment invoice={invoice} paymentMethods="external" />

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
          <div className="flex flex-col gap-4 max-w-md text-muted-foreground">
            <img
              src={lightningNetworkDark}
              className="w-full hidden dark:block"
            />
            <img src={lightningNetworkLight} className="w-full dark:hidden" />
            {canPayForFirstChannel ? (
              <>
                <p>
                  You currently have{" "}
                  <span className="font-medium">
                    {new Intl.NumberFormat().format(albyBalance?.sats)} sats
                  </span>{" "}
                  on your Alby hosted balance.
                </p>
                <p>
                  Those funds will be used to open your first lightning channel
                  and then migrated to your Alby Hub's spending balance.
                </p>
              </>
            ) : (
              <>
                <p>
                  You're now going to open your first lightning channel and can
                  begin use your Alby Hub in the booming bitcoin economy!
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
                  className="text-muted-foreground text-xs"
                  onClick={() => setShowAdvanced((current) => !current)}
                >
                  <ChevronDown className="w-4 h-4 mr-2" />
                  Advanced Options
                </Button>
              </div>
            )}
            <LoadingButton loading={isLoading} onClick={openChannel}>
              Open Channel
              {albyBalance && albyBalance?.sats > 0 && <> and Migrate Funds</>}
            </LoadingButton>
          </div>
        </>
      )}
    </>
  );
}
