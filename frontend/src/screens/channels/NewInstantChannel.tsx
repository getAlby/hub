import { Payment, init } from "@getalby/bitcoin-connect-react";
import React from "react";
import { LoadingButton } from "src/components/ui/loading-button";
import { MIN_0CONF_BALANCE } from "src/constants";
import { useCSRF } from "src/hooks/useCSRF";
import { useChannels } from "src/hooks/useChannels";
import {
  LSPOption,
  LSP_OPTIONS,
  NewWrappedInvoiceRequest,
  NewWrappedInvoiceResponse,
} from "src/types";
import { request } from "src/utils/request";
init({
  showBalance: false,
});

export default function NewInstantChannel() {
  const { data: csrf } = useCSRF();
  const { data: channels } = useChannels();
  const [lsp, setLsp] = React.useState<LSPOption | undefined>("OLYMPUS");
  const [amount, setAmount] = React.useState("");
  const [prePurchaseChannelAmount, setPrePurchaseChannelAmount] =
    React.useState<number | undefined>();
  const [isRequestingInvoice, setRequestingInvoice] = React.useState(false);
  const [wrappedInvoiceResponse, setWrappedInvoiceResponse] = React.useState<
    NewWrappedInvoiceResponse | undefined
  >();
  const amountSats = React.useMemo(() => {
    try {
      const _amountSats = parseInt(amount);
      if (_amountSats >= MIN_0CONF_BALANCE) {
        return _amountSats;
      }
    } catch (error) {
      console.error(error);
    }
    return 0;
  }, [amount]);

  // This is not a good check if user already has enough inbound liquidity
  // - check balance instead or how else to check the invoice is paid?
  const hasOpenedChannel =
    channels &&
    prePurchaseChannelAmount !== undefined &&
    channels.length > prePurchaseChannelAmount;

  const requestWrappedInvoice = React.useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault();
      try {
        if (!channels) {
          throw new Error("Channels not loaded");
        }
        setPrePurchaseChannelAmount(channels.length);
        if (!lsp) {
          throw new Error("no lsp selected");
        }
        if (!csrf) {
          throw new Error("csrf not loaded");
        }
        setRequestingInvoice(true);
        const newJITChannelRequest: NewWrappedInvoiceRequest = {
          lsp,
          amount: amountSats,
        };
        const response = await request<NewWrappedInvoiceResponse>(
          "/api/wrapped-invoices",
          {
            method: "POST",
            headers: {
              "X-CSRF-Token": csrf,
              "Content-Type": "application/json",
            },
            body: JSON.stringify(newJITChannelRequest),
          }
        );
        if (!response?.wrappedInvoice) {
          throw new Error("No wrapped invoice in response");
        }
        setWrappedInvoiceResponse(response);
      } catch (error) {
        alert("Failed to connect to request wrapped invoice: " + error);
      } finally {
        setRequestingInvoice(false);
      }
    },
    [amountSats, channels, csrf, lsp]
  );

  return (
    <div className="flex flex-col justify-center items-center gap-4">
      <h1>1. Choose an LSP</h1>
      <div className="flex gap-2">
        {LSP_OPTIONS.map((option) => (
          <button
            key={option}
            className={`shadow-lg p-4 bg-gray-100 rounded-lg hover:bg-yellow-100 active:bg-yellow-300 ${
              option === lsp && "bg-yellow-300"
            } `}
            onClick={() => setLsp(option)}
          >
            {option}
          </button>
        ))}
      </div>
      {lsp && (
        <>
          <h1 className="mt-8">2. Purchase Liquidity</h1>
          <p className="italic text-xs max-w-sm">
            Enter at least {MIN_0CONF_BALANCE} sats. You'll receive outgoing
            liquidity of this amount minus any LSP fees. You'll also get some
            incoming liquidity.
          </p>
          <form onSubmit={requestWrappedInvoice}>
            <p className="text-gray-500 text-sm">Amount in sats</p>
            <div className="flex gap-2 w-full justify-center items-center relative">
              <input
                className="font-mono shadow-md"
                type="number"
                inputMode="numeric"
                pattern="[0-9]*"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
              ></input>{" "}
            </div>
            <LoadingButton
              type="submit"
              disabled={amountSats === 0}
              loading={isRequestingInvoice}
            >
              Create channel
            </LoadingButton>
          </form>
        </>
      )}
      {wrappedInvoiceResponse && (
        <>
          <h1 className="mt-8">3. Complete Payment</h1>
          <p className="font-bold">
            Fee included: {wrappedInvoiceResponse.fee} sats
          </p>
          <Payment
            invoice={wrappedInvoiceResponse.wrappedInvoice}
            payment={
              hasOpenedChannel ? { preimage: "dummy preimage" } : undefined
            }
          />
        </>
      )}
      {hasOpenedChannel && (
        <p className="mt-8 text-green-400">Channel Opened!</p>
      )}
    </div>
  );
}
