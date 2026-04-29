import { CheckIcon, SparklesIcon, StarsIcon } from "lucide-react";
import { ReactNode, useState } from "react";
import { Badge } from "src/components/ui/badge";
import { ExternalLinkButton } from "src/components/ui/custom/external-link-button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "src/components/ui/dialog";
import { DropdownMenuItem } from "src/components/ui/dropdown-menu";
import { Separator } from "src/components/ui/separator";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useInfo } from "src/hooks/useInfo";

interface UpgradeDialogProps {
  children?: ReactNode;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
}

interface ProDropdownMenuItemProps {
  children: ReactNode;
  onClick?: () => void;
}

export function ProDropdownMenuItem({
  children,
  onClick,
}: ProDropdownMenuItemProps) {
  const { data: albyMe } = useAlbyMe();

  if (!albyMe?.subscription.plan_code) {
    return (
      <UpgradeDialog>
        <div className="cursor-pointer">
          <DropdownMenuItem className="w-full pointer-events-none">
            {children}
            <Badge variant="outline">
              <StarsIcon />
              Pro
            </Badge>
          </DropdownMenuItem>
        </div>
      </UpgradeDialog>
    );
  }

  return (
    <DropdownMenuItem className="cursor-pointer" onClick={onClick}>
      {children}
      <Badge variant="outline">
        <StarsIcon />
        Pro
      </Badge>
    </DropdownMenuItem>
  );
}

export const UpgradeDialog = ({
  children,
  open: controlledOpen,
  onOpenChange: controlledOnOpenChange,
}: UpgradeDialogProps) => {
  const [internalOpen, setInternalOpen] = useState(false);
  const open = controlledOpen ?? internalOpen;
  const setOpen = controlledOnOpenChange ?? setInternalOpen;

  const { data: albyMe } = useAlbyMe();
  const { data: info } = useInfo();

  if (!info || (info.albyAccountConnected && !albyMe)) {
    // still loading - make sure button does not incorrectly show
    return;
  }

  if (albyMe?.subscription.plan_code) {
    return null;
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      {children && (
        <span onClick={() => setOpen(true)} className="cursor-pointer">
          {children}
        </span>
      )}
      <DialogContent className="md:max-w-md">
        <DialogHeader className="sr-only">
          <DialogTitle>Upgrade to Pro</DialogTitle>
          <DialogDescription>
            Upgrade to Alby Hub Pro to unlock advanced features and help fund
            Alby's open-source work.
          </DialogDescription>
        </DialogHeader>

        <div className="flex flex-col">
          <h2 className="flex flex-wrap items-center gap-2 text-xl font-semibold tracking-tight">
            Do more with your Hub
            <Badge
              variant="outline"
              className="border-primary/40 bg-primary/10 text-foreground"
            >
              <SparklesIcon />
              Pro
            </Badge>
          </h2>
          <p className="mt-1.5 text-sm text-muted-foreground">
            Unlock advanced features and help fund Alby's open-source work.
          </p>

          <div className="mt-5 flex items-baseline gap-1.5">
            <span className="text-4xl font-semibold tracking-tight tabular-nums">
              $3
            </span>
            <span className="text-sm text-muted-foreground">/ month</span>
          </div>
          <p className="mt-1 text-xs text-muted-foreground">
            Billed $36 yearly · bitcoin or credit card · cancel anytime
          </p>
        </div>

        <Separator />

        <ul className="grid gap-2.5 text-sm">
          {[
            "Unlimited sub-wallets",
            info.backendType === "LDK" && "Encrypted remote backups",
            "Custom lightning address",
            "Export transactions",
            "Email notifications",
            "Priority support",
          ]
            .filter((b): b is string => Boolean(b))
            .map((benefit) => (
              <li key={benefit} className="flex items-center gap-2.5">
                <span className="flex size-4.5 shrink-0 items-center justify-center rounded-full bg-primary/15">
                  <CheckIcon
                    className="size-3 text-foreground/70"
                    strokeWidth={2.5}
                  />
                </span>
                <span>{benefit}</span>
              </li>
            ))}
        </ul>

        <div className="mt-2 space-y-2.5">
          <ExternalLinkButton
            size="lg"
            className="w-full"
            to="https://www.getalby.com/subscription/pro"
          >
            Upgrade to Pro
          </ExternalLinkButton>
          <p className="text-center text-xs text-muted-foreground">
            30-day money-back guarantee
          </p>
        </div>
      </DialogContent>
    </Dialog>
  );
};
