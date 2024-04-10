import { Payment, init } from "@getalby/bitcoin-connect-react";
import React from "react";
import { useNavigate } from "react-router-dom";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "src/components/ui/select";
import { Table, TableBody, TableCell, TableRow } from "src/components/ui/table";
import { useToast } from "src/components/ui/use-toast";
import { MIN_0CONF_BALANCE } from "src/constants";
import { useCSRF } from "src/hooks/useCSRF";
import { useChannels } from "src/hooks/useChannels";
import { useInfo } from "src/hooks/useInfo";
import {
  LSPOption,
  LSP_OPTIONS,
  NewInstantChannelInvoiceRequest,
  NewInstantChannelInvoiceResponse,
} from "src/types";
import { request } from "src/utils/request";
init({
  showBalance: false,
});

export default function NewInstantChannel() {
  const { data: csrf } = useCSRF();
  const { mutate: refetchInfo } = useInfo();
  const navigate = useNavigate();
  const { toast } = useToast();
  const { data: channels } = useChannels(true);
  const [lsp, setLsp] = React.useState<LSPOption | undefined>("ALBY");
  const [amount, setAmount] = React.useState("");
  const [prePurchaseChannelAmount, setPrePurchaseChannelAmount] =
    React.useState<number | undefined>();
  const [isRequestingInvoice, setRequestingInvoice] = React.useState(false);
  const [wrappedInvoiceResponse, setWrappedInvoiceResponse] = React.useState<
    NewInstantChannelInvoiceResponse | undefined
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

  React.useEffect(() => {
    if (hasOpenedChannel) {
      (async () => {
        toast({ title: "Channel opened!" });
        await refetchInfo();
        navigate("/");
      })();
    }
  }, [hasOpenedChannel, navigate, refetchInfo, toast]);

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
        const newJITChannelRequest: NewInstantChannelInvoiceRequest = {
          lsp,
          amount: amountSats,
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
        alert("Failed to connect to request wrapped invoice: " + error);
      } finally {
        setRequestingInvoice(false);
      }
    },
    [amountSats, channels, csrf, lsp]
  );

  return (
    <div className="flex flex-col justify-center items-center gap-5">
      <TwoColumnLayoutHeader
        title={"Buy an Instant Channel"}
        description={"Choose your LSP to open an instant channel to your node"}
      />
      {!wrappedInvoiceResponse && (
        <form onSubmit={requestWrappedInvoice} className="grid gap-3">
          <div className="grid w-full max-w-sm items-center gap-1.5">
            <Label htmlFor="lsp">LSP</Label>
            <Select
              value={lsp}
              onValueChange={(value) => setLsp(value as LSPOption)}
            >
              <SelectTrigger>
                <SelectValue placeholder="Select a LSP" />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  {LSP_OPTIONS.map((option) => (
                    <SelectItem value={option}>
                      {option.charAt(0).toUpperCase() +
                        option.slice(1).toLowerCase()}
                    </SelectItem>
                  ))}
                </SelectGroup>
              </SelectContent>
            </Select>
          </div>
          <div className="grid w-full max-w-sm items-center gap-1.5">
            <Label htmlFor="amount">Amount</Label>
            <Input
              type="number"
              id="amount"
              placeholder="Amount in sats"
              inputMode="numeric"
              pattern="[0-9]*"
              min={MIN_0CONF_BALANCE}
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
            />
            <div className="text-muted-foreground text-xs">
              Enter at least {MIN_0CONF_BALANCE} sats. You'll receive outgoing
              liquidity of this amount minus any LSP fees. You'll also get some
              incoming liquidity.
            </div>
          </div>
          <LoadingButton
            type="submit"
            disabled={amountSats === 0}
            loading={isRequestingInvoice}
          >
            Request Channel Offer
          </LoadingButton>
        </form>
      )}
      {wrappedInvoiceResponse && (
        <>
          <div className="border rounded-lg w-96">
            <Table>
              <TableBody>
                <TableRow>
                  <TableCell className="font-medium p-3 flex flex-row gap-1.5 items-center">
                    Fee
                  </TableCell>
                  <TableCell className="text-right p-3">
                    {new Intl.NumberFormat().format(wrappedInvoiceResponse.fee)}{" "}
                    sats
                  </TableCell>
                </TableRow>
                <TableRow>
                  <TableCell className="font-medium p-3">
                    Amount to pay
                  </TableCell>
                  <TableCell className="font-semibold text-right p-3">
                    {new Intl.NumberFormat().format(amountSats)} sats
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </div>
          <Payment
            invoice={wrappedInvoiceResponse.invoice}
            payment={
              hasOpenedChannel ? { preimage: "dummy preimage" } : undefined
            }
          />
        </>
      )}
    </div>
  );
}
