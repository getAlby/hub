import {
  LifeBuoyIcon,
  MailIcon,
  RefreshCwIcon,
  SparklesIcon,
  StarsIcon,
  UsersIcon,
  ZapIcon,
} from "lucide-react";
import { ReactNode, useState } from "react";
import { Badge } from "src/components/ui/badge";
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
import { ExternalLinkButton } from "./ui/custom/external-link-button";

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
          <DialogTitle className="flex flex-row gap-2 items-center">
            <SparklesIcon className="size-6" />
            Unlock Pro
          </DialogTitle>
          <DialogDescription>
            Take your Alby Hub experience to the next level
          </DialogDescription>
        </DialogHeader>
        <div className="flex flex-col gap-6 my-4">
          <div className="space-y-5">
            <h3 className="font-medium text-lg">Pro Features Include:</h3>
            <ul className="space-y-3">
              <li className="flex items-center">
                <UsersIcon className="size-5 mr-3" />
                <div>
                  <span className="font-medium">Unlimited Sub-wallets</span>
                  <p className="text-sm text-muted-foreground">
                    Share with friends, family, coworkers
                  </p>
                </div>
              </li>
              {info?.backendType === "LDK" && (
                <li className="flex items-center">
                  <RefreshCwIcon className="size-5 mr-3" />
                  <div>
                    <span className="font-medium">
                      Encrypted Remote Backups
                    </span>
                    <p className="text-sm text-muted-foreground">
                      Secure wallet backups in the cloud
                    </p>
                  </div>
                </li>
              )}
              <li className="flex items-center">
                <LifeBuoyIcon className="size-5 mr-3" />
                <div>
                  <span className="font-medium">Priority Support</span>
                  <p className="text-sm text-muted-foreground">
                    Get help faster when you need it
                  </p>
                </div>
              </li>
              <li className="flex items-center">
                <ZapIcon className="size-5 mr-3" />
                <div>
                  <span className="font-medium">Custom Lightning Address</span>
                  <p className="text-sm text-muted-foreground">
                    Create your personalized address
                  </p>
                </div>
              </li>
              <li className="flex items-center">
                <MailIcon className="size-5 mr-3" />
                <div>
                  <span className="font-medium">Email Notifications</span>
                  <p className="text-sm text-muted-foreground">
                    Stay updated on important activity
                  </p>
                </div>
              </li>
            </ul>
          </div>
        </div>
        <div className="flex-col gap-1">
          <ExternalLinkButton
            variant="premium"
            size="lg"
            className="w-full"
            to="https://www.getalby.com/subscription/pro"
          >
            Upgrade Now
          </ExternalLinkButton>
          <p className="text-xs text-center text-muted-foreground mt-2">
            30 day money back guarantee • Cancel anytime
          </p>
        </div>
      </DialogContent>
    </Dialog>
  );
};
