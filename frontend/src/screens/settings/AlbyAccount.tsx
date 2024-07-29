import { ExitIcon } from "@radix-ui/react-icons";
import { ExternalLinkIcon } from "lucide-react";
import { useNavigate } from "react-router-dom";

import ExternalLink from "src/components/ExternalLink";
import SettingsHeader from "src/components/SettingsHeader";
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { useToast } from "src/components/ui/use-toast";
import { useCSRF } from "src/hooks/useCSRF";
import { request } from "src/utils/request";

export function AlbyAccount() {
  const { data: csrf } = useCSRF();
  const { toast } = useToast();
  const navigate = useNavigate();

  const unlink = async () => {
    if (
      !confirm(
        "Are you sure you want to change the Alby Account for your hub? Your Alby Account will be disconnected from your hub and you'll need to login with a new Alby Account to access your hub."
      )
    ) {
      return;
    }

    try {
      if (!csrf) {
        throw new Error("No CSRF token");
      }
      await request("/api/alby/unlink-account", {
        method: "POST",
        headers: {
          "X-CSRF-Token": csrf,
          "Content-Type": "application/json",
        },
      });
      navigate("/alby/auth?force_login=true");
      toast({
        title: "Alby Account Unlinked",
        description: "Please login with another Alby Account",
      });
    } catch (error) {
      toast({
        title: "Unlink account failed",
        description: (error as Error).message,
        variant: "destructive",
      });
    }
  };

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
      <Card className="w-full cursor-pointer" onClick={unlink}>
        <CardHeader>
          <CardTitle>Change Alby Account</CardTitle>
          <CardDescription className="flex gap-2 items-center">
            <ExitIcon className="w-4 h-4" /> Link your Hub to a different Alby
            Account
          </CardDescription>
        </CardHeader>
      </Card>
    </>
  );
}
