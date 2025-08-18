import { CopyIcon, PencilIcon, ReceiptTextIcon } from "lucide-react";
import React from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import Loading from "src/components/Loading";
import QRCode from "src/components/QRCode";
import { Button } from "src/components/ui/button";
import { Card, CardContent } from "src/components/ui/card";
import { LinkButton } from "src/components/ui/custom/link-button";
import UserAvatar from "src/components/UserAvatar";
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

      navigate("/wallet/receive/invoice");
    }
  }, [info, meError, navigate]);

  if (!info || (info.albyAccountConnected && !me)) {
    return <Loading />;
  }

  return (
    <div className="w-full md:max-w-lg">
      <div className="grid gap-5">
        {info?.albyAccountConnected && me?.lightning_address && (
          <Card className="w-full md:max-w-xs">
            <CardContent className="flex flex-col items-center gap-6 pt-6">
              <div className="relative flex items-center justify-center">
                <QRCode
                  value={me.lightning_address}
                  className="w-full h-auto"
                />
                <UserAvatar className="w-14 h-14 absolute border-4 border-white bg-white" />
              </div>
              <p className="text-center font-semibold break-all">
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

              <div className="flex gap-4 w-full border-t pt-6">
                <LinkButton
                  to="invoice"
                  variant="outline"
                  className="flex-1 flex gap-2 items-center justify-center"
                >
                  <PencilIcon className="size-4" /> Amount
                </LinkButton>
                {info.backendType === "LDK" && (
                  <LinkButton
                    to="/wallet/receive/offer"
                    variant="outline"
                    className="flex-1"
                  >
                    <ReceiptTextIcon />
                    BOLT-12
                  </LinkButton>
                )}
              </div>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
}
