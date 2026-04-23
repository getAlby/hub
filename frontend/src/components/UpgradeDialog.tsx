import { StarsIcon } from "lucide-react";
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
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle className="sr-only">Unlock Pro</DialogTitle>
          <DialogDescription className="sr-only">
            Upgrade to Alby Hub Pro to remove limits and unlock additional
            features.
          </DialogDescription>
        </DialogHeader>

        <div className="flex flex-col px-1 pt-2">
          <Badge variant="outline" className="w-fit">
            <StarsIcon />
            Pro
          </Badge>
          <h2 className="mt-3 text-2xl font-semibold tracking-tight">
            More wallets. Better backups. Less hassle.
          </h2>
          <div className="mt-4 flex flex-col">
            <span className="text-5xl font-semibold leading-none tracking-tighter tabular-nums">
              $3
            </span>
            <span className="mt-1 text-xs text-muted-foreground">
              per month, billed yearly
            </span>
          </div>
        </div>

        <ul className="grid grid-cols-2 gap-x-4 gap-y-2.5 text-sm">
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
              <li
                key={benefit}
                className="text-muted-foreground before:mr-2 before:text-foreground before:content-['·']"
              >
                {benefit}
              </li>
            ))}
        </ul>

        <div className="mt-5 space-y-2">
          <ExternalLinkButton
            size="lg"
            className="w-full"
            to="https://www.getalby.com/subscription/pro"
          >
            Upgrade to Pro · $3/mo
          </ExternalLinkButton>
          <p className="text-center text-xs text-muted-foreground">
            30 day money back guarantee · Cancel anytime
          </p>
        </div>
      </DialogContent>
    </Dialog>
  );
};
