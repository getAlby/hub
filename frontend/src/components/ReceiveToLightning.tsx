import {
  ChevronRightIcon,
  CopyIcon,
  LinkIcon,
  ReceiptTextIcon,
  ZapIcon,
} from "lucide-react";
import { Link } from "react-router";
import FirstChannelJitAlert from "src/components/FirstChannelJitAlert";
import Loading from "src/components/Loading";
import QRCode from "src/components/QRCode";
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "src/components/ui/accordion";
import { Button } from "src/components/ui/button";
import { Card, CardContent } from "src/components/ui/card";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useInfo } from "src/hooks/useInfo";
import { copyToClipboard } from "src/lib/clipboard";

export function ReceiveToLightning() {
  const { data: info } = useInfo();
  const { data: me } = useAlbyMe();

  if (!info || !me) {
    return <Loading />;
  }

  return (
    <div className="flex flex-col gap-5">
      <Card>
        <CardContent className="flex flex-col items-center gap-6">
          <FirstChannelJitAlert />
          <QRCode value={me.lightning_address} className="w-full h-auto" />
          <div className="flex max-w-full items-center justify-center gap-1">
            <p className="min-w-0 text-center font-medium text-lg break-all">
              {me.lightning_address}
            </p>
            <Button
              variant="ghost"
              size="icon"
              aria-label="Copy Lightning Address"
              className="shrink-0"
              onClick={() => {
                copyToClipboard(me.lightning_address);
              }}
            >
              <CopyIcon className="size-4" />
            </Button>
          </div>
        </CardContent>
      </Card>
      <Card className="py-2">
        <CardContent>
          <Accordion type="single" collapsible className="w-full">
            <AccordionItem value="more-options">
              <AccordionTrigger>Other ways to receive</AccordionTrigger>
              <AccordionContent className="flex flex-col divide-y pb-1">
                <Link
                  to="/wallet/receive/invoice"
                  className="group flex items-center gap-3 py-3"
                >
                  <ZapIcon className="size-5 shrink-0 text-muted-foreground" />
                  <div className="flex-1">
                    <p className="text-sm font-medium">Create Invoice</p>
                    <p className="text-xs text-muted-foreground">
                      Request a specific amount with a one-time invoice
                    </p>
                  </div>
                  <ChevronRightIcon className="size-4 shrink-0 text-muted-foreground transition-transform group-hover:translate-x-0.5" />
                </Link>
                {info.supportsBolt12 && (
                  <Link
                    to="/wallet/receive/offer"
                    className="group flex items-center gap-3 py-3"
                  >
                    <ReceiptTextIcon className="size-5 shrink-0 text-muted-foreground" />
                    <div className="flex-1">
                      <p className="text-sm font-medium">Lightning Offer</p>
                      <p className="text-xs text-muted-foreground">
                        Share a reusable payment code that never expires
                      </p>
                    </div>
                    <ChevronRightIcon className="size-4 shrink-0 text-muted-foreground transition-transform group-hover:translate-x-0.5" />
                  </Link>
                )}
                <Link
                  to="/wallet/receive/onchain"
                  className="group flex items-center gap-3 py-3"
                >
                  <LinkIcon className="size-5 shrink-0 text-muted-foreground" />
                  <div className="flex-1">
                    <p className="text-sm font-medium">
                      On-chain or Other Cryptocurrency
                    </p>
                    <p className="text-xs text-muted-foreground">
                      Swap funds from on-chain bitcoin or other cryptocurrencies
                    </p>
                  </div>
                  <ChevronRightIcon className="size-4 shrink-0 text-muted-foreground transition-transform group-hover:translate-x-0.5" />
                </Link>
              </AccordionContent>
            </AccordionItem>
          </Accordion>
        </CardContent>
      </Card>
    </div>
  );
}
