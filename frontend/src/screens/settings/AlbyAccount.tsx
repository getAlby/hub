import { ExitIcon } from "@radix-ui/react-icons";
import { ExternalLinkIcon } from "lucide-react";

import ExternalLink from "src/components/ExternalLink";
import SettingsHeader from "src/components/SettingsHeader";
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { UnlinkAlbyAccount } from "src/components/UnlinkAlbyAccount";

export function AlbyAccount() {
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
    </>
  );
}
