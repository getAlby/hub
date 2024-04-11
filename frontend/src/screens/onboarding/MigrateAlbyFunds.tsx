import { AlertTriangle } from "lucide-react";
import React from "react";
import { Link, useNavigate } from "react-router-dom";
import Loading from "src/components/Loading";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { Button } from "src/components/ui/button";
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
  const { data: channels } = useChannels(true);
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

  const [instantChannelResponse, setInstantChannelResponse] = React.useState<
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
        setInstantChannelResponse(response);
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
        if (!instantChannelResponse) {
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
            invoice: instantChannelResponse.invoice,
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
    [csrf, toast, instantChannelResponse]
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

  if (!albyMe || !albyBalance || !channels || !instantChannelResponse) {
    return <Loading />;
  }

  if (error) {
    return <p>{error}</p>;
  }

  /*  TODO: Remove? At least display a link to where to go from here.
  if (channels.length) {
    return (
      <p>You already have a channel.</p>
    );
  }*/

  return (
    <div className="flex flex-col justify-center gap-5 p-5 max-w-md items-stretch">
      <TwoColumnLayoutHeader
        title="Open a Channel"
        description="You can use your remaining balance on Alby hosted lightning wallet to
      fund your first lightning channel."
      />
      {albyBalance.sats >= MIN_ALBY_BALANCE ? (
        <>
          <div className="border rounded-lg">
            <Table>
              <TableBody>
                <TableRow className="border-b-0">
                  <TableCell className="font-medium p-3">
                    Current Account balance
                  </TableCell>
                  <TableCell className="text-right p-3">
                    {new Intl.NumberFormat().format(albyBalance.sats)} sats
                  </TableCell>
                </TableRow>
                <TableRow>
                  <TableCell className="font-medium p-3 flex flex-row gap-1.5 items-center">
                    Fee
                  </TableCell>
                  <TableCell className="text-right p-3">
                    {new Intl.NumberFormat().format(
                      albyBalance.sats - amount + instantChannelResponse.fee
                    )}{" "}
                    sats
                  </TableCell>
                </TableRow>
                <TableRow>
                  <TableCell className="font-medium p-3">
                    Alby Hub Balance
                  </TableCell>
                  <TableCell className="font-semibold text-right p-3">
                    {new Intl.NumberFormat().format(
                      amount - instantChannelResponse.fee
                    )}{" "}
                    sats
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </div>
          <form className="flex flex-col justify-between text-center gap-2">
            <LoadingButton
              onClick={payWrappedInvoice}
              disabled={isOpeningChannel}
              loading={isOpeningChannel}
            >
              Migrate Funds and Open Channel
            </LoadingButton>
            <Link to="../channels/new/instant">
              <Button variant="link">Open a Channel manually</Button>
            </Link>
          </form>
        </>
      ) : (
        <>
          <Alert>
            <AlertTriangle className="h-4 w-4" />
            <AlertTitle>Not enough funds available!</AlertTitle>
            <AlertDescription>
              You don't have enough funds in your Alby account to fund a new
              channel right now. You can open a channel manually and pay with an
              external wallet though.
            </AlertDescription>
          </Alert>
          <Link to="../channels/new/instant" className="w-full">
            <Button className="w-full">Open a Channel manually</Button>
          </Link>
        </>
      )}
    </div>
  );
}
