import type { Invoice } from "@getalby/lightning-tools/bolt11";
import { CircleCheckIcon, CopyIcon, ExternalLinkIcon } from "lucide-react";
import { useEffect } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import ExternalLink from "src/components/ExternalLink";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { useToast } from "src/components/ui/use-toast";
import { copyToClipboard } from "src/lib/clipboard";

export default function PaymentSuccess() {
  const { state } = useLocation();
  const navigate = useNavigate();
  const { toast } = useToast();

  useEffect(() => {
    if (!state?.preimage) {
      navigate("/wallet/send");
    }
  }, [state, navigate]);

  if (!state?.preimage || !state?.invoice) {
    return null;
  }

  const invoice = state?.invoice as Invoice;

  const copy = () => {
    copyToClipboard(state.preimage as string, toast);
  };

  return (
    <>
      <Card className="w-full">
        <CardHeader>
          <CardTitle className="text-center">Payment Successful</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col items-center gap-4">
          <CircleCheckIcon className="w-32 h-32 mb-2" />
          <div className="flex flex-col gap-2 items-center">
            <p className="text-xl font-bold slashed-zero">
              {new Intl.NumberFormat().format(invoice.satoshi)} sats
            </p>
            <FormattedFiatAmount amount={invoice.satoshi} />
          </div>
          <Button onClick={copy} variant="outline">
            <CopyIcon />
            Copy Preimage
          </Button>
        </CardContent>
      </Card>
      {invoice.successAction && invoice.successAction.tag === "message" && (
        <div className="w-full p-4 mt-4 border rounded-lg">
          <div className="font-medium">Message</div>
          <p className="text-sm">{invoice.successAction.message}</p>
        </div>
      )}
      {invoice.successAction && invoice.successAction.tag === "url" && (
        <>
          <div className="w-full p-4 mt-4 border rounded-lg">
            <div className="font-medium">Description</div>
            <p className="text-sm">{invoice.successAction.description}</p>
          </div>
          <div className="w-full p-4 mt-4 border rounded-lg">
            <div className="font-medium">URL</div>
            <ExternalLink
              to={invoice.successAction.url}
              className="underline flex items-center"
            >
              <p className="text-sm">{invoice.successAction.url}</p>
              <ExternalLinkIcon className="size-4 ml-2" />
            </ExternalLink>
          </div>
        </>
      )}
      <Link to="/wallet/send">
        <Button className="mt-4 w-full">Make Another Payment</Button>
      </Link>
      <Link to="/wallet">
        <Button className="mt-4 w-full" variant="secondary">
          Back To Wallet
        </Button>
      </Link>
    </>
  );
}
