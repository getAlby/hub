import {
  LifeBuoyIcon,
  MailIcon,
  RefreshCwIcon,
  SparklesIcon,
  UsersIcon,
  ZapIcon,
} from "lucide-react";
import { ReactNode } from "react";
import { ExternalLinkButton } from "src/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "src/components/ui/dialog";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useInfo } from "src/hooks/useInfo";

interface UpgradeDialogProps {
  children: ReactNode;
}

export const UpgradeDialog = ({ children }: UpgradeDialogProps) => {
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
    <Dialog>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle className="flex flex-row gap-2 items-center">
            <SparklesIcon className="w-6 h-6" />
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
                <UsersIcon className="w-5 h-5 mr-3" />
                <div>
                  <span className="font-medium">Unlimited Sub-wallets</span>
                  <p className="text-sm text-muted-foreground">
                    Share with friends, family, coworkers
                  </p>
                </div>
              </li>
              {info?.backendType === "LDK" && (
                <li className="flex items-center">
                  <RefreshCwIcon className="w-5 h-5 mr-3" />
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
                <LifeBuoyIcon className="w-5 h-5 mr-3" />
                <div>
                  <span className="font-medium">Priority Support</span>
                  <p className="text-sm text-muted-foreground">
                    Get help faster when you need it
                  </p>
                </div>
              </li>
              <li className="flex items-center">
                <ZapIcon className="w-5 h-5 mr-3" />
                <div>
                  <span className="font-medium">Custom Lightning Address</span>
                  <p className="text-sm text-muted-foreground">
                    Create your personalized address
                  </p>
                </div>
              </li>
              <li className="flex items-center">
                <MailIcon className="w-5 h-5 mr-3" />
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
