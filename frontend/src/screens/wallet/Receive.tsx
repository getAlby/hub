import { Copy, Edit } from "lucide-react";
import React from "react";
import QRCode from "react-qr-code";
import { useNavigate } from "react-router-dom";
import Loading from "src/components/Loading";
import { Button, LinkButton } from "src/components/ui/button";
import { useToast } from "src/components/ui/use-toast";
import UserAvatar from "src/components/UserAvatar";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useInfo } from "src/hooks/useInfo";
import { copyToClipboard } from "src/lib/clipboard";

export default function Receive() {
  const { data: info } = useInfo();
  const { data: me } = useAlbyMe();
  const { toast } = useToast();
  const navigate = useNavigate();

  // TODO: remove this once we have a CTA to connect an Alby Account to use a lightning address
  React.useEffect(() => {
    if (info && !info.albyAccountConnected) {
      navigate("/wallet/receive/invoice");
    }
  }, [info, navigate]);

  if (!info || (info.albyAccountConnected && !me)) {
    return <Loading />;
  }

  return (
    <div className="grid gap-5">
      {info?.albyAccountConnected && me?.lightning_address && (
        <div className="flex flex-col items-center justify-center gap-6 w-full sm:w-64">
          <div className="relative flex flex-col items-center justify-center w-full">
            <QRCode value={me.lightning_address} className="w-full h-auto" />
            <UserAvatar className="w-16 h-auto absolute border-6 border-primary-foreground" />
          </div>
          <p className="font-semibold break-all">{me.lightning_address}</p>
          <div className="flex gap-4 w-full">
            <LinkButton
              to="invoice"
              variant="outline"
              className="flex-1 flex gap-2 items-center justify-center"
            >
              <Edit className="w-4 h-4" /> Amount
            </LinkButton>
            <Button
              variant="secondary"
              onClick={() => {
                copyToClipboard(me.lightning_address, toast);
              }}
              className="flex-1 flex gap-2 items-center justify-center"
            >
              <Copy className="w-4 h-4" /> Copy
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
