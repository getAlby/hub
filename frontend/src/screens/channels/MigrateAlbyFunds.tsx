import React from "react";
import { useNavigate } from "react-router-dom";
import Loading from "src/components/Loading";
import toast from "src/components/Toast";
import { ALBY_FEE_RESERVE, MIN_ALBY_BALANCE } from "src/constants";
import { useAlbyBalance } from "src/hooks/useAlbyBalance";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useCSRF } from "src/hooks/useCSRF";
import { useChannels } from "src/hooks/useChannels";
import { useInfo } from "src/hooks/useInfo";
import {
  LSPOption,
  NewWrappedInvoiceRequest,
  NewWrappedInvoiceResponse,
} from "src/types";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";

const DEFAULT_LSP: LSPOption = "OLYMPUS";

export default function MigrateAlbyFunds() {
  const { data: albyMe } = useAlbyMe();
  const { data: albyBalance } = useAlbyBalance();
  const { data: csrf } = useCSRF();
  const { data: channels } = useChannels();
  const { mutate: refetchInfo } = useInfo();
  const [prePurchaseChannelCount, setPrePurchaseChannelCount] = React.useState<
    number | undefined
  >();
  const [error, setError] = React.useState("");
  const [hasRequestedInvoice, setRequestedInvoice] = React.useState(false);
  const navigate = useNavigate();
  const [amount, setAmount] = React.useState(0);

  const [wrappedInvoiceResponse, setWrappedInvoiceResponse] = React.useState<
    NewWrappedInvoiceResponse | undefined
  >();

  const requestWrappedInvoice = React.useCallback(
    async (amount: number) => {
      try {
        if (!channels) {
          throw new Error("Channels not loaded");
        }
        setPrePurchaseChannelCount(channels.length);
        if (!csrf) {
          throw new Error("csrf not loaded");
        }
        const newJITChannelRequest: NewWrappedInvoiceRequest = {
          lsp: DEFAULT_LSP,
          amount,
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
        setError("Failed to connect to request wrapped invoice: " + error);
      }
    },
    [channels, csrf]
  );

  const payWrappedInvoice = React.useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault();
      try {
        if (!wrappedInvoiceResponse) {
          throw new Error("No wrapped invoice");
        }
        if (!csrf) {
          throw new Error("No csrf token");
        }
        await request("/api/alby/pay", {
          method: "POST",
          headers: {
            "X-CSRF-Token": csrf,
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            invoice: wrappedInvoiceResponse.wrappedInvoice,
          }),
        });
      } catch (error) {
        handleRequestError("Failed to pay channel funding invoice", error);
      }
    },
    [csrf, wrappedInvoiceResponse]
  );

  React.useEffect(() => {
    if (hasRequestedInvoice || !channels || !albyMe || !albyBalance) {
      return;
    }
    setRequestedInvoice(true);
    const _amount = albyBalance.sats - ALBY_FEE_RESERVE;
    setAmount(_amount);
    requestWrappedInvoice(_amount);
  }, [
    hasRequestedInvoice,
    albyBalance,
    channels,
    albyMe,
    requestWrappedInvoice,
  ]);

  const hasOpenedChannel =
    channels &&
    prePurchaseChannelCount !== undefined &&
    channels.length > prePurchaseChannelCount;

  React.useEffect(() => {
    if (hasOpenedChannel) {
      (async () => {
        toast.success("Channel opened!");
        await refetchInfo();
        navigate("/");
      })();
    }
  }, [hasOpenedChannel, navigate, refetchInfo]);

  if (!albyMe || !albyBalance || !channels || !wrappedInvoiceResponse) {
    return <Loading />;
  }

  if (error) {
    return <p>{error}</p>;
  }

  if (albyBalance.sats < MIN_ALBY_BALANCE) {
    return (
      <p>You don't have enough sats in your Alby account to open a channel.</p>
    );
  }

  /*if (channels.length) {
    return (
      <p>You already have a channel.</p>
    );
  }*/

  const LSP_FREE_INCOMING = 100000;
  const estimatedChannelSize =
    amount - wrappedInvoiceResponse.fee + LSP_FREE_INCOMING;
  return (
    <div className="flex flex-col justify-center items-center gap-4">
      <h1 className="mt-8">Migrate Alby Account Funds</h1>
      <p className="font-bold">Alby Account Balance: {albyBalance.sats} sats</p>
      <p className="font-bold">
        LSP fee ({DEFAULT_LSP}): {wrappedInvoiceResponse.fee} sats
      </p>
      <p className="font-bold">Alby fee reserve: {ALBY_FEE_RESERVE} sats</p>
      <p className="font-bold">
        Estimated Channel size: {estimatedChannelSize} sats
      </p>
      <p className="font-bold">
        Estimated sendable: {amount - wrappedInvoiceResponse.fee} sats
      </p>
      <p className="font-bold">
        Estimated receivable: {LSP_FREE_INCOMING - wrappedInvoiceResponse.fee}{" "}
        sats
      </p>
      <form className="mt-16">
        <button
          className="bg-blue-300 hover:bg-blue-200 px-8 py-4 font-bold text-lg rounded-lg"
          onClick={payWrappedInvoice}
        >
          Open Channel
        </button>
      </form>
    </div>
  );
}
