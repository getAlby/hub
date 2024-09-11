import { ExitIcon } from "@radix-ui/react-icons";
import { ExternalLinkIcon } from "lucide-react";
import { useNavigate } from "react-router-dom";

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
import { useToast } from "src/components/ui/use-toast";

import { request } from "src/utils/request";

export function AlbyAccount() {
  const { toast } = useToast();
  const navigate = useNavigate();

  const disconnect = async () => {
    try {
      await request("/api/alby/unlink-account", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
      });
      navigate("/");
      toast({
        title: "Alby Account Disconnected",
        description: "Your hub is no longer connected to an Alby Account.",
      });
    } catch (error) {
      toast({
        title: "Disconnect account failed",
        description: (error as Error).message,
        variant: "destructive",
      });
    }
  };
  const unlink = async () => {
    try {
      await request("/api/alby/unlink-account", {
        method: "POST",
        headers: {
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
            <AlertDialogAction onClick={unlink}>Confirm</AlertDialogAction>
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
                <p className="text-primary font-medium mt-4">
                  Your Alby Account will be disconnected and all Alby Account
                  features such as your lightning address will stop working.
                </p>
              </div>
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={disconnect}>Confirm</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
