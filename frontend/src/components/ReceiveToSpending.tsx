import { CopyIcon, LinkIcon, ReceiptTextIcon, ZapIcon } from "lucide-react";
import Loading from "src/components/Loading";
import QRCode from "src/components/QRCode";
import { Button } from "src/components/ui/button";
import { Card, CardContent, CardFooter } from "src/components/ui/card";
import { LinkButton } from "src/components/ui/custom/link-button";
import { Separator } from "src/components/ui/separator";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useInfo } from "src/hooks/useInfo";
import { copyToClipboard } from "src/lib/clipboard";

export function ReceiveToSpending() {
  const { data: info } = useInfo();
  const { data: me } = useAlbyMe();

  if (!info || !me) {
    return <Loading />;
  }

  return (
    <Card>
      <CardContent className="flex flex-col items-center gap-6">
        <QRCode value={me.lightning_address} className="w-full h-auto" />
        <p className="text-center font-medium text-lg break-all">
          {me.lightning_address}
        </p>
      </CardContent>
      <CardFooter className="flex flex-col gap-2 pt-2">
        <Button
          variant="secondary"
          onClick={() => {
            copyToClipboard(me.lightning_address);
          }}
          className="w-full"
        >
          <CopyIcon className="size-4" /> Copy Lightning Address
        </Button>
        <Separator className="my-4" />
        <LinkButton to="invoice" variant="outline" className="w-full">
          <ZapIcon className="w-4 h-4 mr-2" />
          Create Invoice
        </LinkButton>
        {info.backendType === "LDK" && (
          <LinkButton
            to="/wallet/receive/offer"
            variant="outline"
            className="w-full"
          >
            <ReceiptTextIcon className="h-4 w-4 mr-2" />
            Lightning Offer
          </LinkButton>
        )}
        <LinkButton to="onchain" variant="outline" className="w-full">
          <LinkIcon className="w-4 h-4 mr-2" />
          Receive from On-chain / Other Cryptocurrency
        </LinkButton>
      </CardFooter>
    </Card>
  );
}
