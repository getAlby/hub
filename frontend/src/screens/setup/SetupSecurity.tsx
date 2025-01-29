import {
  AlertTriangleIcon,
  HandCoins,
  ShieldAlert,
  Unlock,
} from "lucide-react";
import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import ExternalLink from "src/components/ExternalLink";

import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { Button } from "src/components/ui/button";
import { Checkbox } from "src/components/ui/checkbox";
import { Label } from "src/components/ui/label";
import useSetupStore from "src/state/SetupStore";

export function SetupSecurity() {
  const navigate = useNavigate();
  const [hasConfirmed, setConfirmed] = useState<boolean>(false);
  const store = useSetupStore();

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    navigate("/setup/finish");
  }

  return (
    <>
      <div className="grid max-w-sm">
        <form onSubmit={onSubmit} className="flex flex-col items-center w-full">
          <TwoColumnLayoutHeader
            title="Security & Recovery"
            description="Take your time to understand how to secure and recover your funds on Alby Hub."
          />
          <div className="flex flex-col items-center w-full gap-5">
            {(store.nodeInfo.backendType === "LND" ||
              store.nodeInfo.backendType === "PHOENIX") && (
              <Alert>
                <AlertTriangleIcon className="h-4 w-4" />
                <AlertTitle>
                  {store.nodeInfo.backendType} backups are not handled by Alby
                  Hub
                </AlertTitle>
                <AlertDescription>
                  Please take care of your own backups. If you use Umbrel or
                  Start9, please refer to their documentation regarding backups.
                </AlertDescription>
              </Alert>
            )}
          </div>

          <div className="flex flex-col gap-6 w-full mt-10">
            <div className="flex gap-3 items-center">
              <div className="shrink-0">
                <HandCoins className="w-6 h-6" />
              </div>
              <span className="text-sm text-muted-foreground">
                Alby Hub is a spending wallet - do not keep all your savings on
                it!
              </span>
            </div>
            <div className="flex gap-3 items-center">
              <div className="shrink-0">
                <Unlock className="w-6 h-6" />
              </div>
              <span className="text-sm text-muted-foreground">
                Access to your Alby Hub is protected by an unlock password you
                set.
              </span>
            </div>
            <div className="flex gap-3 items-center">
              <div className="shrink-0">
                <ShieldAlert className="w-6 h-6" />
              </div>
              <span className="text-sm text-muted-foreground">
                Your wallet can be recovered only with your 12-word recovery
                phrase.
              </span>
            </div>
            <ExternalLink
              className="text-muted-foreground flex items-center text-sm"
              to="https://guides.getalby.com/user-guide/v/alby-account-and-browser-extension/alby-hub/backups"
            >
              <p>
                Learn more about backups and recovery process on{" "}
                <span className="font-semibold underline">Alby Guides</span>.
              </p>
            </ExternalLink>
            <div className="flex items-center">
              <Checkbox
                id="securePassword"
                required
                onCheckedChange={() => setConfirmed(!hasConfirmed)}
              />
              <Label
                htmlFor="securePassword"
                className="ml-2 text-foreground leading-4"
              >
                I understand how to secure and recover funds
              </Label>
            </div>
            <Button className="w-full" disabled={!hasConfirmed} type="submit">
              Continue
            </Button>
          </div>
        </form>
      </div>
    </>
  );
}
