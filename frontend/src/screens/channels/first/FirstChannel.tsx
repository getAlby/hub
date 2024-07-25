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
import { useChannels } from "src/hooks/useChannels";
import { useCSRF } from "src/hooks/useCSRF";
import { useInfo } from "src/hooks/useInfo";
import { AutoChannelRequest, AutoChannelResponse } from "src/types";
import { request } from "src/utils/request";

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

  return (
    <>
      <AppHeader
        title="Open Your First Channel"
        description="Open a channel to another lightning network node to join the lightning network"
      />

      {invoice && channelSize && (
        <div className="flex flex-col gap-4 items-center justify-center">
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
              Open Channel with Onchain Bitcoin
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

          <div className="flex flex-col gap-2 max-w-sm">
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
            <LoadingButton loading={isLoading} onClick={openChannel}>
              Open Channel
            </LoadingButton>
          </div>
        </>
      )}
    </>
  );
}
