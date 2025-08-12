import type { Invoice } from "@getalby/lightning-tools/bolt11";
import {
  ArrowLeftIcon,
  CopyIcon,
  ExternalLinkIcon,
  HandCoinsIcon,
} from "lucide-react";
import { useEffect } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import ExternalLink from "src/components/ExternalLink";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import { Button, LinkButton } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { useToast } from "src/components/ui/use-toast";
import { copyToClipboard } from "src/lib/clipboard";

import TickSVG from "public/images/illustrations/tick.svg";
import AppHeader from "src/components/AppHeader";

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

  const to = state?.to as string;
  const pageTitle = state?.pageTitle as string;
  const invoice = state?.invoice as Invoice;

  const copy = () => {
    copyToClipboard(state.preimage as string, toast);
  };

  return (
    <div className="grid gap-4">
      <AppHeader title={pageTitle || "Send"} />
      <div className="w-full md:max-w-lg">
        <Card className="w-full">
          <CardHeader>
            <CardTitle className="text-center">Payment Successful</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col items-center gap-6">
            <img src={TickSVG} className="w-48" />
            <div className="flex flex-col gap-1 items-center">
              <p className="text-2xl font-medium slashed-zero">
                {new Intl.NumberFormat().format(invoice.satoshi)} sats
              </p>
              <FormattedFiatAmount
                amount={invoice.satoshi}
                className="text-xl"
              />
            </div>
            {(to || invoice.description || invoice.successAction) && (
              <div className="flex flex-col items-center w-full gap-4">
                {to && (
                  <p className="text-muted-foreground">
                    to <span className="text-foreground">{to}</span>
                  </p>
                )}
                {invoice.description && (
                  <p className="text-muted-foreground">{invoice.description}</p>
                )}
                <SuccessAction action={invoice.successAction} />
              </div>
            )}
          </CardContent>
          <CardFooter className="flex flex-col gap-2 pt-2">
            <Button onClick={copy} variant="outline" className="w-full">
              <CopyIcon className="w-4 h-4 mr-2" />
              Copy Preimage
            </Button>
            <LinkButton to="/wallet/send" variant="outline" className="w-full">
              <HandCoinsIcon className="w-4 h-4 mr-2" />
              Make Another Payment
            </LinkButton>
            <LinkButton to="/wallet" variant="link" className="w-full">
              <ArrowLeftIcon className="w-4 h-4 mr-2" />
              Back to Wallet
            </LinkButton>
          </CardFooter>
        </Card>
      </div>
    </div>
  );
}

function SuccessAction({ action }: { action: Invoice["successAction"] }) {
  if (!action) {
    return null;
  }

  switch (action.tag) {
    case "message":
      return (
        <div className="w-full p-4 border rounded-lg">
          <div className="font-medium">Message</div>
          <p className="text-sm">{action.message}</p>
        </div>
      );
    case "url":
      return (
        <>
          <div className="w-full p-4 border rounded-lg">
            <div className="font-medium">Description</div>
            <p className="text-sm">{action.description}</p>
          </div>
          <div className="w-full p-4 border rounded-lg">
            <div className="font-medium">URL</div>
            <ExternalLink
              to={action.url}
              className="underline flex items-center"
            >
              <p className="text-sm">{action.url}</p>
              <ExternalLinkIcon className="w-4 h-4 ml-2" />
            </ExternalLink>
          </div>
        </>
      );
  }
}
