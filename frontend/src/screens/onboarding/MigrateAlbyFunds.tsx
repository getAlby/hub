import React from "react";
import { useNavigate } from "react-router-dom";
import Loading from "src/components/Loading";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { LoadingButton } from "src/components/ui/loading-button";
import { Table, TableBody, TableCell, TableRow } from "src/components/ui/table";
import { useToast } from "src/components/ui/use-toast";
import {
  ALBY_FEE_RESERVE,
  ALBY_SERVICE_FEE,
  MIN_ALBY_BALANCE,
} from "src/constants";
import { useAlbyBalance } from "src/hooks/useAlbyBalance";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useCSRF } from "src/hooks/useCSRF";
import { useChannels } from "src/hooks/useChannels";
import { useInfo } from "src/hooks/useInfo";
import {
  LSPOption,
  NewInstantChannelInvoiceRequest,
  NewInstantChannelInvoiceResponse,
} from "src/types";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";

const DEFAULT_LSP: LSPOption = "ALBY";

export default function MigrateAlbyFunds() {
  const { data: albyMe } = useAlbyMe();
  const { data: albyBalance } = useAlbyBalance();
  const { data: csrf } = useCSRF();
  const { data: channels } = useChannels();
  const { mutate: refetchInfo } = useInfo();
  const { toast } = useToast();
  const [prePurchaseChannelCount, setPrePurchaseChannelCount] = React.useState<
    number | undefined
  >();
  const [error, setError] = React.useState("");
  const [hasRequestedInvoice, setRequestedInvoice] = React.useState(false);
  const [isOpeningChannel, setOpeningChannel] = React.useState(false);
  const navigate = useNavigate();
  const [amount, setAmount] = React.useState(0);

  const [wrappedInvoiceResponse, setWrappedInvoiceResponse] = React.useState<
    NewInstantChannelInvoiceResponse | undefined
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
        const newJITChannelRequest: NewInstantChannelInvoiceRequest = {
          lsp: DEFAULT_LSP,
          amount,
        };
        const response = await request<NewInstantChannelInvoiceResponse>(
          "/api/instant-channel-invoices",
          {
            method: "POST",
            headers: {
              "X-CSRF-Token": csrf,
              "Content-Type": "application/json",
            },
            body: JSON.stringify(newJITChannelRequest),
          }
        );
        if (!response?.invoice) {
          throw new Error("No invoice in response");
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
          throw new Error("No invoice");
        }
        if (!csrf) {
          throw new Error("No csrf token");
        }
        setOpeningChannel(true);
        await request("/api/alby/pay", {
          method: "POST",
          headers: {
            "X-CSRF-Token": csrf,
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            invoice: wrappedInvoiceResponse.invoice,
          }),
        });
      } catch (error) {
        handleRequestError(
          toast,
          "Failed to pay channel funding invoice",
          error
        );
        setOpeningChannel(false);
      }
    },
    [csrf, wrappedInvoiceResponse]
  );

  React.useEffect(() => {
    if (hasRequestedInvoice || !channels || !albyMe || !albyBalance) {
      return;
    }
    setRequestedInvoice(true);
    const _amount = Math.floor(
      albyBalance.sats * (1 - ALBY_FEE_RESERVE) * (1 - ALBY_SERVICE_FEE)
    );
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
        toast({ title: "Channel opened!" });
        await refetchInfo();
        navigate("/");
      })();
    }
  }, [hasOpenedChannel, navigate, refetchInfo, toast]);

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
    <div className="flex flex-col justify-center items-center gap-5 p-5 max-w-md">
      <div className="grid gap-2 text-center">
        <h1 className="text-2xl font-semibold">Open a Channel</h1>
        <p className="text-muted-foreground">
          You can use your remaining balance on Alby hosted lightning wallet to
          fund your first lightning channel.
        </p>
      </div>
      <Card className="w-full">
        <CardHeader>
          <CardTitle>Summary</CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableBody>
              <TableRow>
                <TableCell className="font-medium">
                  Current Account balance
                </TableCell>
                <TableCell className="text-right">
                  {albyBalance.sats} sats
                </TableCell>
              </TableRow>
              <TableRow>
                <TableCell className="font-medium">Fees</TableCell>
                <TableCell className="text-right">
                  {Math.floor(amount * ALBY_SERVICE_FEE) +
                    wrappedInvoiceResponse.fee}{" "}
                  sats
                </TableCell>
              </TableRow>
              <TableRow className="border-0">
                <TableCell className="font-medium">Alby Hub balance</TableCell>
                <TableCell className="font-medium text-right">
                  {amount - wrappedInvoiceResponse.fee} sats
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {/* <h1 className="mt-8">Migrate Alby Account Funds</h1> */}
      {/* <p className="font-bold">Invoice to pay: {amount} sats</p>
      <p className="font-bold">Alby Account Balance: {albyBalance.sats} sats</p>
      <p className="font-bold">
        LSP fee ({DEFAULT_LSP}): {wrappedInvoiceResponse.fee} sats
      </p>
      <p className="font-bold">
        Alby service fee: {Math.floor(amount * ALBY_SERVICE_FEE)} sats
      </p>
      <p className="font-bold">
        Alby fee reserve: {Math.floor(albyBalance.sats * ALBY_FEE_RESERVE)} sats
      </p>
      <p className="font-bold">
        Estimated Channel size: {estimatedChannelSize} sats
      </p>
      <p className="font-bold">
        Estimated spendable: {amount - wrappedInvoiceResponse.fee} sats
      </p>
      <p className="font-bold">
        Estimated receivable: {LSP_FREE_INCOMING - wrappedInvoiceResponse.fee}{" "}
        sats
      </p> */}
      <form>
        <LoadingButton
          onClick={payWrappedInvoice}
          disabled={isOpeningChannel}
          loading={isOpeningChannel}
        >
          Migrate Funds and Open Channel
        </LoadingButton>
      </form>
    </div>
  );
}
