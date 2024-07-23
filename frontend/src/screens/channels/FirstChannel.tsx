import React from "react";
import { useNavigate } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import Loading from "src/components/Loading";
import { Checkbox } from "src/components/ui/checkbox";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useAlbyBalance } from "src/hooks/useAlbyBalance";
import { useChannels } from "src/hooks/useChannels";
import { useCSRF } from "src/hooks/useCSRF";
import { useInfo } from "src/hooks/useInfo";
import useChannelOrderStore from "src/state/ChannelOrderStore";
import {
  NewChannelOrder,
  NewInstantChannelInvoiceRequest,
  NewInstantChannelInvoiceResponse,
} from "src/types";
import { request } from "src/utils/request";

export function FirstChannel() {
  const { data: albyBalance } = useAlbyBalance();
  const { data: info, hasChannelManagement } = useInfo();
  const { data: channels } = useChannels();
  const [isLoading, setLoading] = React.useState(false);
  const [isPublic, setPublic] = React.useState(false);
  const { data: csrf } = useCSRF();
  const navigate = useNavigate();

  if (!albyBalance || !info || !channels) {
    return <Loading />;
  }

  async function openChannel() {
    if (!albyBalance || !info || !channels || !csrf) {
      return;
    }
    setLoading(true);

    // if migrating funds, add 500K and round up to nearest million
    // TODO: otherwise let API decide channel size
    const amount =
      Math.ceil((albyBalance.sats + 500_000) / 1_000_000) * 1_000_000;

    const lspUrl = `https://api.getalby.com/internal/lsp/alby/${info.network}/v1`;

    const newInstantChannelInvoiceRequest: NewInstantChannelInvoiceRequest = {
      lspType: "LSPS1",
      lspUrl,
      amount,
      public: isPublic,
    };
    const channelOrderResponse =
      await request<NewInstantChannelInvoiceResponse>(
        "/api/instant-channel-invoices",
        {
          method: "POST",
          headers: {
            "X-CSRF-Token": csrf,
            "Content-Type": "application/json",
          },
          body: JSON.stringify(newInstantChannelInvoiceRequest),
        }
      );
    if (!channelOrderResponse) {
      throw new Error("No order in response");
    }

    const order: NewChannelOrder = {
      status: "paid",
      prevChannelIds: channels.map((channel) => channel.id),
      isPublic,
      paymentMethod: "lightning",
      lspType: "LSPS1",
      lspUrl,
      amount: amount.toString(),
    };

    // if there is an invoice in the response we must pay it
    if (channelOrderResponse.invoice) {
      try {
        // try pay with alby shared wallet funds
        // TODO: remove this code once all Alby users have migrated to Alby Hub
        if (hasChannelManagement && info.backendType === "LDK") {
          await request("/api/alby/pay", {
            method: "POST",
            headers: {
              "X-CSRF-Token": csrf,
              "Content-Type": "application/json",
            },
            body: JSON.stringify({
              invoice: channelOrderResponse.invoice,
            }),
          });
          order.status = "paid";
        }
      } catch (error) {
        console.error("Failed to pay with alby shared balance", error);
      }
    } else {
      // alby API already paid the invoice
      order.status = "paid";
    }

    useChannelOrderStore.getState().setOrder(order);
    navigate("/channels/order");
    return true;
  }

  return (
    <>
      <AppHeader
        title="Open Your First Channel"
        description="Open a channel to another lightning network node to join the lightning network"
      />

      <div className="mt-2 flex items-top space-x-2">
        <Checkbox
          id="public-channel"
          onCheckedChange={() => setPublic(!isPublic)}
          className="mr-2"
        />
        <div className="grid gap-1.5 leading-none">
          <Label htmlFor="public-channel" className="flex items-center gap-2">
            Public Channel
          </Label>
          <p className="text-xs text-muted-foreground">
            Only enable if you want to receive keysend payments. (e.g.
            podcasting)
          </p>
        </div>
      </div>

      <div>
        <LoadingButton loading={isLoading} onClick={openChannel}>
          Open Channel
        </LoadingButton>
      </div>
    </>
  );
}
