import { ExitIcon } from "@radix-ui/react-icons";
import { ExternalLinkIcon } from "lucide-react";

import ExternalLink from "src/components/ExternalLink";
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
import { useUnlinkAlbyAccount } from "src/hooks/useUnlinkAlbyAccount";

export function AlbyAccount() {
  const unlinkAccount = useUnlinkAlbyAccount();
  const switchAccount = useUnlinkAlbyAccount(
    "/alby/auth?force_login=true",
    "Please login with another Alby Account"
  );

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
              Settings
            </CardDescription>
          </CardHeader>
        </Card>
      </ExternalLink>
      <AlertDialog>
        <AlertDialogTrigger asChild>
          <Card className="w-full cursor-pointer">
            <CardHeader>
              <CardTitle>Change Alby Account</CardTitle>
              <CardDescription className="flex gap-2 items-center">
                <ExitIcon className="w-4 h-4" /> Link your Hub to a different
                Alby Account
              </CardDescription>
            </CardHeader>
          </Card>
        </AlertDialogTrigger>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Unlink Alby Account</AlertDialogTitle>
            <AlertDialogDescription>
              <div>
                <p>
                  Are you sure you want to change the Alby Account for your hub?
                </p>
                <p className="text-primary font-medium mt-4">
                  Your Alby Account will be disconnected from your hub and
                  you'll need to login with a new Alby Account to access your
                  hub.
                </p>
              </div>
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={switchAccount}>
              Confirm
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
      <AlertDialog>
        <AlertDialogTrigger asChild>
          <Card className="w-full cursor-pointer">
            <CardHeader>
              <CardTitle>Disconnect Alby Account</CardTitle>
              <CardDescription className="flex gap-2 items-center">
                <ExitIcon className="w-4 h-4" /> Use Alby Hub without an Alby
                Account
              </CardDescription>
            </CardHeader>
          </Card>
        </AlertDialogTrigger>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Disconnect Alby Account</AlertDialogTitle>
            <AlertDialogDescription>
              <div>
                <p>Are you sure you want to disconnect your Alby Account?</p>
                <p className="text-destructive font-medium mt-4">
                  Your Alby Account will be disconnected and all Alby Account
                  features such as your lightning address will stop working.
                </p>
              </div>
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={unlinkAccount}>
              Confirm
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
