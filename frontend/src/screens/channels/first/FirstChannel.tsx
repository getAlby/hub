import {
  ChevronDownIcon,
  CreditCardIcon,
  InfoIcon,
  WalletIcon,
} from "lucide-react";
import React from "react";
import { Link, useNavigate } from "react-router-dom";
import { toast } from "sonner";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import Loading from "src/components/Loading";
import { Button } from "src/components/ui/button";
import { Checkbox } from "src/components/ui/checkbox";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Label } from "src/components/ui/label";
import { Separator } from "src/components/ui/separator";
import { useChannels } from "src/hooks/useChannels";

import { useInfo } from "src/hooks/useInfo";
import { AutoChannelRequest, AutoChannelResponse } from "src/types";
import { request } from "src/utils/request";

import { Invoice } from "@getalby/lightning-tools";
import { MempoolAlert } from "src/components/MempoolAlert";
import { PayLightningInvoice } from "src/components/PayLightningInvoice";
import { Table, TableBody, TableCell, TableRow } from "src/components/ui/table";

import LightningNetworkDarkSVG from "public/images/illustrations/lightning-network-dark.svg";
import LightningNetworkLightSVG from "public/images/illustrations/lightning-network-light.svg";
import { LSPTermsDialog } from "src/components/channels/LSPTermsDialog";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "src/components/ui/tooltip";
import { useLSPChannelOffer } from "src/hooks/useLSPChannelOffer";
import { cn } from "src/lib/utils";

export function FirstChannel() {
  const { data: info } = useInfo();
  const { data: channels } = useChannels(true);
  const [isLoading, setLoading] = React.useState(false);
  const [showAdvanced, setShowAdvanced] = React.useState(false);
  const [isPublic, setPublic] = React.useState(false);
  const { data: lspChannelOffer } = useLSPChannelOffer();

  const navigate = useNavigate();
  const [invoice, setInvoice] = React.useState<string>();
  const [channelSize, setChannelSize] = React.useState<number>();

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

  if (!info?.albyAccountConnected || !channels || !lspChannelOffer) {
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
      toast.error("Something went wrong. Please try again");
    }
  }

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
            Alby Hub works with selected service providers (LSPs) which provide
            the best network connectivity and liquidity to receive payments. To
            quickly get started you can buy a channel from an LSP by paying the
            lightning invoice below.
          </p>
          <div className="border rounded-lg slashed-zero w-full">
            <Table>
              <TableBody>
                <TableRow>
                  <TableCell className="font-medium p-3">
                    Incoming Liquidity
                  </TableCell>
                  <TableCell className="text-right p-3">
                    <FormattedBitcoinAmount amount={channelSize * 1000} />
                  </TableCell>
                </TableRow>
                {invoice && (
                  <TableRow>
                    <TableCell className="font-medium p-3">
                      Amount to pay
                    </TableCell>
                    <TableCell className="font-semibold text-right p-3">
                      <FormattedBitcoinAmount
                        amount={new Invoice({ pr: invoice }).satoshi * 1000}
                      />
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </div>
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
                You're now going to open your first lightning channel and can
                begin using your Hub in the booming bitcoin economy!
              </p>
              <p className="text-muted-foreground">
                Alby Hub works with selected service providers (LSPs) which
                provide the best network connectivity and liquidity to receive
                payments.
              </p>
              <p>
                A payment is required to purchase a channel from{" "}
                <ExternalLink
                  to={lspChannelOffer.lspContactUrl}
                  className="underline"
                >
                  {lspChannelOffer.lspName}
                </ExternalLink>
                . Once your channel is opened, you'll immediately be able to
                receive and send bitcoin with your Hub.
              </p>
            </>

            <Table>
              <TableBody>
                <TableRow>
                  <TableCell className="font-medium p-3">
                    Channel Cost
                  </TableCell>
                  <TableCell className="p-3 flex flex-col gap-2 items-end justify-center">
                    <p>
                      <span
                        className={cn(
                          lspChannelOffer.currentPaymentMethod === "included" &&
                            "line-through"
                        )}
                      >
                        {new Intl.NumberFormat(undefined, {
                          style: "currency",
                          currency: "USD",
                        }).format(lspChannelOffer.feeTotalUsd / 100)}
                      </span>
                      {lspChannelOffer.currentPaymentMethod === "included" && (
                        <span> $0.00</span>
                      )}
                    </p>
                  </TableCell>
                </TableRow>
                <TableRow>
                  <TableCell className="font-medium p-3 align-top">
                    <div className="flex flex-1 items-center gap-1">
                      Receiving Capacity{" "}
                      <TooltipProvider>
                        <Tooltip>
                          <TooltipTrigger>
                            <div className="flex flex-row items-center">
                              <InfoIcon className="h-4 w-4 shrink-0 text-muted-foreground" />
                            </div>
                          </TooltipTrigger>
                          <TooltipContent className="max-w-sm">
                            You will be able to receive up to this amount in
                            this channel.
                          </TooltipContent>
                        </Tooltip>
                      </TooltipProvider>
                    </div>
                  </TableCell>
                  <TableCell className="p-3 flex flex-col gap-2 items-end justify-center align-top">
                    <span>
                      <FormattedBitcoinAmount
                        amount={lspChannelOffer.lspBalanceSat * 1000}
                      />
                    </span>
                    <FormattedFiatAmount
                      amount={lspChannelOffer.lspBalanceSat}
                      className="text-xs"
                      showApprox
                    />
                  </TableCell>
                </TableRow>
                {lspChannelOffer.currentPaymentMethod !== "prepaid" &&
                  lspChannelOffer.currentPaymentMethod !== "included" && (
                    <TableRow>
                      <TableCell className="font-medium p-3 flex items-center gap-2">
                        Payment method
                      </TableCell>

                      <TableCell className="p-3 text-right">
                        <ExternalLink to="https://getalby.com/payment_details">
                          <div className="capitalize flex items-center justify-end gap-1 font-medium">
                            {lspChannelOffer.currentPaymentMethod === "card" ? (
                              <CreditCardIcon className="size-4" />
                            ) : (
                              <WalletIcon className="size-4" />
                            )}
                            {lspChannelOffer.currentPaymentMethod.replace(
                              "_",
                              " "
                            )}
                          </div>
                        </ExternalLink>
                      </TableCell>
                    </TableRow>
                  )}
                <TableRow>
                  <TableCell className="font-medium p-3 flex items-center gap-2">
                    Terms
                    {/* <ExternalLink to="https://guides.getalby.com/user-guide/alby-hub/faq/how-to-open-a-payment-channel">
                      <InfoIcon className="size-4 text-muted-foreground" />
                    </ExternalLink> */}
                  </TableCell>

                  <TableCell className="p-3 text-right">
                    <LSPTermsDialog
                      contactUrl={lspChannelOffer.lspContactUrl}
                      description={lspChannelOffer.lspDescription}
                      name={lspChannelOffer.lspName}
                      terms={lspChannelOffer.terms}
                      trigger=<span className="font-medium">View</span>
                    />
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
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
              <div className="flex items-center justify-center -mt-5">
                <Button
                  type="button"
                  variant="link"
                  className="text-muted-foreground text-xs px-0"
                  onClick={() => setShowAdvanced((current) => !current)}
                >
                  Advanced Options
                  <ChevronDownIcon className="size-4" />
                </Button>
              </div>
            )}
            {lspChannelOffer.currentPaymentMethod !== "prepaid" &&
              lspChannelOffer.currentPaymentMethod !== "included" && (
                <p className="text-xs text-muted-foreground flex items-center justify-center -mb-4">
                  The payment for the channel will be due immediately.
                </p>
              )}
            <p className="text-center text-xs -mb-2">
              By continuing, you consent the channel opens immediately and that
              you lose the right to revoke once it is open.
            </p>
            {lspChannelOffer.currentPaymentMethod === "included" && (
              <p className="text-xs text-muted-foreground flex items-center justify-center -mb-2">
                This channel comes free with your Alby Pro subscription.
              </p>
            )}
            <LoadingButton loading={isLoading} onClick={openChannel}>
              {lspChannelOffer.currentPaymentMethod === "prepaid" ? (
                <>Continue</>
              ) : lspChannelOffer.currentPaymentMethod === "included" ? (
                <>Confirm</>
              ) : (
                <>Confirm Payment</>
              )}
            </LoadingButton>
          </div>
        </>
      )}
    </>
  );
}
