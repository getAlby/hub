import { InfoIcon } from "lucide-react";
import ExternalLink from "src/components/ExternalLink";
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "src/components/ui/alert-dialog";

type LSPTermsDialogProps = {
  name: string;
  description: string;
  contactUrl: string;
  terms: string | undefined;
  trigger: React.ReactNode;
};
export function LSPTermsDialog({
  name,
  description,
  contactUrl,
  terms,
  trigger,
}: LSPTermsDialogProps) {
  return (
    <AlertDialog>
      <AlertDialogTrigger asChild>
        <div className="cursor-pointer">{trigger}</div>
      </AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Channel Terms - {name}</AlertDialogTitle>
          <AlertDialogDescription>
            <div className="grid gap-4">
              <p>{description}</p>
              <p>
                Learn more about{" "}
                <ExternalLink to={contactUrl} className="underline">
                  {name}
                </ExternalLink>
              </p>

              <div className="flex items-center gap-2">
                Duration: at least 3 months
                <ExternalLink to="https://guides.getalby.com/user-guide/alby-hub/faq/how-to-open-a-payment-channel">
                  <InfoIcon className="size-4 text-muted-foreground" />
                </ExternalLink>
              </div>
              {terms && <p>{terms}</p>}

              <p>
                The duration for which a Lightning Channel remains open is not
                determined or guaranteed by Alby; we will make reasonable
                efforts to share information provided by the relevant LSP, but
                actual availability depends on the Lightning Network and the
                LSP's operations. Channels may be closed at any time, including
                by force closure initiated by the network or counterparties.
              </p>

              <p>The purchase of a payment channel is non-refundable.</p>

              <p>
                To learn more about opening channels, see{" "}
                <ExternalLink
                  className="underline"
                  to="https://guides.getalby.com/user-guide/alby-hub/faq/how-to-open-a-payment-channel"
                >
                  How to open a payment channel?
                </ExternalLink>
              </p>
            </div>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Close</AlertDialogCancel>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
