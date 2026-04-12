import {
  Link2OffIcon,
  RefreshCcwIcon,
  SquareArrowOutUpRightIcon,
  ZapIcon,
} from "lucide-react";

import Loading from "src/components/Loading";
import { ProBadge } from "src/components/ProBadge";
import SettingsHeader from "src/components/SettingsHeader";
import UserAvatar from "src/components/UserAvatar";
import { Button } from "src/components/ui/button";
import { Card, CardContent } from "src/components/ui/card";
import { Separator } from "src/components/ui/separator";
import { UnlinkAlbyAccount } from "src/components/UnlinkAlbyAccount";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useInfo } from "src/hooks/useInfo";

export function AlbyAccount() {
  const { data: info } = useInfo();
  const { data: albyMe } = useAlbyMe();

  if (!info) {
    return <Loading />;
  }

  const hasPlan = !!albyMe?.subscription.plan_code;

  return (
    <>
      <SettingsHeader
        pageTitle="Alby Account"
        title="Alby Account"
        description=""
      />
      <div className="flex flex-col gap-6">
        <Card>
          {albyMe && (
            <>
              <CardContent className="flex items-center gap-3">
                <UserAvatar className="size-10" />
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 font-medium">
                    <span className="truncate">
                      {albyMe.name || albyMe.email}
                    </span>
                    {hasPlan && <ProBadge />}
                  </div>
                  {albyMe.name && (
                    <div className="truncate text-sm text-muted-foreground">
                      {albyMe.email}
                    </div>
                  )}
                  {albyMe.lightning_address && (
                    <div className="flex items-center gap-1 text-sm text-muted-foreground">
                      <ZapIcon className="size-3" />
                      {albyMe.lightning_address}
                    </div>
                  )}
                </div>
                <a
                  href="https://getalby.com/settings"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-sm text-muted-foreground hover:text-foreground transition-colors inline-flex items-center gap-1 shrink-0"
                >
                  Manage
                  <SquareArrowOutUpRightIcon className="size-4" />
                </a>
              </CardContent>
              <Separator />
            </>
          )}
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between gap-4">
              <div className="text-sm">
                <p className="font-medium">Change Alby Account</p>
                <p className="text-muted-foreground">
                  Link your Hub to a different Alby Account
                </p>
              </div>
              <UnlinkAlbyAccount
                navigateTo="/alby/auth?force_login=true"
                successMessage="Please login with another Alby Account"
              >
                <Button variant="outline" className="shrink-0 gap-2">
                  <RefreshCcwIcon className="size-4" />
                  Change
                </Button>
              </UnlinkAlbyAccount>
            </div>
            <Separator />
            <div className="flex items-center justify-between gap-4">
              <div className="text-sm">
                <p className="font-medium">Unlink Alby Account</p>
                <p className="text-muted-foreground">
                  Use your Alby Hub without an Alby Account
                </p>
              </div>
              <UnlinkAlbyAccount>
                <Button
                  variant="outline"
                  className="shrink-0 gap-2 text-destructive border-destructive/30 hover:bg-destructive hover:text-destructive-foreground"
                >
                  <Link2OffIcon className="size-4" />
                  Unlink
                </Button>
              </UnlinkAlbyAccount>
            </div>
          </CardContent>
        </Card>
      </div>
    </>
  );
}
