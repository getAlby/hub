import { CopyIcon, LinkIcon, ReceiptTextIcon, ZapIcon } from "lucide-react";
import React from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import AppHeader from "src/components/AppHeader";
import Loading from "src/components/Loading";
import QRCode from "src/components/QRCode";
import { Button } from "src/components/ui/button";
import { Card, CardContent } from "src/components/ui/card";
import { LinkButton } from "src/components/ui/custom/link-button";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useInfo } from "src/hooks/useInfo";
import { copyToClipboard } from "src/lib/clipboard";

export default function Receive() {
  const { data: info } = useInfo();
  const { data: me, error: meError } = useAlbyMe();
  const navigate = useNavigate();

  // TODO: remove this once we have a CTA to connect an Alby Account to use a lightning address
  React.useEffect(() => {
    if (info && (!info.albyAccountConnected || meError)) {
      if (meError) {
        toast.error("Failed to load lightning address");
      }

      navigate("/wallet/receive/invoice", { replace: true });
    }
  }, [info, meError, navigate]);

  if (!info || (info.albyAccountConnected && !me)) {
    return <Loading />;
  }

  return (
    <div className="grid gap-5">
      <AppHeader title="Receive" />
      <div className="w-full max-w-lg">
        {info?.albyAccountConnected && me?.lightning_address && (
          <Card>
            <CardContent className="flex flex-col items-center gap-6 pt-6">
              <QRCode value={me.lightning_address} className="w-full h-auto" />
              <p className="text-center font-medium text-lg break-all my-1">
                {me.lightning_address}
              </p>
              <div className="flex gap-4 w-full">
                <Button
                  variant="secondary"
                  onClick={() => {
                    copyToClipboard(me.lightning_address);
                  }}
                  className="flex-1 flex gap-2 items-center justify-center"
                >
                  <CopyIcon className="size-4" /> Copy Lightning Address
                </Button>
              </div>
              <div className="flex flex-col gap-2 w-full border-t pt-6">
                <LinkButton to="invoice" variant="outline" className="flex-1">
                  <ZapIcon className="w-4 h-4 mr-2" />
                  Create Invoice
                </LinkButton>
                {info.backendType === "LDK" && (
                  <LinkButton
                    to="/wallet/receive/offer"
                    variant="outline"
                    className="flex-1"
                  >
                    <ReceiptTextIcon className="h-4 w-4 mr-2" />
                    Lightning Offer
                  </LinkButton>
                )}
                <LinkButton to="onchain" variant="outline" className="flex-1">
                  <LinkIcon className="w-4 h-4 mr-2" />
                  Receive from On-chain
                </LinkButton>
              </div>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
}
