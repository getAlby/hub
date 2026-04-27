import {
  ArrowLeftRightIcon,
  Link2OffIcon,
  MailIcon,
  SquareArrowOutUpRightIcon,
  ZapIcon,
} from "lucide-react";

import Loading from "src/components/Loading";
import SettingsHeader from "src/components/SettingsHeader";
import { Button } from "src/components/ui/button";
import { Card, CardContent } from "src/components/ui/card";
import { ExternalLinkButton } from "src/components/ui/custom/external-link-button";
import { UnlinkAlbyAccount } from "src/components/UnlinkAlbyAccount";
import UserAvatar from "src/components/UserAvatar";
import { useAlbyMe } from "src/hooks/useAlbyMe";

export function AlbyAccount() {
  const { data: albyMe, error: albyMeError, isLoading } = useAlbyMe();

  return (
    <>
      <SettingsHeader
        pageTitle="Alby Account"
        title="Alby Account"
        description="Manage the Alby Account linked to your Hub."
      />
      <div className="flex flex-col gap-6">
        {isLoading ? (
          <Card>
            <CardContent className="flex items-center justify-center">
              <Loading />
            </CardContent>
          </Card>
        ) : albyMeError ? (
          <Card>
            <CardContent className="flex items-center justify-center">
              <p className="text-sm text-muted-foreground">
                Failed to load Alby Account details
              </p>
            </CardContent>
          </Card>
        ) : albyMe ? (
          <Card>
            <CardContent className="flex items-center justify-between gap-4">
              <div className="flex items-center gap-4 min-w-0">
                <UserAvatar className="h-14 w-14" />
                <div className="flex flex-col min-w-0">
                  <span className="font-semibold truncate">
                    {albyMe.name || albyMe.email}
                  </span>
                  {albyMe.name && albyMe.email && (
                    <div className="flex items-center gap-1 text-sm text-muted-foreground min-w-0">
                      <MailIcon className="size-4 shrink-0" />
                      <span className="truncate">{albyMe.email}</span>
                    </div>
                  )}
                  {albyMe.lightning_address && (
                    <div className="flex items-center gap-1 text-sm text-muted-foreground min-w-0">
                      <ZapIcon className="size-4 shrink-0" />
                      <span className="truncate">
                        {albyMe.lightning_address}
                      </span>
                    </div>
                  )}
                </div>
              </div>
              <ExternalLinkButton
                variant="outline"
                size="sm"
                to="https://getalby.com/settings"
                className="gap-2 shrink-0"
              >
                Manage on getalby.com <SquareArrowOutUpRightIcon />
              </ExternalLinkButton>
            </CardContent>
          </Card>
        ) : null}

        <div className="flex flex-col gap-3">
          <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wide">
            Danger Zone
          </h3>
          <Card className="border-destructive/40">
            <CardContent className="flex items-center justify-between gap-4">
              <div className="flex flex-col gap-1 text-sm min-w-0">
                <h3 className="font-semibold">Switch Alby Account</h3>
                <p className="text-muted-foreground">
                  Disconnects this account, then prompts you to sign in with
                  another.
                </p>
              </div>
              <UnlinkAlbyAccount
                navigateTo="/alby/auth?force_login=true"
                successMessage="Please login with another Alby Account"
              >
                <Button variant="outline" className="gap-2 shrink-0">
                  <ArrowLeftRightIcon /> Switch Account
                </Button>
              </UnlinkAlbyAccount>
            </CardContent>
            <div className="border-t" />
            <CardContent className="flex items-center justify-between gap-4">
              <div className="flex flex-col gap-1 text-sm min-w-0">
                <h3 className="font-semibold">Disconnect Alby Account</h3>
                <p className="text-muted-foreground">
                  Stops lightning address, notifications, and subscription
                  payments.
                </p>
              </div>
              <UnlinkAlbyAccount>
                <Button variant="destructive" className="gap-2 shrink-0">
                  <Link2OffIcon />
                  Disconnect
                </Button>
              </UnlinkAlbyAccount>
            </CardContent>
          </Card>
        </div>
      </div>
    </>
  );
}
