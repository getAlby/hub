import { ExitIcon } from "@radix-ui/react-icons";
import { ExternalLinkIcon } from "lucide-react";

import ExternalLink from "src/components/ExternalLink";
import Loading from "src/components/Loading";
import SettingsHeader from "src/components/SettingsHeader";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "src/components/ui/alert-dialog";
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { LoadingButton } from "src/components/ui/loading-button";
import { UnlinkAlbyAccount } from "src/components/UnlinkAlbyAccount";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useInfo } from "src/hooks/useInfo";
import { useMigrateLDKStorage } from "src/hooks/useMigrateLDKStorage";

export function AlbyAccount() {
  const { data: info } = useInfo();
  const { data: me } = useAlbyMe();
  const { isMigratingStorage, migrateLDKStorage } = useMigrateLDKStorage();
  if (!info || !me) {
    return <Loading />;
  }

  return (
    <>
      <SettingsHeader
        title="Alby Account"
        description="Manage your Alby Account"
      />
      <ExternalLink
        to="https://getalby.com/settings"
        className="w-full flex flex-row items-center gap-2"
      >
        <Card className="w-full">
          <CardHeader>
            <CardTitle>Your Alby Account</CardTitle>
            <CardDescription className="flex gap-2 items-center">
              <ExternalLinkIcon className="w-4 h-4" /> Manage your Alby Account
              Settings such as your lightning address on getalby.com
            </CardDescription>
          </CardHeader>
        </Card>
      </ExternalLink>
      <UnlinkAlbyAccount
        navigateTo="/alby/auth?force_login=true"
        successMessage="Please login with another Alby Account"
      >
        <Card className="w-full cursor-pointer">
          <CardHeader>
            <CardTitle>Change Alby Account</CardTitle>
            <CardDescription className="flex gap-2 items-center">
              <ExitIcon className="w-4 h-4" /> Link your Hub to a different Alby
              Account
            </CardDescription>
          </CardHeader>
        </Card>
      </UnlinkAlbyAccount>

      <UnlinkAlbyAccount>
        <Card className="w-full cursor-pointer">
          <CardHeader>
            <CardTitle>Disconnect Alby Account</CardTitle>
            <CardDescription className="flex gap-2 items-center">
              <ExitIcon className="w-4 h-4" /> Use Alby Hub without an Alby
              Account
            </CardDescription>
          </CardHeader>
        </Card>
      </UnlinkAlbyAccount>
      <div /* spacing */ />
      <SettingsHeader title="VSS" description="Versioned Storage Service" />
      <p>
        Versioned Storage Service (VSS) provides a secure, encrypted server-side
        storage of essential lightning and onchain data.
      </p>
      <p>
        This service is enabled by your Alby account and provides additional
        backup security which allows you to recover your lightning data with
        your recovery phrase alone, without having to close your channels.
      </p>
      {info && (
        <>
          {info.ldkVssEnabled && (
            <p>
              ✅ VSS <b>enabled</b>.{" "}
              {info.ldkVssEnabled && (
                <>Migration from VSS will be available shortly.</>
              )}
            </p>
          )}
          {!me.subscription.buzz && (
            <p>VSS is only available to Alby users with a paid subscription.</p>
          )}
          {!info.ldkVssEnabled && me.subscription.buzz && (
            <AlertDialog>
              <AlertDialogTrigger asChild>
                {
                  <LoadingButton loading={isMigratingStorage}>
                    Enable VSS
                  </LoadingButton>
                }
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>Alby Hub Restart Required</AlertDialogTitle>
                  <AlertDialogDescription>
                    <div>
                      <p>
                        As part of enabling VSS your hub will be shut down, and
                        you will need to enter your unlock password to start it
                        again.
                      </p>
                      <p>
                        Please ensure you have no pending payments or channel
                        closures before continuing.
                      </p>
                    </div>
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>Cancel</AlertDialogCancel>
                  <AlertDialogAction onClick={() => migrateLDKStorage("VSS")}>
                    Confirm
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          )}
        </>
      )}
    </>
  );
}
