import {
  Link2OffIcon,
  RefreshCcwIcon,
  SquareArrowOutUpRightIcon,
} from "lucide-react";

import Loading from "src/components/Loading";
import SettingsHeader from "src/components/SettingsHeader";
import { Button } from "src/components/ui/button";
import { ExternalLinkButton } from "src/components/ui/custom/external-link-button";
import { Separator } from "src/components/ui/separator";
import { UnlinkAlbyAccount } from "src/components/UnlinkAlbyAccount";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useInfo } from "src/hooks/useInfo";

export function AlbyAccount() {
  const { data: info } = useInfo();
  const { data: me } = useAlbyMe();

  if (!info || !me) {
    return <Loading />;
  }

  return (
    <>
      <SettingsHeader
        title="Alby Account"
        description="Manage your Alby Account."
      />
      <div className="flex flex-col gap-6">
        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-1 text-sm">
            <h3 className="font-semibold">Manage Alby Account</h3>
            <p className="text-muted-foreground">
              Manage your Alby Account settings such as lightning address or
              notifications on getalby.com
            </p>
          </div>
          <div>
            <ExternalLinkButton
              size={"lg"}
              variant={"secondary"}
              to="https://getalby.com/settings"
              className="flex-1 gap-2 items-center justify-center"
            >
              Alby Account Settings{" "}
              <SquareArrowOutUpRightIcon className="size-4 mr-2" />
            </ExternalLinkButton>
          </div>
        </div>
        <Separator />
        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-1 text-sm">
            <h3 className="font-semibold">Change Alby Account</h3>
            <p className="text-muted-foreground">
              Link your Hub to a different Alby Account
            </p>
          </div>
          <div>
            <UnlinkAlbyAccount
              navigateTo="/alby/auth?force_login=true"
              successMessage="Please login with another Alby Account"
            >
              <Button
                size="lg"
                variant="destructive"
                className="flex-1 gap-2 items-center justify-center py-2 px-4"
              >
                <RefreshCcwIcon className="size-4 mr-2" /> Change Alby Account
              </Button>
            </UnlinkAlbyAccount>
          </div>
        </div>
        <Separator />
        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-1 text-sm">
            <h3 className="font-semibold">Unlink Alby Account</h3>
            <p className="text-muted-foreground">
              Use your Alby Hub without an Alby Account.
            </p>
          </div>
          <div>
            <UnlinkAlbyAccount>
              <Button
                size={"lg"}
                variant={"destructive_outline"}
                className="flex-1 gap-2 items-center justify-center py-2 px-4"
              >
                <Link2OffIcon className="size-4 mr-2" /> Unlink Alby Account
              </Button>
            </UnlinkAlbyAccount>
          </div>
        </div>
      </div>
    </>
  );
}
