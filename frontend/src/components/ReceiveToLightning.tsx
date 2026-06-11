import { CopyIcon, LinkIcon, ReceiptTextIcon, ZapIcon } from "lucide-react";
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
import { LinkButton } from "src/components/ui/custom/link-button";
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
      <Card className="py-0">
        <CardContent>
          <Accordion type="single" collapsible className="w-full">
            <AccordionItem value="more-options">
              <AccordionTrigger>View other ways to receive</AccordionTrigger>
              <AccordionContent className="flex flex-col gap-2">
                <LinkButton
                  to="/wallet/receive/invoice"
                  variant="outline"
                  className="w-full"
                >
                  <ZapIcon className="size-4" />
                  Create Invoice
                </LinkButton>
                {info.supportsBolt12 && (
                  <LinkButton
                    to="/wallet/receive/offer"
                    variant="outline"
                    className="w-full"
                  >
                    <ReceiptTextIcon className="size-4" />
                    Lightning Offer
                  </LinkButton>
                )}
                <LinkButton
                  to="/wallet/receive/onchain"
                  variant="outline"
                  className="w-full"
                >
                  <LinkIcon className="size-4" />
                  Receive from On-chain / Other Cryptocurrency
                </LinkButton>
              </AccordionContent>
            </AccordionItem>
          </Accordion>
        </CardContent>
      </Card>
    </div>
  );
}
