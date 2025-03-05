import { LifeBuoy, RefreshCw, Users } from "lucide-react";
import { ReactNode } from "react";
import { Button } from "src/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "src/components/ui/dialog";
import { ALBY_PRO_PLAN } from "src/constants";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { openLink } from "src/utils/openLink";

interface UpgradeDialogProps {
  children: ReactNode;
}

export const UpgradeDialog = ({ children }: UpgradeDialogProps) => {
  const { data: albyMe } = useAlbyMe();

  const handleUpgrade = () => {
    openLink(`https://www.getalby.com/subscription/new?plan=${ALBY_PRO_PLAN}`);
  };

  if (albyMe?.subscription.buzz) {
    return null;
  }

  return (
    <Dialog>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>✨ Unlock Alby Pro</DialogTitle>
          <DialogDescription>
            Take your Alby Hub experience to the next level
          </DialogDescription>
        </DialogHeader>
        <div className="flex flex-col gap-6 my-4">
          <div className="space-y-5">
            <h3 className="font-medium text-lg">Pro Features Include:</h3>
            <ul className="space-y-3">
              <li className="flex items-center">
                <RefreshCw className="w-5 h-5 mr-3 text-primary" />
                <div>
                  <span className="font-medium">Encrypted Remote Backups</span>
                  <p className="text-sm text-muted-foreground">
                    Secure LDK wallet backups in the cloud
                  </p>
                </div>
              </li>
              <li className="flex items-center">
                <Users className="w-5 h-5 mr-3 text-primary" />
                <div>
                  <span className="font-medium">Unlimited Sub-wallets</span>
                  <p className="text-sm text-muted-foreground">
                    Share with friends, family, coworkers
                  </p>
                </div>
              </li>
              <li className="flex items-center">
                <LifeBuoy className="w-5 h-5 mr-3 text-primary" />
                <div>
                  <span className="font-medium">Priority Support</span>
                  <p className="text-sm text-muted-foreground">
                    Get help faster when you need it
                  </p>
                </div>
              </li>
              <li className="flex items-center">
                <div>
                  <span className="">...and many more</span>
                </div>
              </li>
              {/*
              <li className="flex items-center">
                <Zap className="w-5 h-5 mr-3 text-primary" />
                <div>
                  <span className="font-medium">Custom Lightning Address</span>
                  <p className="text-sm text-muted-foreground">
                    Create your personalized address
                  </p>
                </div>
              </li>
              <li className="flex items-center">
                <Mail className="w-5 h-5 mr-3 text-primary" />
                <div>
                  <span className="font-medium">Email Notifications</span>
                  <p className="text-sm text-muted-foreground">
                    Stay updated on important activity
                  </p>
                </div>
              </li>
              <li className="flex items-center">
                <Server className="w-5 h-5 mr-3 text-primary" />
                <div>
                  <span className="font-medium">Node Monitoring</span>
                  <p className="text-sm text-muted-foreground">
                    Keep your node running smoothly
                  </p>
                </div>
              </li>
              */}
            </ul>
          </div>
        </div>
        <div className="flex-col gap-1">
          <Button
            variant="premium"
            size="lg"
            className="w-full"
            onClick={handleUpgrade}
          >
            Upgrade Now
          </Button>
          <p className="text-xs text-center text-muted-foreground mt-2">
            30 day money back guarantee • Cancel anytime
          </p>
        </div>
      </DialogContent>
    </Dialog>
  );
};
